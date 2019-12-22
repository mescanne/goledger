package app

import (
	"testing"
)

func TryLength(t *testing.T, s string, e int) {
	l := Length(s)
	if Length(s) != 3 {
		t.Fatalf("expected length %d, got %d in '%s'", e, l, s)
	}
}

func TryPad(t *testing.T, s string, m int, left_justify bool, e string) {
	p := PadString(s, m, left_justify)
	if p != e {
		t.Fatalf("expected padded string '%s', got '%s' from PadString('%s', %d, %v)", e, p, s, m, left_justify)
	}
}

func TestPrintFormat(t *testing.T) {

	s := string(Blue) + "abc" + string(Reset)

	TryLength(t, s, 3)
	TryPad(t, s, 2, true, string(Blue)+"ab"+string(Reset))
	TryPad(t, s, 2, false, string(Blue)+"ab"+string(Reset))
	TryPad(t, s, 5, true, string(Blue)+"abc"+string(Reset)+"  ")
	TryPad(t, s, 5, false, "  "+string(Blue)+"abc"+string(Reset))
	TryPad(t, s, 10, true, string(Blue)+"abc"+string(Reset)+"       ")
	TryPad(t, s, 10, false, "       "+string(Blue)+"abc"+string(Reset))
}
