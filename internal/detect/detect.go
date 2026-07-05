// Package detect makes Chromaprint/AcoustID matching robust to the pitch,
// tempo, resample and stereo-delay perturbations that defeat a raw fingerprint.
//
// Strategy: at ingest, fingerprint each reference track across a small warp
// grid (the same family an evader applies) and store all variants. At query
// time, take one fingerprint of the upload and accept it if it matches any
// stored variant above a threshold. fpcalc downmixes to mono internally, so
// stereo-delay tricks are already neutralised.
package detect

import (
	"context"
	"fmt"
	"math"
	"math/bits"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"copyrem/internal/ffmpeg"
)

// warpGrid covers the perturbation range CopyRem-style tools use: small pitch
// shifts and tempo slowdowns. Widen it if your evasion range grows.
var warpGrid = []struct{ semitones, tempo float64 }{
	{0, 1.0}, // the untouched reference
	{+0.6, 1.0}, {-0.6, 1.0},
	{0, 0.85}, {+0.6, 0.85}, {-0.6, 0.85},
	{0, 0.75}, {+0.6, 0.75}, {-0.6, 0.75},
}

// minOverlapFrac is the fraction of the shorter fingerprint that must overlap
// for an alignment offset to count — stops a 1-frame overlap scoring 1.0.
const minOverlapFrac = 0.5

// Similarity returns the best bitwise agreement between two raw Chromaprint
// fingerprints over all alignment offsets. 1.0 = identical, ~0.5 = unrelated
// (random 32-bit sub-fingerprints agree on half their bits).
func Similarity(a, b []uint32) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	minLen := min(len(a), len(b))
	minOverlap := int(float64(minLen) * minOverlapFrac)
	best := 0.0
	for off := -len(b) + 1; off < len(a); off++ {
		matchBits, overlap := 0, 0
		for i := range a {
			j := i - off
			if j < 0 || j >= len(b) {
				continue
			}
			matchBits += 32 - bits.OnesCount32(a[i]^b[j])
			overlap++
		}
		if overlap < minOverlap || overlap == 0 {
			continue
		}
		if r := float64(matchBits) / float64(overlap*32); r > best {
			best = r
		}
	}
	return best
}

// Match reports whether query matches any indexed variant at or above threshold.
// Threshold ~0.70 catches these perturbations while unrelated tracks sit near 0.5.
func Match(query []uint32, variants [][]uint32, threshold float64) bool {
	for _, v := range variants {
		if Similarity(query, v) >= threshold {
			return true
		}
	}
	return false
}

// Fingerprint returns the raw Chromaprint fingerprint of path via fpcalc.
func Fingerprint(ctx context.Context, path string) ([]uint32, error) {
	out, err := exec.CommandContext(ctx, "fpcalc", "-raw", "-plain", path).Output()
	if err != nil {
		return nil, fmt.Errorf("fpcalc: %w", err)
	}
	return parseRaw(strings.TrimSpace(string(out)))
}

// IndexFingerprints fingerprints path under the whole warp grid. Run once per
// track at ingest; store the result and match uploads against it with Match.
func IndexFingerprints(ctx context.Context, path string) ([][]uint32, error) {
	variants := make([][]uint32, 0, len(warpGrid))
	for _, w := range warpGrid {
		src := path
		if w.semitones != 0 || w.tempo != 1.0 {
			tmp, err := renderWarp(ctx, path, w.semitones, w.tempo)
			if err != nil {
				return nil, err
			}
			defer os.Remove(tmp)
			src = tmp
		}
		fp, err := Fingerprint(ctx, src)
		if err != nil {
			return nil, err
		}
		variants = append(variants, fp)
	}
	return variants, nil
}

// renderWarp writes a pitch/tempo-warped mono copy of in to a temp wav.
func renderWarp(ctx context.Context, in string, semitones, tempo float64) (string, error) {
	p := math.Pow(2, semitones/12)
	// Normalise to 44100 first so asetrate's pitch factor is rate-independent.
	af := fmt.Sprintf("aresample=44100,asetrate=44100*%.6f,aresample=44100,atempo=%.6f,atempo=%.6f",
		p, 1/p, tempo)
	out := filepath.Join(os.TempDir(), "detect-"+strconv.FormatInt(int64(semitones*1000)+int64(tempo*1000)+1<<20, 16)+".wav")
	cmd := exec.CommandContext(ctx, ffmpeg.FindBinary(), "-y", "-i", in, "-ac", "1", "-af", af, out)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("warp render: %w", err)
	}
	return out, nil
}

func parseRaw(s string) ([]uint32, error) {
	if s == "" {
		return nil, fmt.Errorf("empty fingerprint")
	}
	parts := strings.Split(s, ",")
	fp := make([]uint32, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad fingerprint value %q: %w", p, err)
		}
		fp[i] = uint32(int32(v)) // fpcalc emits signed 32-bit ints
	}
	return fp, nil
}
