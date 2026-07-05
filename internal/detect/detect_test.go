package detect

import "testing"

// deterministic pseudo-fingerprint so the test never flakes.
func fp(seed uint32, n int) []uint32 {
	out := make([]uint32, n)
	x := seed | 1
	for i := range out {
		x = x*2654435761 + 40503 // Knuth LCG
		out[i] = x
	}
	return out
}

// flipBits flips one bit in every kth word — stands in for a perturbed upload
// whose fingerprint has drifted from the reference.
func flipBits(a []uint32, everyK int) []uint32 {
	out := append([]uint32(nil), a...)
	for i := range out {
		if i%everyK == 0 {
			out[i] ^= 1 << uint(i%32)
		}
	}
	return out
}

func TestSimilarityDistinguishes(t *testing.T) {
	ref := fp(1, 300)

	if s := Similarity(ref, ref); s != 1.0 {
		t.Errorf("identical: got %v, want 1.0", s)
	}
	// A lightly-drifted copy must still score well above unrelated.
	drift := Similarity(ref, flipBits(ref, 8))
	if drift < 0.9 {
		t.Errorf("drifted copy scored too low: %v", drift)
	}
	// Unrelated fingerprints agree on ~half their bits.
	if s := Similarity(ref, fp(999, 300)); s > 0.65 {
		t.Errorf("unrelated scored too high: %v", s)
	}
	// Time-shifted copy (evader crops/pads) still aligns.
	if s := Similarity(ref, ref[20:]); s < 0.95 {
		t.Errorf("shifted copy scored too low: %v", s)
	}
}

func TestMatchThreshold(t *testing.T) {
	ref := fp(7, 300)
	index := [][]uint32{fp(100, 300), fp(200, 300), ref} // ref is one indexed variant

	if !Match(flipBits(ref, 6), index, 0.70) {
		t.Error("perturbed query should match its indexed variant")
	}
	if Match(fp(555, 300), index, 0.70) {
		t.Error("unrelated query should not match")
	}
}
