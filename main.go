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

const (
	PARENT_A Parent = iota
	PARENT_B
	BOTH
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
	return h.CurrTurn == parent
}

func CreateSchedule(pa, pb Allotments, startDate, endDate time.Time) Schedule {
	s := Schedule{
		Days: make(map[string]Day, 0),
	}
	if pa.Parent() != PARENT_A {
		log.Printf("parent b is first on schedule\n")
	}
	DASHES := strings.Repeat("-", 10)
	fmt.Printf("\n%s[Parental Debug]%s\n\n", DASHES, DASHES)
	gParentACount = 0
	gParentBCount = 0
	holidays := CreateEntries(pa, pb)
	pa.AppendDays(&startDate, &endDate, &pb, &holidays, s)
	fmt.Printf("\n%s [END] %s\n", DASHES, DASHES)
	return s
}

func CreateEntries(p1, p2 Allotments) []PriorityDay {
	entries := make([]PriorityDay, 0)
	emap := make(map[string]PriorityDay, 0)
	for _, h1 := range p1.PriorityDays {
		hName := h1.getName()
		emap[h1.getName()] = PriorityDay{h1, p1.Parent(), gHolidayPref[hName]}
	}
	for _, h2 := range p2.PriorityDays {
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

var ParentA_Allotment = Allotments{
	Type:         MONTH,
	Count:        2,
	PriorityDays: []Holiday{Christmas, ThanksgivingDay, MothersDay, ChildBirthday},
	Custodian:    PARENT_A,
}

var ParentB_Allotment = Allotments{
	Type:         MONTH,
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
	serverIP := "localhost:9090"
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
