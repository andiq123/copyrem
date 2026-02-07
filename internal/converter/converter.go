package converter

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"

	"copyrem/internal/config"
	"copyrem/internal/ffmpeg"
)

func ConvertWithProgress(ctx context.Context, cfg config.Params, input, output string, onProgress func(int)) error {
	binary := ffmpeg.FindBinary()

	var totalUs float64
	if onProgress != nil {
		if dur, err := ffmpeg.Duration(binary, input); err == nil && dur > 0 {
			totalUs = float64(dur.Microseconds())
		}
	}

	args := buildArgs(cfg, input, output)
	if onProgress != nil {
		args = append([]string{"-progress", "pipe:1"}, args...)
	}

	cmd := exec.CommandContext(ctx, binary, args...)

	if onProgress != nil && totalUs > 0 {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("stdout pipe: %w", err)
		}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("ffmpeg start: %w", err)
		}
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 256), 256)
		lastPct := 0
		for scanner.Scan() {
			if !strings.HasPrefix(scanner.Text(), "out_time_us=") {
				continue
			}
			us, err := strconv.ParseFloat(strings.TrimPrefix(scanner.Text(), "out_time_us="), 64)
			if err != nil {
				continue
			}
			pct := int(math.Min(99, (us/totalUs)*100))
			if pct > lastPct {
				lastPct = pct
				onProgress(pct)
			}
		}
	} else {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("ffmpeg start: %w", err)
		}
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("ffmpeg: %w", err)
	}
	if onProgress != nil {
		onProgress(100)
	}
	return nil
}

func buildArgs(cfg config.Params, input, output string) []string {
	sr := cfg.SampleRate
	p := math.Pow(2, cfg.PitchSemitones/12)

	pitch := fmt.Sprintf("asetrate=%d*%.6f,aresample=%d,atempo=%.6f", sr, p, sr, 1/p)
	tempo := fmt.Sprintf("atempo=%.4f", cfg.TempoFactor)

	var resample []string
	for _, r := range cfg.ResampleRates {
		resample = append(resample, fmt.Sprintf("aresample=%d", r))
	}
	resample = append(resample, fmt.Sprintf("aresample=%d", sr))

	delay := fmt.Sprintf("adelay=%d|%d", cfg.DelayLeftMs, cfg.DelayRightMs)
	filter := strings.Join(append([]string{pitch, tempo}, append(resample, delay)...), ",")

	return []string{
		"-y", "-i", input,
		"-af", filter,
		"-b:a", cfg.Bitrate,
		"-ar", strconv.Itoa(sr),
		"-ac", strconv.Itoa(cfg.Channels),
		output,
	}
}
