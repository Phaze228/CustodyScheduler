package main

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
	// "log"
	// "os"
)

const (
	PARENTING_TIME_PERCENT     = 0.70
	DAYS_AFTER_HOLIDAY_MONTHS  = 2
	DAYS_AFTER_HOLIDAY_DAYS    = 1
	DAYS_BEFORE_HOLIDAY_MONTHS = 14
	DAYS_BEFORE_HOLIDAY_DAYS   = 1
	TRANSITION_MONTHS          = 3
	TRANSITION_DAYS            = 1
	YYYYMMDD                   = "2006-01-02"
	MAX_HOURS_EXCEEDED         = (3 * DAYS_BEFORE_HOLIDAY_MONTHS) * 24
	MAX_TIME_BEFORE_SWITCH_PCT = 1.5
)

type TimeUnit int

const (
	MONTH TimeUnit = iota
	DAY
)

type Parent int8

func (p Parent) String() string {
	switch p {
	case PARENT_A:
		return "Parent A"
	case PARENT_B:
		return "Parent B"
	default:
		return "Both"
	}
}

const (
	PARENT_A Parent = iota
	PARENT_B
	BOTH
)

type Schedule struct {
	Days map[string]Day
}

type PriorityDay struct {
	Day       Holiday
	MattersTo Parent
	CurrTurn  Parent
}

// Switches the current parent's holiday time
// If Mom -> Dad; if Dad -> Mom; Only if both care
func (h *PriorityDay) Switch() {
	if h.MattersTo != BOTH {
		return
	}
	if h.CurrTurn == PARENT_A {
		h.CurrTurn = PARENT_B
	} else {
		h.CurrTurn = PARENT_A
	}
}

func (h PriorityDay) isTurnOf(parent Parent) bool {
	if h.CurrTurn != parent && h.MattersTo == BOTH {
		return false
	}
	return true
}

type DayAllotment struct {
	Count        int
	PriorityDays []Holiday
	Custodian    Parent
	Template     string
}

func (a *DayAllotment) Amount() int {
	return a.Count
}

func (a *DayAllotment) Allotment() (int, int, int) {
	return 0, 0, a.Count
}
func (a *DayAllotment) Parent() Parent {
	return a.Custodian
}

func (a *DayAllotment) Holidays() []Holiday {
	return a.PriorityDays
}

func (a *DayAllotment) TemplateType() string {
	return a.Template
}

func (a DayAllotment) DaysForHoliday() int {
	return DAYS_BEFORE_HOLIDAY_DAYS
}

func (a DayAllotment) Transition() int {
	return TRANSITION_DAYS
}

func (a DayAllotment) AfterHoliday() int {
	return DAYS_AFTER_HOLIDAY_DAYS
}

func (a *DayAllotment) UpdateAllotment(start time.Time) {
	switch a.TemplateType() {
	case "3-4-4-3":
		if a.Count == 3 {
			a.Count = 4
		} else {
			a.Count = 3
		}
	case "2-2-5-5":
		if a.Count == 2 {
			a.Count = 5
		} else {
			a.Count = 2
		}

	default:
		return
	}
}

func (a *DayAllotment) AppendDays(start, yearEnd *time.Time, b Allotment, holidays *[]PriorityDay, dates *map[string]Day) {
	count := 0
	var parent Allotment
	aLeft := time.Duration(0)
	bLeft := time.Duration(0)
	left := time.Duration(0)

	for start.Before(*yearEnd) {
		if count%2 == 1 {
			parent = b
			aLeft = left
			left = bLeft
		} else {
			parent = a
			bLeft = left
			left = aLeft
		}
		count++

		if left.Hours() > MAX_HOURS_EXCEEDED {
			log.Fatalf("MAX HOURS OF %d EXCEEDED---LeftOver: %.1f\n", MAX_HOURS_EXCEEDED, left.Hours())
		}

		// switch parent.TemplateType() {
		// case: "3-4-4-3":
		//
		// }

		end, holidayList := newEndFromHolidays(parent, start, &left, holidays)
		appendDaysToSchedule(start, &end, parent, holidayList, dates)
		parent.UpdateAllotment(*start)
	}
	switchMonths := ((gEndDate.Sub(gStartDate).Hours() / 24) / 30) / float64(count)
	fmt.Printf("--- %d switches for %d days : Approximately %2.f per month ---\n", count, int(gEndDate.Sub(gStartDate).Hours()/24), switchMonths)

}

type MonthAllotment struct {
	Count        int
	PriorityDays []Holiday
	Custodian    Parent
	Template     string
	NextCount    int
}

func (a *MonthAllotment) Amount() int {
	return a.Count
}

func (a *MonthAllotment) Allotment() (int, int, int) {
	return 0, a.Count, 0
}
func (a *MonthAllotment) Parent() Parent {
	return a.Custodian
}

func (a *MonthAllotment) Holidays() []Holiday {
	return a.PriorityDays
}

func (a MonthAllotment) TemplateType() string {
	return ""
}

func (a MonthAllotment) DaysForHoliday() int {
	return DAYS_BEFORE_HOLIDAY_MONTHS
}

func (a MonthAllotment) Transition() int {
	return TRANSITION_MONTHS
}

func (a MonthAllotment) AfterHoliday() int {
	return DAYS_BEFORE_HOLIDAY_MONTHS
}

func (a *MonthAllotment) UpdateAllotment(start time.Time) {
	switch a.TemplateType() {
	case "3-4-4-3":
		if a.Count == 3 {
			a.Count = 4
		} else {
			a.Count = 3
		}
	case "2-2-5-5":
		if a.Count == 2 {
			a.Count = 5
		} else {
			a.Count = 2
		}

	default:
		return
	}
}

func (a *MonthAllotment) AppendDays(start, yearEnd *time.Time, b Allotment, holidays *[]PriorityDay, dates *map[string]Day) {
	count := 0
	var parent Allotment
	aLeft := time.Duration(0)
	bLeft := time.Duration(0)
	left := time.Duration(0)

	for start.Before(*yearEnd) {
		if count%2 == 1 {
			parent = b
			aLeft = left
			left = bLeft
		} else {
			parent = a
			bLeft = left
			left = aLeft
		}
		count++

		if left.Hours() > MAX_HOURS_EXCEEDED {
			log.Fatalf("MAX HOURS OF %d EXCEEDED---LeftOver: %.1f\n", MAX_HOURS_EXCEEDED, left.Hours())
		}

		end, holidayList := newEndFromHolidays(parent, start, &left, holidays)
		appendDaysToSchedule(start, &end, parent, holidayList, dates)
		parent.UpdateAllotment(*start)

	}

	switchMonths := ((gEndDate.Sub(gStartDate).Hours() / 24) / 30) / float64(count)
	fmt.Printf("--- %d switches for %d days : Approximately %2.f per month ---\n", count, int(gEndDate.Sub(gStartDate).Hours()/24), switchMonths)
}

type Allotment interface {
	Amount() int
	Allotment() (int, int, int)
	Parent() Parent
	Holidays() []Holiday
	TemplateType() string
	AppendDays(start, yearEnd *time.Time, b Allotment, holidays *[]PriorityDay, dates *map[string]Day)
	DaysForHoliday() int
	Transition() int
	AfterHoliday() int
	UpdateAllotment(start time.Time)
}

func newEndFromHolidays(parent Allotment, start *time.Time, leftover *time.Duration, holidays *[]PriorityDay) (time.Time, map[string][]string) {
	left := *leftover
	typicalEnd := start.AddDate(parent.Allotment())
	maxDuration := getMaxAllowedDays(parent, *start)
	durationLeft := left
	if maxDuration < left {
		left = left - maxDuration
		durationLeft = maxDuration
	} else {
		durationLeft = left
		left = 0
	}
	end := start.Add(durationLeft).AddDate(parent.Allotment())
	months := make([]time.Month, 0)
	transitionDate := end.AddDate(0, 0, parent.DaysForHoliday())
	for s := *start; s.Before(transitionDate); s = s.AddDate(0, 1, 0) {
		months = append(months, s.Month())
	}
	holidayList := make(map[string][]string)
	tmpEnd := end
	for i := 0; i < len(*holidays); i++ {
		h := (*holidays)[i]

		//Ignore holidays not during parenting time
		if !slices.Contains(months, h.Day.getMonth()) {
			continue
		}

		hDate := h.Day.CalculateDate(start.Year(), gTimezone)
		if end.Year() > start.Year() && hDate.Before(*start) {
			hDate = h.Day.CalculateDate(end.Year(), gTimezone)

		}
		holidayList[hDate.Format(YYYYMMDD)] = append(holidayList[hDate.Format(YYYYMMDD)], h.Day.getName())
		// Skip holidays that have passed the start point
		if hDate.Before(*start) {
			continue
		}
		if h.isTurnOf(parent.Parent()) {
			if hDate.After(end) && hDate.Before(transitionDate) {
				newEnd := hDate.AddDate(0, 0, parent.AfterHoliday())
				left += end.Sub(newEnd)
				(*holidays)[i].Switch()
				fmt.Printf("[%s] Keeps for %s; Added %d to their time, they'll lose it next\n\tOriginal: %s - \tNew: %s\n", parent.Parent().String(), h.Day.getName(), int(left.Hours()/24), end.Format(YYYYMMDD), newEnd.Format(YYYYMMDD))
				tmpEnd = newEnd
				continue
			} else if hDate.Before(end) {
				(*holidays)[i].Switch()
			}
		} else {
			allotedDays := typicalEnd.Sub(*start).Hours() / 24
			daysCut := typicalEnd.Sub(hDate).Hours() / 24
			timePercent := (allotedDays - daysCut) / allotedDays
			if hDate.After(*start) && hDate.Before(typicalEnd) && timePercent >= PARENTING_TIME_PERCENT {
				newEnd := hDate.AddDate(0, 0, -parent.Transition())
				left += end.Sub(newEnd)
				fmt.Printf("[%s] loses  %d days due to %s;\n\tOriginal: %s - \tNew: %s\n", parent.Parent().String(), int(left.Hours()/24), h.Day.getName(), end.Format(YYYYMMDD), newEnd.Format(YYYYMMDD))
				tmpEnd = newEnd
				break
			} else if hDate.After(*start) && hDate.Before(typicalEnd) && timePercent <= PARENTING_TIME_PERCENT {
				fmt.Printf("[%s] kept %s day - %d due to parenting time constraints: %.2f%% of parenting time when holiday hits\n", parent.Parent().String(), h.Day.getName(), hDate.Year(), timePercent*100)
				continue
			}
		}
	}
	end = tmpEnd
	return end, holidayList
}

func appendDaysToSchedule(start, endP *time.Time, parent Allotment, holidayList map[string][]string, dates *map[string]Day) {
	var color string
	end := *endP

	for start.Before(end) {
		if parent.Parent() == PARENT_A {
			gParentACount++
			color = A_COLOR
		} else {
			gParentBCount++
			color = B_COLOR
		}
		highlight := false
		if len(holidayList[start.Format(YYYYMMDD)]) > 0 {
			highlight = true
			if parent.Parent() == PARENT_A {
				color = A_HOL_COLOR
			} else {
				color = B_HOL_COLOR
			}
		}
		newDay := Day{
			*start,
			start.Day(),
			parent.Parent(),
			highlight,
			color,
			holidayList[start.Format(YYYYMMDD)],
		}
		(*dates)[start.Format(YYYYMMDD)] = newDay
		*start = start.AddDate(0, 0, 1)
		gEndDate = *start
	}
}

func CreateSchedule(pa, pb Allotment, startDate, endDate time.Time) Schedule {
	if pa.Parent() != PARENT_A {
		log.Printf("parent b is first on schedule\n")
	}
	DASHES := strings.Repeat("-", 10)
	fmt.Printf("\n%s[Parental Debug]%s\n\n", DASHES, DASHES)
	gParentACount = 0
	gParentBCount = 0
	holidays := CreateEntries(pa, pb)
	schedule_dates := make(map[string]Day, 0)
	pa.AppendDays(&startDate, &endDate, pb, &holidays, &schedule_dates)
	fmt.Printf("\n%s [END] %s\n", DASHES, DASHES)
	return Schedule{
		schedule_dates,
	}
}

// func CreateSchedule(pa, pb Allotment, startDate, endDate time.Time) Schedule {
// 	if pa.Parent() != PARENT_A {
// 		log.Printf("parent b is first on schedule\n")
// 	}
// 	DASHES := strings.Repeat("-", 10)
// 	fmt.Printf("\n%s[Parental Debug]%s\n\n", DASHES, DASHES)
// 	gParentACount = 0
// 	gParentBCount = 0
// 	holidays := CreateEntries(pa, pb)
// 	dleftover := time.Duration(0)
// 	mleftover := time.Duration(0)
// 	schedule_dates := make(map[string]Day, 0)
// 	switched := 0
// 	curYear := startDate.Year()
//
// 	for startDate.Before(endDate) {
// 		getDays(&startDate, pa, &holidays, &schedule_dates, &dleftover)
// 		if startDate.Year() > curYear {
// 			fmt.Printf("[%d] - %d switches\n", curYear, switched)
// 			switched = 0
// 			curYear = startDate.Year()
// 		}
// 		switched++
// 		getDays(&startDate, pb, &holidays, &schedule_dates, &mleftover)
// 		if startDate.Year() > curYear {
// 			fmt.Printf("[%d] - %d switches\n", curYear, switched)
// 			switched = 0
// 			curYear = startDate.Year()
// 		}
// 		switched++
// 	}
// 	fmt.Printf("\n%s [END] %s\n", DASHES, DASHES)
//
// 	return Schedule{
// 		schedule_dates,
// 	}
// }

func getMaxAllowedDays(p Allotment, s time.Time) time.Duration {
	e := s.AddDate(p.Allotment())
	duration := e.Sub(s)
	maxDuration := float64(duration) * MAX_TIME_BEFORE_SWITCH_PCT
	return time.Duration(int64(maxDuration))
}

// Go through holidays and find the ones in the time allotted for the specified parent
// if the holidays are specified for the other parent then cut current parent's time short
// Provide a leftover to be made up for the following time.

func getDays(start *time.Time, p Allotment, holidays *[]PriorityDay, dates *map[string]Day, left *time.Duration) {
	if left.Hours() > MAX_HOURS_EXCEEDED {
		log.Fatalf("MAX HOURS OF %d EXCEEDED---LeftOver: %.1f\n", MAX_HOURS_EXCEEDED, left.Hours())
	}

	typicalEnd := start.AddDate(p.Allotment())
	maxDuration := getMaxAllowedDays(p, *start)
	durationLeft := *left
	if maxDuration < *left {
		*left = *left - maxDuration
		durationLeft = maxDuration
	} else {
		durationLeft = *left
		*left = 0
	}
	end := start.Add(durationLeft).AddDate(p.Allotment())
	months := make([]time.Month, 0)
	transitionDate := end.AddDate(0, 0, DAYS_BEFORE_HOLIDAY_MONTHS)
	for s := *start; s.Before(transitionDate); s = s.AddDate(0, 1, 0) {
		months = append(months, s.Month())
	}
	holidayList := make(map[string][]string)
	tmpEnd := end
	for i := 0; i < len(*holidays); i++ {
		h := (*holidays)[i]

		//Ignore holidays not during parenting time
		if !slices.Contains(months, h.Day.getMonth()) {
			continue
		}

		hDate := h.Day.CalculateDate(start.Year(), gTimezone)
		if end.Year() > start.Year() && hDate.Before(*start) {
			hDate = h.Day.CalculateDate(end.Year(), gTimezone)

		}
		holidayList[hDate.Format(YYYYMMDD)] = append(holidayList[hDate.Format(YYYYMMDD)], h.Day.getName())
		// Skip holidays that have passed the start point
		if hDate.Before(*start) {
			continue
		}

		if h.isTurnOf(p.Parent()) {
			if hDate.After(end) && hDate.Before(transitionDate) {
				newEnd := hDate.AddDate(0, 0, DAYS_AFTER_HOLIDAY_MONTHS)
				*left += end.Sub(newEnd)
				(*holidays)[i].Switch()
				fmt.Printf("[%s] Keeps for %s; Added %d to their time, they'll lose it next\n\tOriginal: %s - \tNew: %s\n", p.Parent().String(), h.Day.getName(), int(left.Hours()/24), end.Format(YYYYMMDD), newEnd.Format(YYYYMMDD))
				tmpEnd = newEnd
				continue
			} else if hDate.Before(end) {
				(*holidays)[i].Switch()
			}
		} else {
			allotedDays := typicalEnd.Sub(*start).Hours() / 24
			daysCut := typicalEnd.Sub(hDate).Hours() / 24
			timePercent := (allotedDays - daysCut) / allotedDays
			// timePercent := (daysCut.Hours() / 24) / (allotedDays.Hours() / 24)
			if hDate.After(*start) && hDate.Before(typicalEnd) && timePercent >= PARENTING_TIME_PERCENT {
				newEnd := hDate.AddDate(0, 0, -TRANSITION_MONTHS)
				*left += end.Sub(newEnd)
				fmt.Printf("[%s] loses  %d days due to %s;\n\tOriginal: %s - \tNew: %s\n", p.Parent().String(), int(left.Hours()/24), h.Day.getName(), end.Format(YYYYMMDD), newEnd.Format(YYYYMMDD))
				tmpEnd = newEnd
				break
			} else if hDate.After(*start) && hDate.Before(typicalEnd) && timePercent <= PARENTING_TIME_PERCENT {
				fmt.Printf("[%s] kept %s day - %d due to parenting time constraints: %.2f%% of parenting time when holiday hits\n", p.Parent().String(), h.Day.getName(), hDate.Year(), timePercent*100)
				continue

			}

		}

	}
	end = tmpEnd
	var color string

	for start.Before(end) {
		if p.Parent() == PARENT_A {
			gParentACount++
			color = A_COLOR
		} else {
			gParentBCount++
			color = B_COLOR
		}
		highlight := false
		if len(holidayList[start.Format(YYYYMMDD)]) > 0 {
			highlight = true
			if p.Parent() == PARENT_A {
				color = A_HOL_COLOR
			} else {
				color = B_HOL_COLOR
			}
		}
		newDay := Day{
			*start,
			start.Day(),
			p.Parent(),
			highlight,
			color,
			holidayList[start.Format(YYYYMMDD)],
		}
		(*dates)[start.Format(YYYYMMDD)] = newDay
		*start = start.AddDate(0, 0, 1)
		gEndDate = *start
	}

}

func CreateEntries(p1, p2 Allotment) []PriorityDay {
	entries := make([]PriorityDay, 0)
	emap := make(map[string]PriorityDay, 0)
	for _, h1 := range p1.Holidays() {
		hName := h1.getName()
		emap[h1.getName()] = PriorityDay{h1, p1.Parent(), gHolidayPref[hName]}
		// entries = append(entries, HolidayEntry{h1, p1.Name, p1.Name})
	}
	for _, h2 := range p2.Holidays() {
		_, ok := emap[h2.getName()]
		hName2 := h2.getName()
		if !ok {
			emap[h2.getName()] = PriorityDay{h2, p2.Parent(), gHolidayPref[hName2]}
		} else {
			emap[h2.getName()] = PriorityDay{h2, BOTH, gHolidayPref[hName2]}
		}
	}
	for _, v := range emap {

		entries = append(entries, v)
	}
	slices.SortFunc(entries, func(a, b PriorityDay) int {
		aday := a.Day.CalculateDate(gStartDate.Year(), gTimezone)
		bday := b.Day.CalculateDate(gStartDate.Year(), gTimezone)
		return aday.Compare(bday)
	})
	return entries
}

// func (a Allotment) getDates(startDate time.Time, projectedEnd time.Time, holidays []PriorityDay, dates *[]time.Time, timezone *time.Location) {
// 	endDate := a.getEndTime(startDate)
// 	for startDate.Compare(endDate) == -1 {
// 		for _, holiday := range holidays {
// 			if holiday.Day.CalculateDate(startDate.Year(), timezone).Compare(startDate) == 0 {
// 				if holiday.isTurnOf(a.Parent) {
// 					holiday.Switch()
// 				}
// 			}
// 		}
// 		*dates = append(*dates, startDate)
//
// 	}
// }

//

type FormData struct {
	StartYear  int
	Timezone   string
	ParentTime int
	TimeUnit   []string
	Timezones  []string
	Holidays   []string
}

func makeFormData() FormData {
	now := time.Now()
	zones := []string{
		"Local",
		"EST",
		"MST",
		"PST",
		"AST",
		"CST",
		"AKST",
		"HST",
	}
	holidays := make([]string, 0)
	for i := 0; i < len(HOLIDAYS); i++ {
		holidays = append(holidays, HOLIDAYS[i].getName())
	}
	timeunits := []string{
		"days",
		"months",
	}
	return FormData{
		now.Year(),
		time.Local.String(),
		2,
		timeunits,
		zones,
		holidays,
	}

}

var ParentA_Allotment = MonthAllotment{
	Count:        2,
	PriorityDays: []Holiday{Christmas, ThanksgivingDay, Mothers, ChildBirthday},
	Custodian:    PARENT_A,
}

var ParentB_Allotment = MonthAllotment{
	Count:        2,
	PriorityDays: []Holiday{Christmas, ThanksgivingDay, FathersDay, ChildBirthday},
	Custodian:    PARENT_B,
}

var (
	gDebug          bool           = false
	gTimezone       *time.Location = time.Local
	gChildsBirthday time.Time
	gSchedule       Schedule
	gForYears       int = 1
	gStartDate      time.Time
	gEndDate        time.Time
	gParentAFirst   bool = true
	gParentACount   int
	gParentBCount   int
	gHolidayPref    map[string]Parent = make(map[string]Parent, 0)
)

type HDay struct {
	Date time.Time
	Name string
}

func main() {
	serverIP := "localhost:8080"
	staticServer := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticServer))
	http.HandleFunc("/", index)
	http.HandleFunc("/calendar", calendar)
	http.HandleFunc("/initSchedule", initSchedule)
	log.Printf("[Server]: Starting on %s\n", serverIP)
	http.ListenAndServe(serverIP, nil)
	// for _, d := range schedule.Mom {
	// 	fmt.Println(d.Format("2006-01-02"))
	// }
}
