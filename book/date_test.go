package book

import (
	"testing"
)

func TestBasicDate(t *testing.T) {
	CheckDiff(t, GetDate(2016, 4, 22), GetDate(2016, 4, 21), 1)
	CheckDiff(t, GetDate(2016, 4, 2), GetDate(2016, 3, 28), 5)

	// Leap year
	// By month
	CheckDiff(t, GetDate(2004, 3, 1), GetDate(2004, 2, 1), 29)
	CheckDiff(t, GetDate(2000, 3, 1), GetDate(2000, 2, 1), 29)
	CheckDiff(t, GetDate(1600, 3, 1), GetDate(1600, 2, 1), 29)
	// By year
	CheckDiff(t, GetDate(2005, 2, 1), GetDate(2004, 2, 1), 366)
	CheckDiff(t, GetDate(2001, 2, 1), GetDate(2000, 2, 1), 366)
	CheckDiff(t, GetDate(1601, 2, 1), GetDate(1600, 2, 1), 366)

	// Not Leap year
	// By month
	CheckDiff(t, GetDate(1900, 3, 1), GetDate(1900, 2, 1), 28)
	CheckDiff(t, GetDate(2001, 3, 1), GetDate(2001, 2, 1), 28)
	// By year
	CheckDiff(t, GetDate(1901, 2, 1), GetDate(1900, 2, 1), 365)
	CheckDiff(t, GetDate(2002, 2, 1), GetDate(2001, 2, 1), 365)

	// Check formatting
	if diff := GetDate(2016, 4, 20).String(); diff != "2016/04/20" {
		t.Fatalf("Expected 2016/04/20, got %v", diff)
	}
}

func CheckDiff(t *testing.T, date Date, dateSince Date, days int) {
	if diff := date.DaysSince(dateSince); diff != days {
		t.Fatalf("From %s to %s, expected %d days, got %v", dateSince, date, days, diff)
	}
	CheckDate(t, date)
	CheckDate(t, dateSince)
}

func CheckDate(t *testing.T, date Date) {
	if ndate := GetDateFromDays(date.AsDays()); ndate != date {
		t.Fatalf("Integrity check of %s => days %d => date %s doesn't match", date, date.AsDays(), ndate)
	}
}
