package main

import (
	"time"
)

const (
	HOLIDAY_EASTER           = "Easter"
	HOLIDAY_COLUMBUS_DAY     = "Columbus Day"
	HOLIDAY_LABOR_DAY        = "Labor Day"
	HOLIDAY_THANKSGIVING     = "Thanksgiving"
	HOLIDAY_MLK_DAY          = "MLK Jr. Day"
	HOLIDAY_PRESIDENTS_DAY   = "President's Day"
	HOLIDAY_MOTHERS_DAY      = "Mother's Day"
	HOLIDAY_ARMED_FORCES_DAY = "Armed Force's Day"
	HOLIDAY_FATHERS_DAY      = "Father's Day"
	HOLIDAY_PARENTS_DAY      = "Parent's Day"
	HOLIDAY_FRIENDSHIP_DAY   = "Friendship Day"
	HOLIDAY_CHRISTMAS        = "Christmas"
	HOLIDAY_VETERANS_DAY     = "Veteran's Day"
	HOLIDAY_HALLOWEEN        = "Halloween"
	HOLIDAY_INDEPENDENCE     = "Independence Day"
	HOLIDAY_JUNETEENTH       = "Juneteenth"
	HOLIDAY_VALENTINES       = "Valentines"
	HOLIDAY_GROUNDHOG        = "Groundhog Day"
	HOLIDAY_NEW_YEARS        = "New Year's Day"
	HOLIDAY_CHILD_BDAY       = "Child's Birthday"
)

const (
	FIRST = iota
	SECOND
	THIRD
	FOURTH
)

var HOLIDAYS = []Holiday{
	// Top Holidays //
	StaticHoliday{HOLIDAY_CHILD_BDAY, time.July, 1},
	StaticHoliday{HOLIDAY_CHRISTMAS, time.December, 25},
	RelativeHoliday{HOLIDAY_THANKSGIVING, time.November, time.Thursday, FOURTH}, // Fourth Thursday of November
	StaticHoliday{HOLIDAY_HALLOWEEN, time.October, 31},
	CalculatedHoliday{HOLIDAY_EASTER},
	RelativeHoliday{HOLIDAY_MOTHERS_DAY, time.May, time.Sunday, SECOND},  // Second Sunday of May
	RelativeHoliday{HOLIDAY_FATHERS_DAY, time.June, time.Sunday, FOURTH}, // Third Sunday of June
	// Rest
	RelativeHoliday{HOLIDAY_MLK_DAY, time.January, time.Monday, THIRD},         // Third Monday of Januaray
	RelativeHoliday{HOLIDAY_PRESIDENTS_DAY, time.February, time.Monday, THIRD}, // Third Monday of Januaray
	RelativeHoliday{HOLIDAY_ARMED_FORCES_DAY, time.May, time.Saturday, THIRD},  // Third Saturday of May
	RelativeHoliday{HOLIDAY_PARENTS_DAY, time.July, time.Sunday, FOURTH},       // Fourth Sunday of June
	RelativeHoliday{HOLIDAY_FRIENDSHIP_DAY, time.August, time.Sunday, FIRST},   // First Sunday of August
	RelativeHoliday{HOLIDAY_LABOR_DAY, time.September, time.Monday, FIRST},     // First Monday of September
	RelativeHoliday{HOLIDAY_COLUMBUS_DAY, time.October, time.Monday, SECOND},   // Second Modnay of October
	StaticHoliday{HOLIDAY_VETERANS_DAY, time.November, 11},
	StaticHoliday{HOLIDAY_INDEPENDENCE, time.July, 4},
	StaticHoliday{HOLIDAY_JUNETEENTH, time.June, 19},
	StaticHoliday{HOLIDAY_VALENTINES, time.February, 14},
	StaticHoliday{HOLIDAY_GROUNDHOG, time.February, 2},
	StaticHoliday{HOLIDAY_NEW_YEARS, time.January, 1},
}

var (
	// Random Defined holidays for testing/defaults
	ColumbusDay     = RelativeHoliday{HOLIDAY_COLUMBUS_DAY, time.October, time.Monday, SECOND}    // Second Modnay of October
	LaborDay        = RelativeHoliday{HOLIDAY_LABOR_DAY, time.September, time.Monday, FIRST}      // First Monday of September
	ThanksgivingDay = RelativeHoliday{HOLIDAY_THANKSGIVING, time.November, time.Thursday, FOURTH} // Fourth Thursday of November
	MLK_Day         = RelativeHoliday{HOLIDAY_MLK_DAY, time.January, time.Monday, THIRD}          // Third Monday of Januaray
	PresidentsDay   = RelativeHoliday{HOLIDAY_PRESIDENTS_DAY, time.February, time.Monday, THIRD}  // Third Monday of Januaray
	MothersDay      = RelativeHoliday{HOLIDAY_PRESIDENTS_DAY, time.May, time.Sunday, SECOND}      // Second Sunday of May
	ArmedForcesDay  = RelativeHoliday{HOLIDAY_ARMED_FORCES_DAY, time.May, time.Saturday, THIRD}   // Third Saturday of May
	FathersDay      = RelativeHoliday{HOLIDAY_FATHERS_DAY, time.June, time.Sunday, FOURTH}        // Third Sunday of June
	ParentsDay      = RelativeHoliday{HOLIDAY_PARENTS_DAY, time.July, time.Sunday, FOURTH}        // Fourth Sunday of June
	FriendshipDay   = RelativeHoliday{HOLIDAY_FRIENDSHIP_DAY, time.August, time.Sunday, FIRST}    // First Sunday of August
	Christmas       = StaticHoliday{HOLIDAY_CHRISTMAS, time.December, 25}
	Veterans        = StaticHoliday{HOLIDAY_VETERANS_DAY, time.November, 11}
	Halloween       = StaticHoliday{HOLIDAY_HALLOWEEN, time.October, 31}
	Independence    = StaticHoliday{HOLIDAY_INDEPENDENCE, time.July, 4}
	Juneteenth      = StaticHoliday{HOLIDAY_JUNETEENTH, time.June, 19}
	Valentines      = StaticHoliday{HOLIDAY_VALENTINES, time.February, 14}
	Groundhog       = StaticHoliday{HOLIDAY_GROUNDHOG, time.February, 2}
	NewYears        = StaticHoliday{HOLIDAY_NEW_YEARS, time.January, 1}
	ChildBirthday   = StaticHoliday{HOLIDAY_CHILD_BDAY, time.February, 1}
)

// / 0 = Sunday
// / 6 = Saturday
type RelativeHoliday struct {
	Name    string
	Month   time.Month
	Weekday time.Weekday
	Week    int
}

func (h RelativeHoliday) CalculateDate(year int, timezone *time.Location) time.Time {
	firstDay := time.Date(year, h.Month, 1, 0, 0, 0, 0, timezone)
	dayNum := (int(h.Weekday) - int(firstDay.Weekday())) % 7
	if dayNum < 0 {
		dayNum += 7
	}
	firstWeekDay := firstDay.AddDate(0, 0, dayNum)
	target := firstWeekDay.AddDate(0, 0, 7*(h.Week))
	return target
}

func (h RelativeHoliday) getName() string {
	return h.Name
}

func (h RelativeHoliday) getMonth() time.Month {
	return h.Month
}

type StaticHoliday struct {
	Name  string
	Month time.Month
	Day   int
}

func (h StaticHoliday) CalculateDate(year int, timezone *time.Location) time.Time {
	return time.Date(year, h.Month, h.Day, 0, 0, 0, 0, timezone)
}

func (h StaticHoliday) getName() string {
	return h.Name
}
func (h StaticHoliday) getMonth() time.Month {
	return h.Month
}

type CalculatedHoliday struct {
	Name string
}

func (h CalculatedHoliday) CalculateDate(year int, timezone *time.Location) time.Time {
	switch h.Name {
	case HOLIDAY_EASTER:
		return calculateEaster(year, timezone)
	default:
		return time.Time{}
	}

}
func (h CalculatedHoliday) getName() string {
	return h.Name
}

func (h CalculatedHoliday) getMonth() time.Month {
	return h.CalculateDate(gStartDate.Year(), gTimezone).Month()
}

type Holiday interface {
	CalculateDate(year int, timezone *time.Location) time.Time
	getName() string
	getMonth() time.Month
}

// This is wild isn't it?
func calculateEaster(year int, tz *time.Location) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, tz)
}
