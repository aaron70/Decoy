package decoy

import (
	"testing"
)

func TestNewDecoyOptions(t *testing.T) {
	d, err := NewDecoy(
		WithSeed(42),
		WithNgram(2),
		WithNgrams(map[string][]string{"a b": {"c"}}),
	)
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}
	if d.Seed() != 42 {
		t.Fatalf("expected seed 42, got %d", d.Seed())
	}
	if d.ngram != 2 {
		t.Fatalf("expected ngram 2, got %d", d.ngram)
	}
	if len(d.keys) != 1 || d.keys[0] != "a b" {
		t.Fatalf("unexpected keys: %v", d.keys)
	}
}

func TestWithNgramsValidation(t *testing.T) {
	tests := []struct {
		name   string
		ngrams map[string][]string
	}{
		{name: "wrong key size", ngrams: map[string][]string{"a": {"b"}}},
		{name: "no successors", ngrams: map[string][]string{"a b": {}}},
		{name: "whitespace successor", ngrams: map[string][]string{"a b": {"c d"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewDecoy(WithNgram(2), WithNgrams(test.ngrams))
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestWithNgramsCanonicalizesKeys(t *testing.T) {
	d, err := NewDecoy(WithNgram(2), WithNgrams(map[string][]string{"a   b": {"c"}}))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}
	if len(d.keys) != 1 || d.keys[0] != "a b" {
		t.Fatalf("unexpected keys: %v", d.keys)
	}
}

func TestZeroValueDecoy(t *testing.T) {
	var d Decoy

	value, err := d.RandomInt(1, 2)
	if err != nil {
		t.Fatalf("RandomInt returned an error: %v", err)
	}
	if value != 1 {
		t.Fatalf("expected 1, got %d", value)
	}

	if got := d.NextIncrementalInt("id", 3, 1); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}

	if err := d.LoadNgrams([]string{"a", "b", "c", "d", "e"}); err != nil {
		t.Fatalf("LoadNgrams returned an error: %v", err)
	}
	if _, err := d.RandomText(1); err != nil {
		t.Fatalf("RandomText returned an error: %v", err)
	}
}
