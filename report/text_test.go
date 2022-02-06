package report

import (
	"testing"
)

func LengthRun(t *testing.T, s string, maxcount int, expected_count int, expected_idx int) {
	count, idx := Length(s, maxcount)
	if count != expected_count {
		t.Fatalf("Length('%s', %d) expected char count %d, but got %d", s, maxcount, expected_count, count)
	}
	if idx != expected_idx {
		t.Fatalf("Length('%s', %d) expected index %d, but got %d", s, maxcount, expected_idx, idx)
	}
}

func TestLength(t *testing.T) {
	//const ansiStart = '\033'
	//const ansiEnd = 'm'
	//func Length(s string, maxcount int) (count int, idx int) {

	// Normal ASCII
	LengthRun(t, "abc", 10, 3, 3)
	LengthRun(t, "abc", 3, 3, 3)
	LengthRun(t, "abc", 2, 2, 2)

	// With UTF-8 type-characters
	LengthRun(t, "£100", 10, 4, 5)
	LengthRun(t, "£100", 4, 4, 5)
	LengthRun(t, "£100", 3, 3, 4)
	LengthRun(t, "£100", 2, 2, 3)

	// With ANSI-encoded strings
	LengthRun(t, "£10\033[33;2m0", 10, 4, 12)
	LengthRun(t, "£10\033[33;2m0", 4, 4, 12)
	LengthRun(t, "£10\033[33;2m0", 3, 3, 4)
}
