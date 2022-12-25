package book

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Date is a simple calendar date YYYYMMDD representated as an integer
type Date int

// Return the year as a string
func (date Date) GetYear() string {
	return fmt.Sprintf("%04d", int(date/10000))
}

var FloorTypes = []string{"yearly", "quarterly", "monthly", "today", "none"}

func (date Date) Floor(by string) Date {
	return date.FloorDiff(by, 0)
}

func (date Date) FloorDiff(by string, diff int) Date {
	if by == "yearly" {
		return date.FloorYear(diff)
	} else if by == "quarterly" {
		return date.FloorQuarter(diff)
	} else if by == "monthly" {
		return date.FloorMonth(diff)
	} else if by == "today" {
		return GetToday()
	} else { // May be none or invalid
		return date
	}
}

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

var dateYYYYMMDD = regexp.MustCompile("^([0-9][0-9][0-9][0-9])[-/\\.]?([0-9]?[0-9])?[-/\\.]?([0-9]?[0-9])?$")
var dateDDMMYYYY = regexp.MustCompile("^([0-9]?[0-9])[-/\\.]?([0-9]?[0-9])[-/\\.]?([0-9][0-9][0-9][0-9])$")
var date_desc_re = regexp.MustCompile("^(this|last|next)[\\._ \t]+(month|year|quarter)$")

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
		if month == 0 {
			month = 1
		}
		day, _ := strconv.Atoi(mat[3])
		if day == 0 {
			day = 1
		}
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
	return date.AsDays() - d2.AsDays()
}

func (date Date) AddDays(days int) Date {
	return GetDateFromDays(date.AsDays() + days)
}

var monthDays = [12]int{31, 28, 31, 30, 31,
	30, 31, 31, 30, 31, 30, 31}

func GetDateFromDays(days int) Date {

	// Search for the right year
	year := days / 365
	for {
		leapDays := year/4 + year/400 - year/100
		if year*365+leapDays <= days {
			days -= year*365 + leapDays
			break
		}
		year--
	}

	// Subtract the months
	month := 0
	for {
		daysInMonth := monthDays[month]

		// if it's February, a year dividable by four,
		// and NOT dividable by 100 OR it is diviable by 400...
		// Add a leap day.
		if (month == 1) &&
			((year+1)%4 == 0) &&
			((year+1)%100 != 0 || (year+1)%400 == 0) {
			daysInMonth += 1
		}

		// If it's within the month, stop
		if days < daysInMonth {
			break
		}

		// Shift month
		days -= daysInMonth
		month++
	}

	return GetDate(year+1, month+1, days+1)
}

func (date Date) AsDays() int {
	year := int(date/10000) - 1
	month := int((date/100)%100) - 1
	day := int(date%100) - 1

	// Calculate previous years
	days := year * 365
	days = days + year/4 + year/400 - year/100

	// Add in month-to-date (without leaps)
	for i := 0; i < month; i++ {
		days += monthDays[i]

		// if it's February, a year dividable by four,
		// and NOT dividable by 100 OR it is diviable by 400...
		// Add a leap day.
		if (i == 1) && YearIsLeapYear(year) {
			days += 1
		}
	}

	// Add in day-of-month
	days += day

	return days
}

// Year is zero-based (year 1 is 0, year 2000 is 1999)
func YearIsLeapYear(year int) bool {
	return ((year+1)%4 == 0) && ((year+1)%100 != 0 || (year+1)%400 == 0)
}

// Day-in-year
func (date Date) AsYears() float64 {
	year := int(date/10000) - 1
	month := int((date/100)%100) - 1
	day := int(date%100) - 1

	isLeapYear := YearIsLeapYear(year)

	// Add in month-to-date (without leaps)
	day = 0
	for i := 0; i < month; i++ {
		day += monthDays[i]

		// if it's February, a year dividable by four,
		// and NOT dividable by 100 OR it is diviable by 400...
		// Add a leap day.
		if (i == 1) && isLeapYear {
			day += 1
		}
	}

	daysInYear := 365
	if isLeapYear {
		daysInYear++
	}

	return float64(year) + float64(day)/float64(daysInYear)
}

func (date Date) GetTime() time.Time {
	t := time.Date(int(date/10000), (time.Month)((date/100)%100), int(date%100), 0, 0, 0, 0, time.UTC)
	return t
}

func (date Date) String() string {
	if date == Date(0) {
		return fmt.Sprintf("%10s", " ")
	}
	return fmt.Sprintf("%04d/%02d/%02d", int(date/10000), int((date/100)%100), int(date%100))
}

func (date Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(date.GetTime().Format(time.RFC3339))
}
