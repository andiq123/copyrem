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

// buildArgs warps pitch/tempo and adds gentle saturation. Intensity 0.5 is
// nearly transparent; 2.5 lands outside the ±0.6 st / 0.75 tempo warp grid
// used by typical Chromaprint matchers.
func buildArgs(cfg config.Params, input, output string, intensity float64) []string {
	t := min(max((intensity-0.5)/2.0, 0), 1)

	semitones := t * cfg.PitchSemitones
	tempo := 1.0 - t*(1.0-cfg.TempoFactor)
	drive := 1.0 + t*cfg.Drive

	p := math.Pow(2, semitones/12)
	sr := cfg.SampleRate

	filter := fmt.Sprintf(
		"aresample=%d,asetrate=%d*%.6f,aresample=%d,atempo=%.6f,atempo=%.6f,volume=%.3f,acompressor=threshold=-18dB:ratio=2.5:attack=15:release=180,alimiter=limit=0.97",
		sr, sr, p, sr, 1/p, tempo, drive,
	)

	return []string{
		"-y", "-i", input,
		"-af", filter,
		"-b:a", cfg.Bitrate,
		"-ar", strconv.Itoa(cfg.SampleRate),
		"-ac", strconv.Itoa(cfg.Channels),
		output,
	}
}
