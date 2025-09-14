package fet

import (
	"W365toFET/timetable"
	"slices"
	"strconv"
)

func (fetinfo *fetInfo) handle_teacher_constraints() {
	natimes := []teacherNotAvailable{}
	tmaxdpw := []maxDaysT{}
	tminlpd := []minLessonsPerDayT{}
	tmaxlpd := []maxLessonsPerDayT{}
	tmaxgpd := []maxGapsPerDayT{}
	tmaxgpw := []maxGapsPerWeekT{}
	tmaxaft := []maxDaysinIntervalPerWeekT{}
	tlblist := []lunchBreakT{}
	tt_data := fetinfo.tt_data
	db := tt_data.Db
	ndays := tt_data.NDays
	nhours := tt_data.NHours
	cmap := tt_data.HardConstraints

	//TODO: If I don't need the teacher index or Db, it might be
	// better to put a pointer to the Teacher in Item ... and
	// maybe have separate types of SpecialConstraint fot teacher
	// and class.

	for _, c := range cmap[timetable.TeacherNotAvailable] {
		cn := c.(*timetable.SpecialConstraint)
		t := db.Teachers[cn.Item]
		matrix := cn.Value.([][]bool)
		// "Not available" times
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
			natimes = append(natimes,
				teacherNotAvailable{
					Weight_Percentage:             100,
					Teacher:                       t.Tag,
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
				})
		}

		for _, c := range cmap[timetable.TeacherMaxDays] {
			cn := c.(*timetable.SpecialConstraint)
			t := db.Teachers[cn.Item]
			n := cn.Value.(int)
			if n >= 0 && n < ndays {
				tmaxdpw = append(tmaxdpw, maxDaysT{
					Weight_Percentage: 100,
					Teacher:           t.Tag,
					Max_Days_Per_Week: n,
					Active:            true,
				})
			}
		}
		for _, c := range cmap[timetable.TeacherMinLessonsPerDay] {
			cn := c.(*timetable.SpecialConstraint)
			t := db.Teachers[cn.Item]
			n := cn.Value.(int)
			if n >= 2 && n <= nhours {
				tminlpd = append(tminlpd, minLessonsPerDayT{
					Weight_Percentage:   100,
					Teacher:             t.Tag,
					Minimum_Hours_Daily: n,
					Allow_Empty_Days:    true,
					Active:              true,
				})
			}
		}
		for _, c := range cmap[timetable.TeacherMaxLessonsPerDay] {
			cn := c.(*timetable.SpecialConstraint)
			t := db.Teachers[cn.Item]
			n := cn.Value.(int)
			if n >= 0 && n < nhours {
				tmaxlpd = append(tmaxlpd, maxLessonsPerDayT{
					Weight_Percentage:   100,
					Teacher:             t.Tag,
					Maximum_Hours_Daily: n,
					Active:              true,
				})
			}
		}
		for _, c := range cmap[timetable.TeacherMaxAfternoons] {
			cn := c.(*timetable.SpecialConstraint)
			t := db.Teachers[cn.Item]
			i := db.Info.FirstAfternoonHour
			maxpm := t.MaxAfternoons
			if maxpm >= 0 && i > 0 {
				tmaxaft = append(tmaxaft, maxDaysinIntervalPerWeekT{
					Weight_Percentage:   100,
					Teacher:             t.Tag,
					Interval_Start_Hour: strconv.Itoa(i),
					Interval_End_Hour:   "", // end of day
					Max_Days_Per_Week:   maxpm,
					Active:              true,
				})
			}
		}

		//TODO...

		// The lunch-break constraint may require adjustment of these:
		//   t.MaxGapsPerDay
		//   t.MaxGapsPerWeek
		// So keep a record of the lunch-break constraints. Actually, that
		// could be done by having a value for every teacher, ordered, like
		// the blocked slots.

		if mbhours := db.Info.MiddayBreak; len(mbhours) != 0 {
			not_availables := cmap[timetable.TeacherNotAvailable]
			for _, c := range cmap[timetable.TeacherLunchBreak] {
				cn := c.(*timetable.SpecialConstraint)
				t := db.Teachers[cn.Item]
				if cn.Value.(bool) {
					nat := not_availables[cn.Item].(*timetable.SpecialConstraint).
						Value.([][]bool)

					// Generate the constraint unless all days have a blocked lesson
					// at lunchtime.
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
							Interval_Start_Hour: strconv.Itoa(mbhours[0]),
							Interval_End_Hour:   strconv.Itoa(mbhours[0] + len(mbhours)),
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
						if mgpweek >= 0 {
							mgpweek += lbdays
						}
					}
				}
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
		ConstraintTeacherNotAvailableTimes = natimes
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
