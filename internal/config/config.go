package config

import (
	"encoding/json"
	"os"
)

type Params struct {
	Bitrate        string `json:"bitrate"`
	SampleRate     int    `json:"sample_rate"`
	Channels       int    `json:"channels"`
	TempoFactor    float64 `json:"tempo_factor"`
	PitchSemitones float64 `json:"pitch_semitones"`
	ResampleRates  []int   `json:"resample_rates"`
	DelayLeftMs    int     `json:"delay_left_ms"`
	DelayRightMs   int     `json:"delay_right_ms"`
}

func Load(path string) (Params, error) {
	p := defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	return p, nil
}

func defaults() Params {
	return Params{
		Bitrate:        "320k",
		SampleRate:     44100,
		Channels:       2,
		TempoFactor:    0.90,
		PitchSemitones: 0.25,
		ResampleRates:  []int{48000, 96000, 48000},
		DelayLeftMs:    1,
		DelayRightMs:   8,
	}
}
