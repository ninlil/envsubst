package envsubst

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

// TestNewReplacer_Defaults verifies that a freshly created Replacer uses the
// default prefix '$' and wrapper '()'.
func TestNewReplacer_Defaults(t *testing.T) {
	r := NewReplacer()
	if r.prefix != defaultPrefix {
		t.Errorf("prefix = %q, want %q", r.prefix, defaultPrefix)
	}
	if r.start != defaultStart || r.end != defaultEnd {
		t.Errorf("wrapper = %q%q, want %q%q", r.start, r.end, defaultStart, defaultEnd)
	}
}

// TestNewReplacer_Options verifies the WithPrefix and WithWrapper options.
func TestNewReplacer_Options(t *testing.T) {
	r := NewReplacer(WithPrefix('%'), WithWrapper('{'))
	if r.prefix != '%' {
		t.Errorf("prefix = %q, want %q", r.prefix, '%')
	}
	if r.start != '{' || r.end != '}' {
		t.Errorf("wrapper = %q%q, want %q%q", r.start, r.end, '{', '}')
	}
}

// TestNewReplacer_InvalidOptions verifies that invalid prefix/wrapper options
// leave the Replacer on its defaults instead of corrupting it.
func TestNewReplacer_InvalidOptions(t *testing.T) {
	r := NewReplacer(WithPrefix('!'), WithWrapper('?'))
	if r.prefix != defaultPrefix {
		t.Errorf("prefix = %q, want default %q", r.prefix, defaultPrefix)
	}
	if r.start != defaultStart || r.end != defaultEnd {
		t.Errorf("wrapper = %q%q, want default %q%q", r.start, r.end, defaultStart, defaultEnd)
	}
}

// TestReplacer_ConvertString exercises the per-instance ConvertString with a
// custom prefix and wrapper, independent of the package-level defaultReplacer.
func TestReplacer_ConvertString(t *testing.T) {
	x := map[string]string{"X": "+"}

	r := NewReplacer(WithPrefix('%'), WithWrapper('{'))

	tests := []struct {
		name    string
		str     string
		want    string
		wantErr bool
	}{
		{"empty", "", "", false},
		{"no-change", "test", "test", false},
		{"simple-var", "1%{X}2", "1+2", false},
		{"double-var", "1%{X}%{X}2", "1++2", false},
		{"single-prefix", "1%2", "1%2", false},
		{"not-ended", "1%{X", "1%{X", false},
		{"empty-wrap", "%{}", "%{}", false},
		{"missing", "%{NOPE}", "", true},
		{"wrong-wrapper", "1${X}2", "1${X}2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.ConvertString(tt.str, Map(x))
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertString() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestReplacer_ConvertBytes verifies the per-instance ConvertBytes.
func TestReplacer_ConvertBytes(t *testing.T) {
	x := map[string]string{"X": "+"}
	r := NewReplacer()

	got, err := r.ConvertBytes([]byte("1$(X)2"), Map(x))
	if err != nil {
		t.Fatalf("ConvertBytes() error = %v", err)
	}
	if !bytes.Equal(got, []byte("1+2")) {
		t.Errorf("ConvertBytes() = %q, want %q", got, "1+2")
	}

	// empty input returns same empty slice without error
	got, err = r.ConvertBytes([]byte{}, Map(x))
	if err != nil {
		t.Fatalf("ConvertBytes(empty) error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ConvertBytes(empty) = %q, want empty", got)
	}
}

// TestReplacer_Convert verifies the streaming Convert and that a nil mapping
// falls back to Getenv.
func TestReplacer_Convert(t *testing.T) {
	x := map[string]string{"X": "+"}
	r := NewReplacer()

	var out bytes.Buffer
	if err := r.Convert(strings.NewReader("1$(X)2"), &out, Map(x)); err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	if out.String() != "1+2" {
		t.Errorf("Convert() = %q, want %q", out.String(), "1+2")
	}

	// nil mapping should default to Getenv and not panic
	t.Setenv("ENVSUBST_TEST_VAR", "yes")
	out.Reset()
	if err := r.Convert(strings.NewReader("v=$(ENVSUBST_TEST_VAR)"), &out, nil); err != nil {
		t.Fatalf("Convert(nil mapping) error = %v", err)
	}
	if out.String() != "v=yes" {
		t.Errorf("Convert(nil mapping) = %q, want %q", out.String(), "v=yes")
	}
}

// TestReplacer_MissingError verifies the error type and message for a missing
// variable.
func TestReplacer_MissingError(t *testing.T) {
	r := NewReplacer()
	_, err := r.ConvertString("$(MISSING)", Map(nil))
	if err == nil {
		t.Fatal("expected error for missing variable, got nil")
	}
	if got, want := err.Error(), "field 'MISSING' is missing"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

// TestReplacer_Independence verifies that two Replacers with different settings
// do not interfere with each other.
func TestReplacer_Independence(t *testing.T) {
	x := map[string]string{"X": "+"}

	a := NewReplacer() // $()
	b := NewReplacer(WithPrefix('%'), WithWrapper('{'))

	gotA, err := a.ConvertString("1$(X)2", Map(x))
	if err != nil {
		t.Fatalf("a.ConvertString() error = %v", err)
	}
	gotB, err := b.ConvertString("1%{X}2", Map(x))
	if err != nil {
		t.Fatalf("b.ConvertString() error = %v", err)
	}
	if gotA != "1+2" || gotB != "1+2" {
		t.Errorf("gotA = %q, gotB = %q, want both %q", gotA, gotB, "1+2")
	}

	// a must not understand b's syntax and vice-versa
	if got, _ := a.ConvertString("1%{X}2", Map(x)); got != "1%{X}2" {
		t.Errorf("a converted b-syntax: got %q", got)
	}
}

// TestReplacer_Concurrent runs many goroutines sharing a single Replacer to
// ensure the Convert* methods are safe for concurrent use.
func TestReplacer_Concurrent(t *testing.T) {
	r := NewReplacer()
	x := map[string]string{"X": "+"}
	mapping := Map(x)

	const goroutines = 50
	const iterations = 10000

	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				got, err := r.ConvertString("1$(X)$(X)2", mapping)
				if err != nil {
					errs <- err
					return
				}
				if got != "1++2" {
					errs <- &concurrentMismatch{got: got}
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent Convert failed: %v", err)
	}
}

type concurrentMismatch struct {
	got string
}

func (e *concurrentMismatch) Error() string {
	return "unexpected result: " + e.got
}
