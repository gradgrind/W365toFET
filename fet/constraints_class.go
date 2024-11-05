package fet

import (
	"fmt"
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

All in all, I think the dummy lessons approach is probably better for the
classes. Note that these special lessons need to be generated for the atomic
groups if it is permissible for the various groups in a class to have their
break at different times.
*/

func addClassConstraints(fetinfo *fetInfo) {
	cminlpd := []minLessonsPerDay{}
	cmaxlpd := []maxLessonsPerDay{}
	cmaxgpd := []maxGapsPerDay{}
	cmaxgpw := []maxGapsPerWeek{}
	cmaxaft := []maxDaysinIntervalPerWeek{}
	cmaxls := []maxLateStarts{}
	ndays := len(fetinfo.days)
	nhours := len(fetinfo.hours)
	lblist := make([]string, ndays)
	mbhours := fetinfo.db.Info.MiddayBreak
	actlist := &fetinfo.fetdata.Activities_List

	for clix := 0; clix < len(fetinfo.db.Classes); clix++ {
		cl := &fetinfo.db.Classes[clix]

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

		n = int(cl.MaxGapsPerDay.(float64))
		if n >= 0 {
			cmaxgpd = append(cmaxgpd, maxGapsPerDay{
				Weight_Percentage: 100,
				Students:          cl.Tag,
				Max_Gaps:          n,
				Active:            true,
			})
		}

		n = int(cl.MaxGapsPerWeek.(float64))
		if n >= 0 {
			cmaxgpw = append(cmaxgpw, maxGapsPerWeek{
				Weight_Percentage: 100,
				Students:          cl.Tag,
				Max_Gaps:          n,
				Active:            true,
			})
		}

		i := fetinfo.db.Info.FirstAfternoonHour
		n = int(cl.MaxAfternoons.(float64))
		if n >= 0 && i > 0 {
			cmaxaft = append(cmaxaft, maxDaysinIntervalPerWeek{
				Weight_Percentage:   100,
				Students:            cl.Tag,
				Interval_Start_Hour: fetinfo.hours[i],
				Interval_End_Hour:   "", // end of day
				Max_Days_Per_Week:   n,
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

		if cl.LunchBreak {
			// Generate the constraint unless all days have a blocked lesson
			// at lunchtime.
			days := make([]bool, ndays)
			d := 0
			for _, ts := range cl.NotAvailable {
				if ts.Day < d {
					continue
				}
				if slices.Contains(mbhours, ts.Hour) {
					days[ts.Day] = true
					d = ts.Day + 1
				}
			}
			// Get days with lunch-breaks.
			lbdays := []string{}
			for d, nok := range days {
				if !nok {
					lb := lblist[d]
					if lb == "" {
						lb = fmt.Sprintf(LUNCH_BREAK_TAG, d)
						lblist[d] = lb
					}
					lbdays = append(lbdays, lb)
				}
			}
			if len(lbdays) != 0 {
				// Add dummy lessons for lunch-breaks.
				// Undivided classes have no atomic groups, but also they
				// need to be handled here.
				agtags := [][]string{}
				for _, ag := range fetinfo.atomicGroups[cl.Id] {
					agtags = append(agtags, []string{ag.Tag})
				}
				if len(agtags) == 0 {
					agtags = append(agtags, []string{cl.Tag})
				}
				for _, agtag := range agtags {
					for _, lb := range lbdays {
						aid := len(actlist.Activity) + 1
						actlist.Activity = append(
							actlist.Activity, fetActivity{
								Id:                aid,
								Teacher:           []string{},
								Subject:           lb,
								Students:          agtag,
								Active:            true,
								Total_Duration:    1,
								Duration:          1,
								Activity_Group_Id: 0,
							})
					}
				}
			}
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
	// Generate special subjects and time-constraints for the lunch-breaks.
	captss := []preferredSlots{}
	for d, lb := range lblist {
		// Dummy subjects for lunch breaks
		fetinfo.fetdata.Subjects_List.Subject = append(
			fetinfo.fetdata.Subjects_List.Subject, fetSubject{
				Name:      lb,
				Long_Name: LUNCH_BREAK_NAME,
			})
		// Constraint it to the lunch slots on the given day.
		dtag := fetinfo.days[d]
		ptlist := []preferredTime{}
		for _, h := range mbhours {
			ptlist = append(ptlist, preferredTime{
				Preferred_Day:  dtag,
				Preferred_Hour: fetinfo.hours[h],
			})
		}
		captss = append(captss, preferredSlots{
			Weight_Percentage:              100,
			Subject:                        lb,
			Number_of_Preferred_Time_Slots: 2,
			Preferred_Time_Slot:            ptlist,
			Active:                         true,
		})
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintActivitiesPreferredTimeSlots = captss
}
