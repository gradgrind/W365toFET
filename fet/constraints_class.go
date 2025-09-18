package fet

import (
	"W365toFET/timetable"
	"strconv"
)

/* Lunch-breaks

Lunch-breaks can be done using max-hours-in-interval constraint, but that
makes specification of max-gaps more difficult (because the lunch breaks
count as gaps).

The alternative is to add dummy lessons, clamped to the midday-break hours,
on the days where none of the midday-break hours are blocked. However, this
can also cause problems with gaps – the dummy lesson can itself create gaps,
for example when a class only has lessons earlier in the day.

Tests with the dummy lessons approach suggest that it is difficult to get the
number of these lessons and their placement on the correct days right.

The code here uses max-hours-in-interval constraints and tries to compensate
for the gaps that are thus created by adjusting the max-gaps constraints.

*/

// ------------------------------------------------------------------------

func (fetinfo *fetInfo) handle_class_constraints() {
	tt_data := fetinfo.tt_data
	shared_data := tt_data.SharedData
	ndays := shared_data.NDays
	nhours := shared_data.NHours
	db := shared_data.Db
	cmap := tt_data.HardConstraints

	natimes := []studentsNotAvailable{}
	for cix, matrix := range tt_data.ClassNotAvailable {
		nats := []notAvailableTime{}
		for d, hlist := range matrix {
			for h, blocked := range hlist {
				if blocked {
					nats = append(nats,
						notAvailableTime{
							Day:  strconv.Itoa(d),
							Hour: strconv.Itoa(h)})
				}
			}
		}
		if len(nats) > 0 {
			cl := db.Classes[cix]
			natimes = append(natimes,
				studentsNotAvailable{
					Weight_Percentage:             100,
					Students:                      cl.Tag,
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
				})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetNotAvailableTimes = natimes

	cminlpd := []minLessonsPerDay{}
	for _, c := range cmap[timetable.ClassMinLessonsPerDay] {
		cn := c.(timetable.SpecialConstraint)
		n := cn.Value.(int)
		if n >= 2 && n <= nhours {
			cl := db.Classes[cn.Item]
			cminlpd = append(cminlpd, minLessonsPerDay{
				Weight_Percentage:   100,
				Students:            cl.Tag,
				Minimum_Hours_Daily: n,
				Allow_Empty_Days:    true,
				Active:              true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMinHoursDaily = cminlpd

	cmaxlpd := []maxLessonsPerDay{}
	for _, c := range cmap[timetable.ClassMaxLessonsPerDay] {
		cn := c.(timetable.SpecialConstraint)
		n := cn.Value.(int)
		if n >= 0 && n < nhours {
			cl := db.Classes[cn.Item]
			cmaxlpd = append(cmaxlpd, maxLessonsPerDay{
				Weight_Percentage:   100,
				Students:            cl.Tag,
				Maximum_Hours_Daily: n,
				Active:              true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxHoursDaily = cmaxlpd

	cmaxaft := []maxDaysinIntervalPerWeek{}
	// Gather the max afternoons constraints as they may influence the
	// max-gaps constraints.
	//    class index -> max number of afternoons
	pmmap := map[int]int{}
	h0 := db.Info.FirstAfternoonHour
	if h0 > 0 {
		for _, c := range cmap[timetable.ClassMaxAfternoons] {
			cn := c.(timetable.SpecialConstraint)
			n := cn.Value.(int)
			cl := db.Classes[cn.Item]
			cmaxaft = append(cmaxaft, maxDaysinIntervalPerWeek{
				Weight_Percentage:   100,
				Students:            cl.Tag,
				Interval_Start_Hour: strconv.Itoa(h0),
				Interval_End_Hour:   "", // end of day
				Max_Days_Per_Week:   n,
				Active:              true,
			})
			pmmap[cn.Item] = n
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetIntervalMaxDaysPerWeek = cmaxaft

	cmaxls := []maxLateStarts{}
	for _, c := range cmap[timetable.ClassForceFirstHour] {
		cn := c.(timetable.SpecialConstraint)
		if cn.Value.(bool) {
			cl := db.Classes[cn.Item]
			cmaxls = append(cmaxls, maxLateStarts{
				Weight_Percentage:             100,
				Max_Beginnings_At_Second_Hour: 0,
				Students:                      cl.Tag,
				Active:                        true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetEarlyMaxBeginningsAtSecondHour = cmaxls

	clblist := []lunchBreak{}
	// Gather the lunch-break constraints as they may influence the
	// max-gaps constraints.
	//    class index -> number of days with lunch break
	lbmap := map[int]int{}
	if mbhours := db.Info.MiddayBreak; len(mbhours) != 0 {
		for _, c := range cmap[timetable.ClassLunchBreak] {
			cn := c.(timetable.SpecialConstraint)
			if cn.Value.(bool) {
				// Generate the constraint unless all days have a blocked
				// lesson at lunchtime.
				nat := tt_data.ClassNotAvailable[cn.Item]
				lbdays := ndays
				for d := range ndays {
					for _, h := range mbhours {
						if nat[d][h] {
							lbdays--
							break
						}
					}
				}
				if lbdays != 0 {
					// Add a lunch-break constraint.
					cl := db.Classes[cn.Item]
					clblist = append(clblist, lunchBreak{
						Weight_Percentage:   100,
						Students:            cl.Tag,
						Interval_Start_Hour: strconv.Itoa(mbhours[0]),
						Interval_End_Hour:   strconv.Itoa(mbhours[0] + len(mbhours)),
						Maximum_Hours_Daily: len(mbhours) - 1,
						Active:              true,
					})
					lbmap[cn.Item] = lbdays
				}
			}
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxHoursDailyInInterval = clblist

	cmaxgpd := []maxGapsPerDay{}
	for _, c := range cmap[timetable.ClassMaxGapsPerDay] {
		cn := c.(timetable.SpecialConstraint)
		n := cn.Value.(int)
		// Ensure that a gap is allowed if there are lunch breaks.
		if n == 0 {
			_, ok := lbmap[cn.Item]
			if ok {
				// lbdays > 0
				maxpm, ok := pmmap[cn.Item]
				if !ok || maxpm != 0 {
					n = 1
				}
			}
		}
		if n >= 0 {
			cl := db.Classes[cn.Item]
			cmaxgpd = append(cmaxgpd, maxGapsPerDay{
				Weight_Percentage: 100,
				Students:          cl.Tag,
				Max_Gaps:          n,
				Active:            true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxGapsPerDay = cmaxgpd

	cmaxgpw := []maxGapsPerWeek{}
	for _, c := range cmap[timetable.ClassMaxGapsPerWeek] {
		cn := c.(timetable.SpecialConstraint)
		n := cn.Value.(int)
		if n >= 0 {
			// Adjust to accommodate lunch breaks
			lbdays, ok := lbmap[cn.Item]
			if ok {
				// lbdays > 0
				maxpm, ok := pmmap[cn.Item]
				if ok && maxpm < lbdays {
					lbdays = maxpm
				}
				n += lbdays
			}
			cl := db.Classes[cn.Item]
			cmaxgpw = append(cmaxgpw, maxGapsPerWeek{
				Weight_Percentage: 100,
				Students:          cl.Tag,
				Max_Gaps:          n,
				Active:            true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxGapsPerWeek = cmaxgpw
}
