package main

import "time"

type Schedule struct {
	Days map[string]Day
}

func (s *Schedule) AppendDays(start, end *time.Time, a *Allotments, holidayList map[string][]string) {
	var color string
	highlight := false
	e := *end

	for start.Before(e) {
		dayString := start.Format(YYYYMMDD)
		if a.Custodian == PARENT_A {
			gParentACount++
			color = A_COLOR
		} else {
			gParentBCount++
			color = B_COLOR
		}
		if len(holidayList[dayString]) > 0 {
			highlight = true
			if a.Custodian == PARENT_A {
				color = A_HOL_COLOR
			} else {
				color = B_HOL_COLOR
			}
		}
		s.Days[dayString] = Day{
			Date:      *start,
			Day:       start.Day(),
			Parent:    a.Custodian,
			Highlight: highlight,
			Color:     color,
			Name:      holidayList[dayString],
		}
		*start = start.AddDate(0, 0, 1)
		gEndDate = *start

	}
}
