package fet

import (
	"W365toFET/timetable"
	"strconv"
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
for the teachers. The code here tries to compensate for the gaps that are
thus created by adjusting the max-gaps constraints.

*/

// ------------------------------------------------------------------------

func (fetinfo *fetInfo) handle_teacher_constraints() {
	tt_data := fetinfo.tt_data
	db := tt_data.Db
	ndays := tt_data.NDays
	nhours := tt_data.NHours
	cmap := tt_data.HardConstraints

	natimes := []teacherNotAvailable{}
	for tix, matrix := range tt_data.TeacherNotAvailable {
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
			t := db.Teachers[tix]
			natimes = append(natimes,
				teacherNotAvailable{
					Weight_Percentage:             100,
					Teacher:                       t.Tag,
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
				})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherNotAvailableTimes = natimes

	tmaxdpw := []maxDaysT{}
	for _, c := range cmap[timetable.TeacherMaxDays] {
		cn := c.(timetable.SpecialConstraint)
		n := cn.Value.(int)
		if n >= 0 && n < ndays {
			t := db.Teachers[cn.Item]
			tmaxdpw = append(tmaxdpw, maxDaysT{
				Weight_Percentage: 100,
				Teacher:           t.Tag,
				Max_Days_Per_Week: n,
				Active:            true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxDaysPerWeek = tmaxdpw

	tminlpd := []minLessonsPerDayT{}
	for _, c := range cmap[timetable.TeacherMinLessonsPerDay] {
		cn := c.(timetable.SpecialConstraint)
		n := cn.Value.(int)
		if n >= 2 && n <= nhours {
			t := db.Teachers[cn.Item]
			tminlpd = append(tminlpd, minLessonsPerDayT{
				Weight_Percentage:   100,
				Teacher:             t.Tag,
				Minimum_Hours_Daily: n,
				Allow_Empty_Days:    true,
				Active:              true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMinHoursDaily = tminlpd

	tmaxlpd := []maxLessonsPerDayT{}
	for _, c := range cmap[timetable.TeacherMaxLessonsPerDay] {
		cn := c.(timetable.SpecialConstraint)
		n := cn.Value.(int)
		if n >= 0 && n < nhours {
			t := db.Teachers[cn.Item]
			tmaxlpd = append(tmaxlpd, maxLessonsPerDayT{
				Weight_Percentage:   100,
				Teacher:             t.Tag,
				Maximum_Hours_Daily: n,
				Active:              true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxHoursDaily = tmaxlpd

	tmaxaft := []maxDaysinIntervalPerWeekT{}
	// Gather the max afternoons constraints as they may influence the
	// max-gaps constraints.
	//    teacher index -> max number of afternoons
	pmmap := map[int]int{}
	h0 := db.Info.FirstAfternoonHour
	if h0 > 0 {
		for _, c := range cmap[timetable.TeacherMaxAfternoons] {
			cn := c.(timetable.SpecialConstraint)
			n := cn.Value.(int)
			t := db.Teachers[cn.Item]
			tmaxaft = append(tmaxaft, maxDaysinIntervalPerWeekT{
				Weight_Percentage:   100,
				Teacher:             t.Tag,
				Interval_Start_Hour: strconv.Itoa(h0),
				Interval_End_Hour:   "", // end of day
				Max_Days_Per_Week:   n,
				Active:              true,
			})
			pmmap[cn.Item] = n
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherIntervalMaxDaysPerWeek = tmaxaft

	tlblist := []lunchBreakT{}
	// Gather the lunch-break constraints as they may influence the
	// max-gaps constraints.
	//    teacher index -> number of days with lunch break
	lbmap := map[int]int{}
	if mbhours := db.Info.MiddayBreak; len(mbhours) != 0 {
		for _, c := range cmap[timetable.TeacherLunchBreak] {
			cn := c.(timetable.SpecialConstraint)
			if cn.Value.(bool) {
				// Generate the constraint unless all days have a blocked
				// lesson at lunchtime.
				nat := tt_data.TeacherNotAvailable[cn.Item]
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
					t := db.Teachers[cn.Item]
					tlblist = append(tlblist, lunchBreakT{
						Weight_Percentage:   100,
						Teacher:             t.Tag,
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
		ConstraintTeacherMaxHoursDailyInInterval = tlblist

	tmaxgpd := []maxGapsPerDayT{}
	for _, c := range cmap[timetable.TeacherMaxGapsPerDay] {
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
			t := db.Teachers[cn.Item]
			tmaxgpd = append(tmaxgpd, maxGapsPerDayT{
				Weight_Percentage: 100,
				Teacher:           t.Tag,
				Max_Gaps:          n,
				Active:            true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxGapsPerDay = tmaxgpd

	tmaxgpw := []maxGapsPerWeekT{}
	for _, c := range cmap[timetable.TeacherMaxGapsPerWeek] {
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
			t := db.Teachers[cn.Item]
			tmaxgpw = append(tmaxgpw, maxGapsPerWeekT{
				Weight_Percentage: 100,
				Teacher:           t.Tag,
				Max_Gaps:          n,
				Active:            true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxGapsPerWeek = tmaxgpw
}
