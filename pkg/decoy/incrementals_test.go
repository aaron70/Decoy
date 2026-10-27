package decoy

import (
	"sync"
	"testing"
)

func TestNextIncrementalInt(t *testing.T) {
	d, err := NewDecoy()
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	expected := []int64{10, 12, 14}
	for _, want := range expected {
		if got := d.NextIncrementalInt("id", 10, 2); got != want {
			t.Fatalf("expected %d, got %d", want, got)
		}
	}
}

func TestSetAndDeleteIncrementalInt(t *testing.T) {
	d, err := NewDecoy()
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	if got := d.SetIncrementalInt("id", 100); got != 100 {
		t.Fatalf("expected 100, got %d", got)
	}
	if got := d.NextIncrementalInt("id", 0, 1); got != 101 {
		t.Fatalf("expected 101, got %d", got)
	}
	if !d.DeleteIncrementalInt("id") {
		t.Fatal("expected DeleteIncrementalInt to return true")
	}
	if d.DeleteIncrementalInt("id") {
		t.Fatal("expected DeleteIncrementalInt to return false")
	}
	if got := d.NextIncrementalInt("id", 5, 1); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestConcurrentIncrementalInt(t *testing.T) {
	d, err := NewDecoy()
	if err != nil {
		t.Fatalf("NewDecoy returned an error: %v", err)
	}

	const goroutines = 100
	results := make(chan int64, goroutines)
	var wg sync.WaitGroup

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- d.NextIncrementalInt("id", 0, 1)
		}()
	}

	wg.Wait()
	close(results)

	seen := make(map[int64]struct{}, goroutines)
	for value := range results {
		if value < 0 || value >= goroutines {
			t.Fatalf("value %d is outside [0, %d)", value, goroutines)
		}
		if _, exists := seen[value]; exists {
			t.Fatalf("duplicate value %d", value)
		}
		seen[value] = struct{}{}
	}
	if len(seen) != goroutines {
		t.Fatalf("expected %d unique values, got %d", goroutines, len(seen))
	}
}
