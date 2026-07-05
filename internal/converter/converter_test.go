package converter

import (
	"strconv"
	"strings"
	"testing"

	"copyrem/internal/config"
)

func TestBuildArgsPerturbationRanges(t *testing.T) {
	cfg := config.Defaults()

	subtle := afArg(buildArgs(cfg, "in.mp3", "out.mp3", 0.5))
	if strings.Contains(subtle, "asetrate=") || strings.Contains(subtle, "atempo=") {
		t.Fatalf("intensity 0.5 should skip warp filters, got %q", subtle)
	}

	for _, intensity := range []float64{1.0, 1.75, 2.5} {
		af := afArg(buildArgs(cfg, "in.mp3", "out.mp3", intensity))
		if !strings.Contains(af, "asetrate=") {
			t.Fatalf("intensity %v: missing pitch warp: %q", intensity, af)
		}
		for _, tempo := range allNums(af, "atempo=") {
			if tempo < 0.5 || tempo > 2.0 {
				t.Errorf("intensity %v: atempo=%v outside [0.5,2.0]", intensity, tempo)
			}
		}
	}

	max := afArg(buildArgs(cfg, "in.mp3", "out.mp3", 2.5))
	pitch := num(t, max, "asetrate=44100*")
	if pitch < 1.05 {
		t.Errorf("max pitch factor %v, want >1.05 (~1 semitone up)", pitch)
	}
	pace := (1 / pitch) * 0.72
	if pace > 0.75 {
		t.Errorf("effective tempo %v, want <=0.75 to beat detect warp grid", pace)
	}
}

func TestAtempoChain(t *testing.T) {
	if got := atempoChain(1.0); got != "" {
		t.Fatalf("unity: %q", got)
	}
	if got := atempoChain(0.68); !strings.HasPrefix(got, "atempo=0.680000,") {
		t.Fatalf("single hop: %q", got)
	}
	if got := atempoChain(0.25); strings.Count(got, "atempo=") != 2 {
		t.Fatalf("split hop: %q", got)
	}
}

func allNums(af, key string) []float64 {
	var out []float64
	rest := af
	for {
		i := strings.Index(rest, key)
		if i < 0 {
			break
		}
		rest = rest[i+len(key):]
		end := strings.IndexAny(rest, ":,")
		if end < 0 {
			end = len(rest)
		}
		v, err := strconv.ParseFloat(rest[:end], 64)
		if err != nil {
			break
		}
		out = append(out, v)
	}
	return out
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

func afArg(args []string) string {
	for i, a := range args {
		if a == "-af" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
