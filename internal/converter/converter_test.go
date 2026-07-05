package converter

import (
	"strconv"
	"strings"
	"testing"

	"copyrem/internal/config"
)

// buildArgs must stay inside ffmpeg atempo/asetrate limits across the slider
// and at 2.5 push past the ±0.6 st / 0.75 tempo warp grid in detect.go.
func TestBuildArgsPerturbationRanges(t *testing.T) {
	cfg := config.Defaults()
	for _, intensity := range []float64{0.5, 1.0, 1.75, 2.5} {
		af := afArg(buildArgs(cfg, "in.mp3", "out.mp3", intensity))
		if !strings.Contains(af, "asetrate=") || !strings.Contains(af, "atempo=") {
			t.Fatalf("intensity %v: chain missing warp stages: %q", intensity, af)
		}
		if tempo := lastNum(t, af, "atempo="); tempo < 0.5 || tempo > 2.0 {
			t.Errorf("intensity %v: atempo=%v outside [0.5,2.0]", intensity, tempo)
		}
	}

	max := afArg(buildArgs(cfg, "in.mp3", "out.mp3", 2.5))
	pitch := num(t, max, "asetrate=44100*")
	if pitch < 1.05 {
		t.Errorf("max pitch factor %v, want >1.05 (~1 semitone up)", pitch)
	}
	if tempo := lastNum(t, max, "atempo="); tempo > 0.75 {
		t.Errorf("max tempo %v, want <=0.75 to beat detect warp grid", tempo)
	}
}

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

func lastNum(t *testing.T, af, key string) float64 {
	i := strings.LastIndex(af, key)
	if i < 0 {
		t.Fatalf("key %q not found in %q", key, af)
	}
	rest := af[i+len(key):]
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
