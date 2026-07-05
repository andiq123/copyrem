package converter

import (
	"strconv"
	"strings"
	"testing"

	"copyrem/internal/config"
)

// buildArgs must keep every scaled lo-fi knob inside the range its ffmpeg
// filter accepts, across the whole intensity slider, or the chain fails at
// runtime. acrusher bits must stay in [4,16] and vibrato depth in [0,1].
func TestBuildArgsLofiRanges(t *testing.T) {
	cfg := config.Defaults()
	for _, intensity := range []float64{0.5, 1.0, 1.75, 2.5} {
		af := afArg(buildArgs(cfg, "in.mp3", "out.mp3", intensity))
		if !strings.Contains(af, "acrusher=") || !strings.Contains(af, "lowpass=") {
			t.Fatalf("intensity %v: chain missing lo-fi stages: %q", intensity, af)
		}
		if b := num(t, af, "bits="); b < 4 || b > 16 {
			t.Errorf("intensity %v: acrusher bits=%v outside [4,16]", intensity, b)
		}
		if d := num(t, af, "d="); d < 0 || d > 1 {
			t.Errorf("intensity %v: vibrato d=%v outside [0,1]", intensity, d)
		}
	}
}

// num pulls the float after key= from a comma/colon-delimited filter string.
func num(t *testing.T, af, key string) float64 {
	_, rest, ok := strings.Cut(af, key)
	if !ok {
		t.Fatalf("key %q not found in %q", key, af)
	}
	end := strings.IndexAny(rest, ":,")
	if end < 0 {
		end = len(rest)
	}
	v, err := strconv.ParseFloat(rest[:end], 64)
	if err != nil {
		t.Fatalf("bad number for %q: %v", key, err)
	}
	return v
}

func afArg(args []string) string {
	for i, a := range args {
		if a == "-af" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
