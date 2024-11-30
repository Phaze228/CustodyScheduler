package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	A_COLOR     = "#56494c"
	A_HOL_COLOR = "#56496c"
	B_COLOR     = "#c2d3cd"
	B_HOL_COLOR = "#c2d3ed"
)

type Day struct {
	Date      time.Time
	Day       int
	Parent    Parent
	Highlight bool
	Color     string
	Name      []string
}

type Week struct {
	Days []Day
}

type Month struct {
	MonthName string
	Year      int
	Weeks     []Week
}

func genWeeks(currentDate time.Time) []Week {
	var weeks []Week
	firstDay := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, gTimezone)
	firstWeek := Week{Days: make([]Day, 7)}
	curDay := firstDay
	for i := 0; i < int(firstDay.Weekday()); i++ {
		firstWeek.Days[i] = Day{Day: 0}
	}
	for i := int(firstDay.Weekday()); i < 7; i++ {
		firstWeek.Days[i] = gSchedule.Days[curDay.Format(YYYYMMDD)] //genDay(curDay, s)
		curDay = curDay.AddDate(0, 0, 1)
	}
	weeks = append(weeks, firstWeek)
	for curDay.Month() == currentDate.Month() {
		week := Week{Days: make([]Day, 7)}
		for i := 0; i < 7 && curDay.Month() == currentDate.Month(); i++ {
			week.Days[i] = gSchedule.Days[curDay.Format(YYYYMMDD)] //genDay(curDay, s)
			curDay = curDay.AddDate(0, 0, 1)
		}
		weeks = append(weeks, week)
	}
	return weeks
}

func containsDate(dates []time.Time, date time.Time) bool {
	for _, d := range dates {
		if d.Year() == date.Year() && d.Month() == date.Month() && d.Day() == date.Day() {
			return true
		}
	}
	return false
}

func mod(a, b int) int {
	return a % b
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.New("index.html")
	tmpl = tmpl.Funcs(template.FuncMap{"mod": mod})
	tmpl, err := tmpl.ParseFiles("templates/index.html")
	if err != nil {
		panic(err)
	}
	// tmpl := template.Must(template.ParseFiles("templates/index.html"))
	form := makeFormData()
	err = tmpl.Execute(w, form)
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func calendar(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Millisecond * 100)
	var months []Month

	curr := gStartDate
	for curr.Before(gEndDate) {
		cMonth := time.Date(curr.Year(), curr.Month(), 1, 0, 0, 0, 0, gTimezone)
		month := Month{
			MonthName: cMonth.Month().String(),
			Year:      cMonth.Year(),
			Weeks:     genWeeks(cMonth),
		}
		months = append(months, month)
		curr = cMonth.AddDate(0, 1, 0)
	}
	aParent := "Parent A"
	aColor := A_COLOR
	aCount := gParentACount
	bParent := "Parent B"
	bColor := B_COLOR
	bCount := gParentBCount
	if !gParentAFirst {
		aParent, bParent = bParent, aParent
		aColor, bColor = bColor, aColor
		aCount, bCount = bCount, aCount
	}

	tmpl := template.Must(template.ParseFiles("templates/calendar.html"))
	err := tmpl.Execute(w, struct {
		Months   []Month
		A_String string
		B_String string
		A_Color  string
		B_Color  string
		A_Count  int
		B_Count  int
	}{
		Months:   months,
		A_String: aParent,
		B_String: bParent,
		A_Color:  aColor,
		B_Color:  bColor,
		A_Count:  aCount,
		B_Count:  bCount,
	})
	if err != nil {
		log.Printf("[Server] Error: %v\n", err)
	}

}

func initSchedule(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error Parsing Form", http.StatusBadRequest)
		return
	}
	a, b := allotmentsFromFormData(r.Form)
	if !gParentAFirst {
		a, b = b, a
	}
	gSchedule = CreateSchedule(a, b, gStartDate, gStartDate.AddDate(gForYears, 0, 0))

}

func allotmentsFromFormData(f url.Values) (Allotment, Allotment) {
	gHolidayPref = make(map[string]Parent, 0)
	tz, err := time.LoadLocation(f.Get("timeZone"))
	if err != nil {
		log.Printf("Could not parse timezone")
	}
	year, err := time.Parse(YYYYMMDD, f.Get("startYear"))
	if err != nil {
		log.Printf("Could not parse start year")
	}
	foryears, err := strconv.Atoi(f.Get("forYears"))
	if err != nil {
		log.Printf("Error getting year end")
	}

	gTimezone = tz
	gForYears = foryears

	if f.Get("AFirst") == "parentA" {
		gParentAFirst = true
	} else {
		gParentAFirst = false

	}
	gStartDate = year.In(gTimezone)
	holidaysA := make([]Holiday, 0)
	holidaysB := make([]Holiday, 0)
	for _, h := range HOLIDAYS {
		name := h.getName()
		val := f.Get(name)
		vals := f[name]
		if len(vals) > 1 {
			val = "both"
		}
		// log.Println(name, val)
		parentOption := f.Get(name + "_Turn")
		if parentOption == "A" || parentOption == "" {
			gHolidayPref[name] = PARENT_A
		} else {
			gHolidayPref[name] = PARENT_B
		}
		switch val {
		case "A":
			holidaysA = append(holidaysA, h)
		case "B":
			holidaysB = append(holidaysB, h)
		case "both":
			holidaysA = append(holidaysA, h)
			holidaysB = append(holidaysB, h)
		default:
			continue
		}

	}

	timeA := f.Get("ParentATime")
	timeB := f.Get("ParentBTime")
	parsedTimeA, err := strconv.Atoi(timeA)
	if err != nil {
		log.Printf("Error parsing A's time: %v", err)
	}
	parsedTimeB, err := strconv.Atoi(timeB)
	if err != nil {
		log.Printf("Error parsing B's time: %v", err)
	}

	unitA := DAY
	unitB := DAY
	if f.Get("ParentAUnit") == "months" {
		unitA = MONTH
	}
	if f.Get("ParentBUnit") == "months" {
		unitB = MONTH
	}

	gChildsBirthday, err = time.Parse(YYYYMMDD, f.Get("childBirthday"))
	if err != nil {
		log.Printf("Error parsing birthday: %v\n", err)
	}
	ChildBirthday.Month = gChildsBirthday.Month()
	ChildBirthday.Day = gChildsBirthday.Day()
	for i := range HOLIDAYS {
		h := HOLIDAYS[i]
		if h.getName() == ChildBirthday.getName() {
			HOLIDAYS[i] = ChildBirthday
			break
		}
	}
	var parentA Allotment
	var parentB Allotment
	if unitA == DAY {
		parentA = &DayAllotment{
			Count:        parsedTimeA,
			PriorityDays: holidaysA,
			Custodian:    PARENT_A,
		}

	} else {
		parentA = &MonthAllotment{
			PriorityDays: holidaysA,
			Count:        parsedTimeA,
			Custodian:    PARENT_A,
		}

	}
	if unitB == DAY {

		parentB = &DayAllotment{
			Count:        parsedTimeB,
			PriorityDays: holidaysB,
			Custodian:    PARENT_B,
		}
	} else {
		parentB = &MonthAllotment{
			Count:        parsedTimeB,
			PriorityDays: holidaysB,
			Custodian:    PARENT_B,
		}

	}

	return parentA, parentB

}
