package book

import (
	"testing"
)

func TestBasicDate(t *testing.T) {
	if diff := GetDate(2016, 4, 22).DaysSince(GetDate(2016, 4, 21)); diff != 1 {
		t.Fatalf("Expected 1 day, got %v", diff)
	}

	if diff := GetDate(2016, 4, 2).DaysSince(GetDate(2016, 3, 28)); diff != 5 {
		t.Fatalf("Expected 4 day, got %v", diff)
	}

	if diff := GetDate(2016, 4, 20).String(); diff != "2016/04/20" {
		t.Fatalf("Expected 2016/04/20, got %v", diff)
	}
}
