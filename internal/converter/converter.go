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
		if _, pace := effectiveWarp(cfg, intensity); pace > 0 && pace < 1 {
			totalUs /= pace // output runs longer when tempo slows
		}
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
		_ = stdout.Close()
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
	return nil
}

func trackProgress(stdout io.ReadCloser, totalUs float64, onProgress func(int)) {
	scanner := bufio.NewScanner(stdout)
	lastPct := 0
	lastReport := time.Time{}
	done := false

	for scanner.Scan() {
		text := scanner.Text()
		if text == "progress=end" {
			if !done {
				onProgress(100)
				done = true
			}
			continue
		}
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
	if !done {
		onProgress(100)
	}
}

// atempoChain emits one or more atempo filters; ffmpeg only accepts 0.5–2.0 each.
func atempoChain(factor float64) string {
	if math.Abs(factor-1.0) < 1e-4 {
		return ""
	}
	var filters []string
	f := factor
	for f < 0.5 || f > 2.0 {
		step := 0.5
		if f > 2.0 {
			step = 2.0
		}
		filters = append(filters, fmt.Sprintf("atempo=%.6f", step))
		f /= step
	}
	if math.Abs(f-1.0) >= 1e-4 {
		filters = append(filters, fmt.Sprintf("atempo=%.6f", f))
	}
	if len(filters) == 0 {
		return ""
	}
	return strings.Join(filters, ",") + ","
}

// effectiveWarp returns pitch (semitones) and playback pace at intensity.
func effectiveWarp(cfg config.Params, intensity float64) (semitones, pace float64) {
	t := min(max((intensity-0.5)/2.0, 0), 1)
	semitones = t * cfg.PitchSemitones
	tempo := 1.0 - t*(1.0-cfg.TempoFactor)
	p := math.Pow(2, semitones/12)
	return semitones, (1 / p) * tempo
}

func buildArgs(cfg config.Params, input, output string, intensity float64) []string {
	t := min(max((intensity-0.5)/2.0, 0), 1)

	semitones := t * cfg.PitchSemitones
	tempo := 1.0 - t*(1.0-cfg.TempoFactor)
	drive := 1.0 + t*cfg.Drive
	sr := cfg.SampleRate

	tail := fmt.Sprintf("volume=%.3f,acompressor=threshold=-18dB:ratio=2.5:attack=15:release=180,alimiter=limit=0.97", drive)

	var filter string
	var inputArgs []string
	if t == 0 {
		filter = tail
	} else {
		p := math.Pow(2, semitones/12)
		downsample := ""
		if t >= 0.75 {
			// ponytail: 32k round-trip kills highs the warp grid never models
			downsample = fmt.Sprintf("aresample=32000,aresample=%d,", sr)
		}
		spectral := fmt.Sprintf(
			"highpass=f=%d,lowpass=f=%d,equalizer=f=800:t=q:w=1.2:g=%.1f,equalizer=f=3500:t=q:w=2:g=%.1f,%s",
			int(100+t*150), int(17000-t*5000), -t*5, t*2.5, downsample,
		)
		filter = fmt.Sprintf(
			"aresample=%d:filter_size=32,asetrate=%d*%.6f,aresample=%d:filter_size=32,%s%s%s",
			sr, sr, p, sr, atempoChain((1/p)*tempo), spectral, tail,
		)
		if skip := t * 0.5; skip >= 0.05 {
			inputArgs = []string{"-ss", fmt.Sprintf("%.3f", skip)}
		}
	}

	out := []string{
		"-y", "-nostdin", "-hide_banner", "-loglevel", "error",
		"-threads", "0",
	}
	out = append(out, inputArgs...)
	out = append(out,
		"-i", input,
		"-vn", "-sn", "-dn",
		"-map_metadata", "-1",
		"-af", filter,
		"-c:a", "libmp3lame",
		"-b:a", cfg.Bitrate,
		"-compression_level", "0",
		"-ar", strconv.Itoa(cfg.SampleRate),
		"-ac", strconv.Itoa(cfg.Channels),
		output,
	)
	return out
}
