package main

import (
	"fmt"
	"log"
	"slices"
	"time"
	// "log"
)

type Allotments struct {
	Type         TimeUnit
	Count        int
	PriorityDays []Holiday
	Custodian    Parent
	Template     string
}

func (a *Allotments) Allotment() (int, int, int) {
	switch a.Type {
	case MONTH:
		return 0, a.Count, 0
	case DAY:
		return 0, 0, a.Count
	default:
		return 0, 0, 0
	}
}

func (a *Allotments) Parent() Parent {
	return a.Custodian
}

func (a *Allotments) DaysForHoliday() int {
	if a.Type == MONTH {
		return DAYS_BEFORE_HOLIDAY_MONTHS
	} else {
		return DAYS_BEFORE_HOLIDAY_DAYS
	}
}

func (a *Allotments) Transition() int {
	if a.Type == MONTH {
		return TRANSITION_MONTHS
	} else {
		return TRANSITION_DAYS
	}
}

func (a *Allotments) AfterHoliday() int {
	if a.Type == MONTH {
		return DAYS_AFTER_HOLIDAY_MONTHS
	} else {
		return DAYS_AFTER_HOLIDAY_DAYS
	}
}

func (a *Allotments) AppendDays(start, yearEnd *time.Time, b *Allotments, holidays *[]PriorityDay, schedule Schedule) {
	count := 0
	var parent *Allotments
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
		fmt.Println("Current Parent: ", parent.Custodian.String())

		if left.Hours() > MAX_HOURS_EXCEEDED {
			log.Fatalf("MAX HOURS OF %d EXCEEDED---LeftOver: %.1f\n", MAX_HOURS_EXCEEDED, left.Hours())
		}

		end, holidayList := adjustForHolidays(*parent, start, &left, holidays)
		schedule.AppendDays(start, &end, parent, holidayList)
	}
	switchMonths := ((gEndDate.Sub(gStartDate).Hours() / 24) / 30) / float64(count)
	fmt.Printf("--- %d switches for %d days : Approximately %2.f per month ---\n", count, int(gEndDate.Sub(gStartDate).Hours()/24), switchMonths)

}

func adjustForHolidays(p Allotments, start *time.Time, timeLeft *time.Duration, holidays *[]PriorityDay) (time.Time, map[string][]string) {
	excessTime := *timeLeft
	normalEnd := start.AddDate(p.Allotment())
	maxTime := getMaxTime(p, *start)
	leftOverTime := excessTime
	if maxTime < leftOverTime {
		leftOverTime = leftOverTime - maxTime
		leftOverTime = maxTime
	} else {
		leftOverTime = excessTime
		excessTime = 0
	}

	end := start.Add(leftOverTime).AddDate(p.Allotment())
	months := make([]time.Month, 0)
	transition := end.AddDate(0, 0, p.DaysForHoliday())
	for s := *start; s.Before(transition); s = s.AddDate(0, 1, 0) {
		months = append(months, s.Month())
	}

	list := make(map[string][]string, 0)
	tEnd := end
	for i := 0; i < len(*holidays); i++ {
		h := (*holidays)[i]
		if !slices.Contains(months, h.Day.getMonth()) {
			continue
		}
		hDate := h.Day.CalculateDate(start.Year(), gTimezone)
		if end.Year() > start.Year() && hDate.Before(*start) {
			hDate = h.Day.CalculateDate(end.Year(), gTimezone)
		}
		list[hDate.Format(YYYYMMDD)] = append(list[hDate.Format(YYYYMMDD)], h.Day.getName())
		fmt.Println(list[hDate.Format(YYYYMMDD)])

		// SKIP PASSED HOLIDAYS
		if hDate.Before(*start) {
			continue
		}

		if h.isTurnOf(p.Custodian) {
			// fmt.Println(p.Custodian.String(), h.Day.getName(), h.MattersTo.String())
			// Check if parents turn and if time is before their transition period
			if hDate.After(end) && hDate.Before(transition) {
				newEnd := hDate.AddDate(0, 0, p.AfterHoliday())

				excessTime += end.Sub(newEnd)
				(*holidays)[i].Switch()

				tEnd = newEnd
				fmt.Printf("%s: Kept - %s\n", p.Custodian.String(), h.Day.getName())
			} else if hDate.Before(end) {
				(*holidays)[i].Switch()
			}

		} else {
			allotted := normalEnd.Sub(*start).Hours() / 24
			daysLost := normalEnd.Sub(hDate).Hours() / 24
			timePct := (allotted - daysLost) / allotted

			if hDate.After(*start) && hDate.Before(normalEnd) && timePct >= PARENTING_TIME_PERCENT {
				newEnd := hDate.AddDate(0, 0, -p.Transition())
				excessTime += end.Sub(newEnd)
				tEnd = newEnd
				// Break here due to parenting time switch on this holiday
				fmt.Printf(
					"[%s] - Lost %d days | %s\n\t [%s] -> [%s]\n",
					p.Custodian.String(), int(leftOverTime.Hours()/24), h.Day.getName(),
					end.Format(YYYYMMDD), newEnd.Format(YYYYMMDD),
				)
				break
			} else {
				fmt.Printf(
					"[%s] - Kept %s - %4d: Time Constraints: %.2f%% Parenting Time at holiday\n",
					p.Custodian.String(), h.Day.getName(), hDate.Year(), timePct*100,
				)
			}

		}
	}
	end = tEnd
	return end, list

}

func getMaxTime(p Allotments, s time.Time) time.Duration {
	e := s.AddDate(p.Allotment())
	duration := e.Sub(s)
	maxDuration := float64(duration) * MAX_TIME_BEFORE_SWITCH_PCT
	return time.Duration(int64(maxDuration))
}
