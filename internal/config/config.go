package config

type Params struct {
	Bitrate        string
	SampleRate     int
	Channels       int
	TempoFactor    float64
	PitchSemitones float64
	ResampleRates  []int
	DelayLeftMs    int
	DelayRightMs   int
}

func Default() Params {
	return Params{
		Bitrate:         "320k",
		SampleRate:      44100,
		Channels:        2,
		TempoFactor:     0.90,
		PitchSemitones:   0.25,
		ResampleRates:   []int{48000, 96000, 48000},
		DelayLeftMs:     1,
		DelayRightMs:    8,
	}
}
