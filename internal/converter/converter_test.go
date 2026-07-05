package converter

import (
	"strconv"
	"strings"
	"testing"

	"copyrem/internal/config"
)

// buildArgs must never emit an atempo factor outside ffmpeg's [0.5, 2.0],
// or the whole filter chain fails at runtime. Check the extremes of the slider.
func TestBuildArgsAtempoInRange(t *testing.T) {
	cfg := config.Defaults()
	for _, intensity := range []float64{0.5, 1.0, 1.75, 2.5} {
		af := afArg(buildArgs(cfg, "in.mp3", "out.mp3", intensity))
		if !strings.Contains(af, "asetrate=") || !strings.Contains(af, "adelay=") {
			t.Fatalf("intensity %v: filter chain missing stages: %q", intensity, af)
		}
		for part := range strings.SplitSeq(af, ",") {
			if !strings.HasPrefix(part, "atempo=") {
				continue
			}
			f, err := strconv.ParseFloat(strings.TrimPrefix(part, "atempo="), 64)
			if err != nil {
				t.Fatalf("intensity %v: bad atempo %q", intensity, part)
			}
			if f < 0.5 || f > 2.0 {
				t.Errorf("intensity %v: atempo=%v outside [0.5,2.0]", intensity, f)
			}
		}
	}
}

func afArg(args []string) string {
	for i, a := range args {
		if a == "-af" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
