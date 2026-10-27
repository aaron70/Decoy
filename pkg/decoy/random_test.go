package decoy

import (
	"math"
	"testing"
)

func TestRandomIntBounds(t *testing.T) {
	d, err := NewDecoy(WithSeed(1))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	for range 1000 {
		value, err := d.RandomInt(-10, 10)
		if err != nil {
			t.Fatalf("RandomInt returned an error: %v", err)
		}
		if value < -10 || value >= 10 {
			t.Fatalf("value %d is outside [-10, 10)", value)
		}
	}

	if _, err := d.RandomInt(1, 1); err == nil {
		t.Fatal("expected an error for an empty range")
	}
}

func TestRandomFloatBounds(t *testing.T) {
	d, err := NewDecoy(WithSeed(2))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	for range 1000 {
		value, err := d.RandomFloat(-1.5, 2.5)
		if err != nil {
			t.Fatalf("RandomFloat returned an error: %v", err)
		}
		if value < -1.5 || value >= 2.5 {
			t.Fatalf("value %v is outside [-1.5, 2.5)", value)
		}
	}

	invalid := [][2]float64{
		{math.NaN(), 1},
		{0, math.Inf(1)},
		{1, 1},
	}
	for _, limits := range invalid {
		if _, err := d.RandomFloat(limits[0], limits[1]); err == nil {
			t.Fatalf("expected an error for limits %v", limits)
		}
	}
}

func TestRandomChoice(t *testing.T) {
	d, err := NewDecoy(WithSeed(3))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	for range 100 {
		value, err := d.RandomChoice("a", "b", "c")
		if err != nil {
			t.Fatalf("RandomChoice returned an error: %v", err)
		}
		if value != "a" && value != "b" && value != "c" {
			t.Fatalf("unexpected choice %q", value)
		}
	}

	if _, err := d.RandomChoice[string](); err == nil {
		t.Fatal("expected an error for no choices")
	}
}

func TestSeededRandomIsReproducible(t *testing.T) {
	first, err := NewDecoy(WithSeed(99))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}
	second, err := NewDecoy(WithSeed(99))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	for range 100 {
		a, err := first.RandomInt(0, 1_000_000)
		if err != nil {
			t.Fatalf("RandomInt returned an error: %v", err)
		}
		b, err := second.RandomInt(0, 1_000_000)
		if err != nil {
			t.Fatalf("RandomInt returned an error: %v", err)
		}
		if a != b {
			t.Fatalf("seeded generators diverged: %d != %d", a, b)
		}
	}
}

func TestProbabilityExtremes(t *testing.T) {
	d, err := NewDecoy(WithSeed(4))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	if d.Probability(0) {
		t.Fatal("probability 0 returned true")
	}
	if !d.Probability(1) {
		t.Fatal("probability 1 returned false")
	}
	if d.Probability(math.NaN()) {
		t.Fatal("probability NaN returned true")
	}
}
