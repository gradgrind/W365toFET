package fet

import (
	"slices"
)

/* Lunch-breaks

Lunch-breaks can be done using max-hours-in-interval constraint, but that
makes specification of max-gaps more difficult (becuase the lunch breaks
count as gaps).

The alternative is to add dummy lessons, clamped to the midday-break hours,
on the days where none of the midday-break hours are blocked. However, this
can also cause problems with gaps – the dummy lesson can itself create gaps,
for example when a teacher's lessons are earlier in the day.

All in all, I think the max-hours-in-interval constraint is probably better
for the teachers. If there is a maximum-gaps constraint, the user may need
to adjust it to take the lunch-breaks into acccount.
*/

func addTeacherConstraints(fetinfo *fetInfo) {
	tmaxdpw := []maxDaysT{}
	tminlpd := []minLessonsPerDayT{}
	tmaxlpd := []maxLessonsPerDayT{}
	tmaxgpd := []maxGapsPerDayT{}
	tmaxgpw := []maxGapsPerWeekT{}
	tmaxaft := []maxDaysinIntervalPerWeekT{}
	tlblist := []lunchBreakT{}
	ndays := len(fetinfo.days)
	nhours := len(fetinfo.hours)

	for tix := 0; tix < len(fetinfo.db.Teachers); tix++ {
		t := &fetinfo.db.Teachers[tix]
		n := int(t.MaxDays.(float64))
		if n >= 0 && n < ndays {
			tmaxdpw = append(tmaxdpw, maxDaysT{
				Weight_Percentage: 100,
				Teacher:           t.Tag,
				Max_Days_Per_Week: n,
				Active:            true,
			})
		}

		n = int(t.MinLessonsPerDay.(float64))
		if n >= 2 && n <= nhours {
			tminlpd = append(tminlpd, minLessonsPerDayT{
				Weight_Percentage:   100,
				Teacher:             t.Tag,
				Minimum_Hours_Daily: n,
				Allow_Empty_Days:    true,
				Active:              true,
			})
		}

		n = int(t.MaxLessonsPerDay.(float64))
		if n >= 0 && n < nhours {
			tmaxlpd = append(tmaxlpd, maxLessonsPerDayT{
				Weight_Percentage:   100,
				Teacher:             t.Tag,
				Maximum_Hours_Daily: n,
				Active:              true,
			})
		}

		i := fetinfo.db.Info.FirstAfternoonHour
		maxpm := int(t.MaxAfternoons.(float64))
		if maxpm >= 0 && i > 0 {
			tmaxaft = append(tmaxaft, maxDaysinIntervalPerWeekT{
				Weight_Percentage:   100,
				Teacher:             t.Tag,
				Interval_Start_Hour: fetinfo.hours[i],
				Interval_End_Hour:   "", // end of day
				Max_Days_Per_Week:   maxpm,
				Active:              true,
			})
		}

		// The lunch-break constraint may require adjustment of these:
		mgpday := int(t.MaxGapsPerDay.(float64))
		mgpweek := int(t.MaxGapsPerWeek.(float64))

		if t.LunchBreak {
			// Generate the constraint unless all days have a blocked lesson
			// at lunchtime.
			mbhours := fetinfo.db.Info.MiddayBreak
			lbdays := ndays
			d := 0
			for _, ts := range t.NotAvailable {
				if ts.Day < d {
					continue
				}
				if slices.Contains(mbhours, ts.Hour) {
					lbdays--
					d = ts.Day + 1
				}
			}
			if lbdays != 0 {
				// Add a lunch-break constraint.
				tlblist = append(tlblist, lunchBreakT{
					Weight_Percentage:   100,
					Teacher:             t.Tag,
					Interval_Start_Hour: fetinfo.hours[mbhours[0]],
					Interval_End_Hour:   fetinfo.hours[mbhours[0]+len(mbhours)],
					Maximum_Hours_Daily: len(mbhours) - 1,
					Active:              true,
				})
				// Adjust gaps
				if maxpm < lbdays {
					lbdays = maxpm
				}
				if mgpday == 0 {
					mgpday = 1
				}
				mgpweek += lbdays
			}
		}

		if mgpday >= 0 {
			tmaxgpd = append(tmaxgpd, maxGapsPerDayT{
				Weight_Percentage: 100,
				Teacher:           t.Tag,
				Max_Gaps:          mgpday,
				Active:            true,
			})
		}

		if mgpweek >= 0 {
			tmaxgpw = append(tmaxgpw, maxGapsPerWeekT{
				Weight_Percentage: 100,
				Teacher:           t.Tag,
				Max_Gaps:          mgpweek,
				Active:            true,
			})
		}

	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxDaysPerWeek = tmaxdpw
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMinHoursDaily = tminlpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxHoursDaily = tmaxlpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxGapsPerDay = tmaxgpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxGapsPerWeek = tmaxgpw
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherIntervalMaxDaysPerWeek = tmaxaft
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxHoursDailyInInterval = tlblist
}
