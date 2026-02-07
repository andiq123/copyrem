package converter

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"copyrem/internal/config"
	"copyrem/internal/ffmpeg"
)

func Convert(cfg config.Params, input, output string) error {
	filter := buildFilterChain(cfg)
	args := buildArgs(cfg, input, output, filter)
	binary := ffmpeg.FindBinary()
	return ffmpeg.Run(binary, args, nil, nil)
}

func buildFilterChain(cfg config.Params) string {
	sr := cfg.SampleRate
	p := math.Pow(2, cfg.PitchSemitones/12)
	atempoPitch := 1 / p

	pitch := fmt.Sprintf("asetrate=%d*%.6f,aresample=%d,atempo=%.6f", sr, p, sr, atempoPitch)
	tempo := fmt.Sprintf("atempo=%.4f", cfg.TempoFactor)
	resample := buildResampleChain(cfg)
	delay := fmt.Sprintf("adelay=%d|%d", cfg.DelayLeftMs, cfg.DelayRightMs)

	parts := []string{pitch, tempo, resample, delay}
	return strings.Join(parts, ",")
}

func buildResampleChain(cfg config.Params) string {
	var chain []string
	for _, r := range cfg.ResampleRates {
		chain = append(chain, fmt.Sprintf("aresample=%d", r))
	}
	chain = append(chain, fmt.Sprintf("aresample=%d", cfg.SampleRate))
	return strings.Join(chain, ",")
}

func buildArgs(cfg config.Params, input, output, filter string) []string {
	return []string{
		"-y", "-i", input,
		"-af", filter,
		"-b:a", cfg.Bitrate,
		"-ar", strconv.Itoa(cfg.SampleRate),
		"-ac", strconv.Itoa(cfg.Channels),
		output,
	}
}
