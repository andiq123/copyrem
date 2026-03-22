package converter

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"copyrem/internal/config"
	"copyrem/internal/ffmpeg"
)

const progressMinStep = 2
const progressMinInterval = 200 * time.Millisecond

func ConvertWithProgress(ctx context.Context, cfg config.Params, input, output string, intensity float64, onProgress func(int)) error {
	binary := ffmpeg.FindBinary()

	var totalUs float64
	if onProgress != nil {
		if dur, err := ffmpeg.Duration(binary, input); err == nil && dur > 0 {
			totalUs = float64(dur.Microseconds())
		}
		if totalUs == 0 {
			onProgress = nil
		}
	}

	args := buildArgs(cfg, input, output, intensity)
	if onProgress != nil {
		args = append([]string{"-progress", "pipe:1"}, args...)
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	var stdout io.ReadCloser
	var err error
	if onProgress != nil {
		if stdout, err = cmd.StdoutPipe(); err != nil {
			return fmt.Errorf("stdout pipe: %w", err)
		}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start: %w", err)
	}

	if onProgress != nil && stdout != nil {
		trackProgress(stdout, totalUs, onProgress)
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if stderr.Len() > 0 {
			return fmt.Errorf("ffmpeg: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
		}
		return fmt.Errorf("ffmpeg: %w", err)
	}
	if onProgress != nil {
		onProgress(100)
	}
	return nil
}

func trackProgress(stdout io.ReadCloser, totalUs float64, onProgress func(int)) {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 256), 256)
	lastPct := 0
	lastReport := time.Time{}

	for scanner.Scan() {
		text := scanner.Text()
		if !strings.HasPrefix(text, "out_time_us=") {
			continue
		}
		us, err := strconv.ParseFloat(strings.TrimPrefix(text, "out_time_us="), 64)
		if err != nil {
			continue
		}
		pct := int(math.Min(99, (us/totalUs)*100))
		if pct <= lastPct {
			continue
		}
		now := time.Now()
		if pct-lastPct >= progressMinStep || now.Sub(lastReport) >= progressMinInterval {
			lastPct = pct
			lastReport = now
			onProgress(pct)
		}
	}
}

func buildArgs(cfg config.Params, input, output string, intensity float64) []string {
	sr := cfg.SampleRate
	p := math.Pow(2, (cfg.PitchSemitones*intensity)/12)

	pitch := fmt.Sprintf("asetrate=%d*%.6f,aresample=%d,atempo=%.6f", sr, p, sr, 1/p)

	// Scale tempo: if TempoFactor < 1, higher intensity makes it slower
	var tf float64
	if cfg.TempoFactor < 1.0 {
		tf = 1.0 - (1.0-cfg.TempoFactor)*intensity
	} else {
		tf = 1.0 + (cfg.TempoFactor-1.0)*intensity
	}
	// Clamp tempo to ffmpeg limits (0.5 to 2.0 per atempo filter)
	tf = math.Max(0.5, math.Min(2.0, tf))
	tempo := fmt.Sprintf("atempo=%.4f", tf)

	var resample []string
	for _, r := range cfg.ResampleRates {
		resample = append(resample, fmt.Sprintf("aresample=%d", r))
	}
	resample = append(resample, fmt.Sprintf("aresample=%d", sr))

	delayL := int(float64(cfg.DelayLeftMs) * intensity)
	delayR := int(float64(cfg.DelayRightMs) * intensity)
	delay := fmt.Sprintf("adelay=%d|%d", delayL, delayR)

	parts := make([]string, 0, 2+len(resample)+1)
	parts = append(parts, pitch, tempo)
	parts = append(parts, resample...)
	parts = append(parts, delay)
	filter := strings.Join(parts, ",")

	return []string{
		"-y", "-i", input,
		"-af", filter,
		"-b:a", cfg.Bitrate,
		"-ar", strconv.Itoa(sr),
		"-ac", strconv.Itoa(cfg.Channels),
		output,
	}
}
