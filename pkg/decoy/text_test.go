package decoy

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestLoadNgramsBuildsEveryWindow(t *testing.T) {
	d, err := NewDecoy(WithSeed(1), WithNgram(2))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	if err := d.LoadNgrams([]string{"a", "b", "c", "d"}); err != nil {
		t.Fatalf("LoadNgrams returned an error: %v", err)
	}

	if !slices.Equal(d.keys, []string{"a b", "b c"}) {
		t.Fatalf("unexpected keys: %v", d.keys)
	}
	if !slices.Equal(d.ngrams["a b"], []string{"c"}) {
		t.Fatalf("unexpected successors for \"a b\": %v", d.ngrams["a b"])
	}
	if !slices.Equal(d.ngrams["b c"], []string{"d"}) {
		t.Fatalf("unexpected successors for \"b c\": %v", d.ngrams["b c"])
	}
}

func TestLoadNgramsValidation(t *testing.T) {
	d, err := NewDecoy(WithNgram(2))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	tests := [][]string{
		nil,
		{"a", "b"},
		{"a", "b c", "d"},
	}
	for _, words := range tests {
		if err := d.LoadNgrams(words); err == nil {
			t.Fatalf("expected an error for words %v", words)
		}
	}
}

func TestSetNgramRebuildsLoadedCorpus(t *testing.T) {
	d, err := NewDecoy(WithSeed(1), WithNgram(2))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}
	if err := d.LoadNgrams([]string{"a", "b", "c", "d"}); err != nil {
		t.Fatalf("LoadNgrams returned an error: %v", err)
	}

	if err := d.SetNgram(1); err != nil {
		t.Fatalf("SetNgram returned an error: %v", err)
	}
	if !slices.Equal(d.keys, []string{"a", "b", "c"}) {
		t.Fatalf("unexpected keys: %v", d.keys)
	}
	if !slices.Equal(d.ngrams["a"], []string{"b"}) {
		t.Fatalf("unexpected successors for \"a\": %v", d.ngrams["a"])
	}
}

func TestRandomTextWithPrecomputedNgrams(t *testing.T) {
	d, err := NewDecoy(WithSeed(1), WithNgram(2), WithNgrams(map[string][]string{"a   b": {"c"}}))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	text, err := d.RandomText(3)
	if err != nil {
		t.Fatalf("RandomText returned an error: %v", err)
	}
	if text != "c c c" {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestSetNgramRejectsPrecomputedNgrams(t *testing.T) {
	d, err := NewDecoy(WithNgram(2), WithNgrams(map[string][]string{"a b": {"c"}}))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}
	if err := d.SetNgram(1); err == nil {
		t.Fatal("expected an error")
	}
}

func TestRandomTextIsReproducible(t *testing.T) {
	words := strings.Fields("the quick brown fox jumps over the lazy dog the quick brown fox")

	first, err := NewDecoy(WithSeed(7), WithNgram(2))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}
	second, err := NewDecoy(WithSeed(7), WithNgram(2))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	for _, d := range []*Decoy{first, second} {
		if err := d.LoadNgrams(words); err != nil {
			t.Fatalf("LoadNgrams returned an error: %v", err)
		}
	}

	firstText, err := first.RandomText(50)
	if err != nil {
		t.Fatalf("RandomText returned an error: %v", err)
	}
	secondText, err := second.RandomText(50)
	if err != nil {
		t.Fatalf("RandomText returned an error: %v", err)
	}
	if firstText != secondText {
		t.Fatalf("seeded text generation diverged:\n%s\n%s", firstText, secondText)
	}
}

func TestRandomTextErrors(t *testing.T) {
	d, err := NewDecoy(WithSeed(1))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	if _, err := d.RandomText(1); !errors.Is(err, ErrNoCorpus) {
		t.Fatalf("expected ErrNoCorpus, got %v", err)
	}
	if _, err := d.RandomText(-1); err == nil {
		t.Fatal("expected an error for a negative length")
	}
}

func TestRandomTextMaxLength(t *testing.T) {
	d, err := NewDecoy(WithSeed(1), WithNgram(2), WithMaxTextLength(5))
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}
	if err := d.LoadNgrams([]string{"a", "b", "c", "d"}); err != nil {
		t.Fatalf("LoadNgrams returned an error: %v", err)
	}

	if _, err := d.RandomText(6); err == nil {
		t.Fatal("expected an error for a length above the maximum")
	}
	text, err := d.RandomText(5)
	if err != nil {
		t.Fatalf("RandomText returned an error: %v", err)
	}
	if len(strings.Fields(text)) != 5 {
		t.Fatalf("expected 5 words, got %d: %q", len(strings.Fields(text)), text)
	}
}
