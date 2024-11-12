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
for example when a class only has lessons earlier in the day.

Tests with the dummy lessons approach suggest that it is difficult to get the
number of these lessons and their placement on the correct days right.

This is an attempt with the max-hours-in-interval constraint.

*/

func addClassConstraints(fetinfo *fetInfo) {
	cminlpd := []minLessonsPerDay{}
	cmaxlpd := []maxLessonsPerDay{}
	cmaxgpd := []maxGapsPerDay{}
	cmaxgpw := []maxGapsPerWeek{}
	cmaxaft := []maxDaysinIntervalPerWeek{}
	cmaxls := []maxLateStarts{}
	clblist := []lunchBreak{}
	ndays := len(fetinfo.days)
	nhours := len(fetinfo.hours)

	for clix := 0; clix < len(fetinfo.db.Classes); clix++ {
		cl := &fetinfo.db.Classes[clix]
		if cl.Tag == "" {
			continue
		}

		n := int(cl.MinLessonsPerDay.(float64))
		if n >= 2 && n <= nhours {
			cminlpd = append(cminlpd, minLessonsPerDay{
				Weight_Percentage:   100,
				Students:            cl.Tag,
				Minimum_Hours_Daily: n,
				Allow_Empty_Days:    true,
				Active:              true,
			})
		}

		n = int(cl.MaxLessonsPerDay.(float64))
		if n >= 0 && n < nhours {
			cmaxlpd = append(cmaxlpd, maxLessonsPerDay{
				Weight_Percentage:   100,
				Students:            cl.Tag,
				Maximum_Hours_Daily: n,
				Active:              true,
			})
		}

		i := fetinfo.db.Info.FirstAfternoonHour
		maxpm := int(cl.MaxAfternoons.(float64))
		if maxpm >= 0 && i > 0 {
			cmaxaft = append(cmaxaft, maxDaysinIntervalPerWeek{
				Weight_Percentage:   100,
				Students:            cl.Tag,
				Interval_Start_Hour: fetinfo.hours[i],
				Interval_End_Hour:   "", // end of day
				Max_Days_Per_Week:   maxpm,
				Active:              true,
			})
		}

		if cl.ForceFirstHour {
			cmaxls = append(cmaxls, maxLateStarts{
				Weight_Percentage:             100,
				Max_Beginnings_At_Second_Hour: 0,
				Students:                      cl.Tag,
				Active:                        true,
			})
		}

		// The lunch-break constraint may require adjustment of these:
		mgpday := int(cl.MaxGapsPerDay.(float64))
		mgpweek := int(cl.MaxGapsPerWeek.(float64))
		if mgpweek < 0 {
			mgpweek = 0
		}

		if cl.LunchBreak {
			// Generate the constraint unless all days have a blocked lesson
			// at lunchtime.
			mbhours := fetinfo.db.Info.MiddayBreak
			lbdays := ndays
			d := 0
			for _, ts := range cl.NotAvailable {
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
				clblist = append(clblist, lunchBreak{
					Weight_Percentage:   100,
					Students:            cl.Tag,
					Interval_Start_Hour: fetinfo.hours[mbhours[0]],
					Interval_End_Hour:   fetinfo.hours[mbhours[0]+len(mbhours)],
					Maximum_Hours_Daily: len(mbhours) - 1,
					Active:              true,
				})
				//fmt.Printf("%s:: lbdays: %d maxpm: %d\n",
				//  cl.Tag, lbdays, maxpm)
				// Adjust gaps
				if maxpm < lbdays {
					lbdays = maxpm
				}
				if mgpday == 0 {
					mgpday = 1
				}
				if mgpweek >= 0 {
					mgpweek += lbdays
				}
			}
			//fmt.Printf("  --> %s::GapsPerDay: %d GapsPerWeek: %d\n",
			//	cl.Tag, mgpday, mgpweek)
		}
		if mgpday >= 0 {
			cmaxgpd = append(cmaxgpd, maxGapsPerDay{
				Weight_Percentage: 100,
				Students:          cl.Tag,
				Max_Gaps:          mgpday,
				Active:            true,
			})
		}

		if mgpweek >= 0 {
			cmaxgpw = append(cmaxgpw, maxGapsPerWeek{
				Weight_Percentage: 100,
				Students:          cl.Tag,
				Max_Gaps:          mgpweek,
				Active:            true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMinHoursDaily = cminlpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxHoursDaily = cmaxlpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxGapsPerDay = cmaxgpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxGapsPerWeek = cmaxgpw
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetIntervalMaxDaysPerWeek = cmaxaft
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetEarlyMaxBeginningsAtSecondHour = cmaxls
	// lunch breaks
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxHoursDailyInInterval = clblist
}
