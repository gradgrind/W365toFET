package fet

import "strconv"

func (fetinfo *fetInfo) handle_class_constraints() {
	natimes := []studentsNotAvailable{}
	cminlpd := []minLessonsPerDay{}
	cmaxlpd := []maxLessonsPerDay{}
	cmaxgpd := []maxGapsPerDay{}
	cmaxgpw := []maxGapsPerWeek{}
	cmaxaft := []maxDaysinIntervalPerWeek{}
	cmaxls := []maxLateStarts{}
	clblist := []lunchBreak{}
	tt_data := fetinfo.tt_data
	ndays := tt_data.NDays
	nhours := tt_data.NHours
	db := tt_data.Db

	for _, cl := range db.Classes {
		if cl.Tag == "" {
			continue
		}

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
}
