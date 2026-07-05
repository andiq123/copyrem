package config

type Params struct {
	Bitrate    string
	SampleRate int
	Channels   int
	// Perturbation knobs at intensity 2.5; the slider scales 0.5→2.5.
	PitchSemitones float64 // chroma shift — primary fingerprint breaker
	TempoFactor    float64 // playback speed at max (<1 slows)
	Drive          float64 // soft-clip amount via boost+limiter
}

func Defaults() Params {
	return Params{
		Bitrate:        "192k",
		SampleRate:     44100,
		Channels:       2,
		PitchSemitones: 1.0,
		TempoFactor:    0.72,
		Drive:          0.5,
	}
}
