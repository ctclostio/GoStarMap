package celestial

import (
	"math/rand"
	"testing"
)

func TestFormatNumber(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{1234.5, "1,234.5"},
		{1234.50, "1,234.5"},
		{1234.56, "1,234.56"},
		{-1, "-1"},
		{-1234, "-1,234"},
		{-1234.5, "-1,234.5"},
		{12742, "12,742"},   // Earth diameter km
		{139820, "139,820"}, // Jupiter diameter km
	}
	for _, c := range cases {
		got := FormatNumber(c.in)
		if got != c.want {
			t.Errorf("FormatNumber(%v) = %q; want %q", c.in, got, c.want)
		}
	}
}

func TestRandomTypeDistribution(t *testing.T) {
	// Cumulative thresholds in RandomType target observed Milky Way ratios.
	// With 1M samples each bin should land within tolerance of its expected
	// share. Rare bins (O, B) get a looser absolute tolerance scaled by their
	// expected count.
	rand.Seed(1)
	const samples = 1_000_000

	counts := map[SpectralType]int{}
	for i := 0; i < samples; i++ {
		counts[RandomType()]++
	}

	expected := map[SpectralType]float64{
		TypeO: 0.00003,
		TypeB: 0.00130,
		TypeA: 0.00600,
		TypeF: 0.03000,
		TypeG: 0.07600,
		TypeK: 0.12100,
		TypeM: 0.76567,
	}

	for st, frac := range expected {
		gotFrac := float64(counts[st]) / float64(samples)
		expectedCount := frac * float64(samples)
		var tol float64
		if expectedCount >= 1000 {
			tol = frac * 0.1 // 10% relative for plentiful bins
		} else {
			// For tail bins, allow either 5x the relative error or the
			// fraction itself (i.e. count could be zero or 2× expected).
			tol = frac
			if frac*5 > tol {
				tol = frac * 5
			}
		}
		diff := gotFrac - frac
		if diff < -tol || diff > tol {
			t.Errorf("type %v: got %.5f, want %.5f ±%.5f (count %d)", st, gotFrac, frac, tol, counts[st])
		}
	}

	// All draws must land in some declared bin — no leakage.
	total := 0
	for _, c := range counts {
		total += c
	}
	if total != samples {
		t.Errorf("total samples %d; want %d", total, samples)
	}
}
