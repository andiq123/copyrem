package config

type Params struct {
	Bitrate    string
	SampleRate int
	Channels   int
	// Lo-fi character at intensity 1.0; the slider scales these up and down.
	CrushBits  float64 // acrusher bit depth — lower is crunchier
	SampleHold int     // acrusher sample-and-hold — higher is grittier
	LowpassHz  int     // muffle cutoff — lower is warmer/duller
	HighpassHz int     // thin-out cutoff — trims the sub
	WowDepth   float64 // tape wow/flutter depth (vibrato)
}

// Defaults are the lo-fi tuning knobs. Edit here to retune the house sound.
func Defaults() Params {
	return Params{
		Bitrate:    "192k",
		SampleRate: 44100,
		Channels:   2,
		CrushBits:  10,
		SampleHold: 2,
		LowpassHz:  5000,
		HighpassHz: 180,
		WowDepth:   0.12,
	}
}
