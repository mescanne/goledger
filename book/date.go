package book

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

/*
* Thinking:
* Need a way to express:
*   - absolute points in time that are aligned to
*     year, quarter, month, week, day.
*     Including: "today"
*     Relative to a specific date
*     Can be done cumulatively
*   - a duration of time
*   - go methods to find the absolute point
*     and then do the duration-of-time-math
*
* Spec for absolute point in time (asof date):
*  - absolute point:
*     YYYYMMDD or YYYY-MM-DD or YYYY/MM/DD or YYYY.MM.DD or today   --> today is implicit
*  - duration and alignment type:
*     (week|day|month|quarter|year)    --> day is implicit
*  - relative numeric is:
*     ([+-]?[1-9][0-9]*) <duration>    --> <duration> is implicitly the alignment as well
*  - relative textual is:
*     (this|last|next) <alignment>   --> this is implicit.
*
* Or the above are separate commands.
* Starting with 'asof' being today.
* -> asof => absolute point
* -> shift => shift asof in a duration
* -> asof =>
*
* Implementation:
*  -> match absolute point.
*  -> if nothing left, just stick with that.
*  -> If no match, pick today.
*  -> While more left to match:
*     -> If match number:
*        -> Match duration or error
*        -> Move by that number in duration
*        -> Set alignment (current) to the duartion
*     -> If match (this|last|next)
*        -> Match duration or error.
*        -> Shift to the alignment specified.
*        -> Break loop.
*      -> If match duration
*        -> Shift to the alignment (as per this)
*        -> Break loop
*   -> After loop, expect nothing less to match. Or error.
*
*
* Spec for duration (going back in time):
* - duration:
*     all   --> implicit
*     ([1-9][0-9]*) <duration>
*     <duration>           --> 1 is implicit
*
 */

// Date is a simple calendar date YYYYMMDD representated as an integer
type Date int

// Return the year as a string
func (date Date) GetYear() string {
	return fmt.Sprintf("%04d", int(date/10000))
}

var FloorTypes = []string{"none", "yearly", "quarterly", "monthly", "today"}

func (date Date) Floor(by string) Date {
	if by == "yearly" {
		return date.FloorYear(0)
	} else if by == "quarterly" {
		return date.FloorQuarter(0)
	} else if by == "monthly" {
		return date.FloorMonth(0)
	} else if by == "today" {
		return GetToday()
	} else {
		return date
	}
}

//
// Need an expression for "time distance" relative to current time
//

func (date Date) FloorYear(diff int) Date {
	return Date(((int(date/10000) + diff) * 10000) + 101)
}

func (date Date) GetYearMonth() string {
	return fmt.Sprintf("%04d/%02d", int(date/10000), int(date/100)%100)
}

func (date Date) FloorMonth(diff int) Date {
	year := int(date / 10000)
	month := int(date/100)%100 - 1 + diff
	for month > 11 {
		month -= 12
		year++
	}
	for month < 0 {
		month += 12
		year--
	}
	return Date((year * 10000) + ((month + 1) * 100) + 1)
}

func (date Date) GetYearQuarter() string {
	return fmt.Sprintf("%04d Q%d", int(date/10000), ((int(date/100)%100-1)/3)+1)
}

func (date Date) FloorQuarter(diff int) Date {
	year := int(date / 10000)
	month := ((int(date/100)%100-1)/3 + diff) * 3
	for month > 11 {
		month -= 12
		year++
	}
	for month < 0 {
		month += 12
		year--
	}
	return Date((year * 10000) + ((month + 1) * 100) + 1)
}

var dateYYYYMMDD = regexp.MustCompile("^([0-9][0-9][0-9][0-9])[-/]]?([0-9][0-9])[-/]?([0-9][0-9])$")
var dateDDMMYYYY = regexp.MustCompile("^([0-9][0-9])[-/]]?([0-9][0-9])[-/]?([0-9][0-9][0-9][0-9])$")
var date_desc_re = regexp.MustCompile("^(this|last|next) +(month|year|quarter)$")

func (date *Date) Set(value string) error {
	*date = DateFromString(value)
	return nil
}

func (date *Date) Type() string {
	return "date"
}

func DateFromString(date string) Date {
	mat := dateYYYYMMDD.FindStringSubmatch(date)
	if mat != nil {
		year, _ := strconv.Atoi(mat[1])
		month, _ := strconv.Atoi(mat[2])
		day, _ := strconv.Atoi(mat[3])
		return Date(year*10000 + month*100 + day)
	}

	mat = dateDDMMYYYY.FindStringSubmatch(date)
	if mat != nil {
		day, _ := strconv.Atoi(mat[1])
		month, _ := strconv.Atoi(mat[2])
		year, _ := strconv.Atoi(mat[3])
		return Date(year*10000 + month*100 + day)
	}

	mat = date_desc_re.FindStringSubmatch(date)
	if mat != nil {
		d := GetToday()
		diff := 0
		if mat[1] == "last" {
			diff = -1
		} else if mat[1] == "next" {
			diff = 1
		}

		if mat[2] == "year" {
			d = d.FloorYear(diff)
		} else if mat[2] == "quarter" {
			d = d.FloorQuarter(diff)
		} else if mat[2] == "month" {
			d = d.FloorMonth(diff)
		}

		return d
	}

	i, err := strconv.Atoi(date)
	if err == nil {
		return Date(i)
	}

	return 0
}

func GetToday() Date {
	n := time.Now()
	return Date((n.Year() * 10000) + (int(n.Month()) * 100) + n.Day())
}

func GetDate(year int, month int, day int) Date {
	d := Date((year * 10000) + (month * 100) + day)
	return d
}

func (date Date) DaysSince(d2 Date) int {
	d := date.GetTime().Sub(d2.GetTime())
	f := int(d.Hours()) / 24
	return f
}

func (date Date) GetTime() time.Time {
	t := time.Date(int(date/10000), (time.Month)((date/100)%100), int(date%100), 0, 0, 0, 0, time.UTC)
	return t
}

func (date Date) String() string {
	return fmt.Sprintf("%04d/%02d/%02d", int(date/10000), int((date/100)%100), int(date%100))
}
