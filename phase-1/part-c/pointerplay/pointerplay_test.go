package pointerplay_test

import (
	"testing"
	"pointerplay"
)

func TestDouble(t *testing.T) {
	t.Parallel()

	x := pointerplay.MyInt(12)
	want := pointerplay.MyInt(24)

	p := &x
	p.Double()

	if want != x {
		t.Errorf("want %d, got %d", want, x)
	}
}

