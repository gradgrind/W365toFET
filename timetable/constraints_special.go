package timetable

type SpecialConstraint struct {
	Item  int
	Value any
}

// Gather the active teacher constraints, according to type, adding them
// to the `TtData.HardConstraints` structure. The "NotAvailable" constraints
// are a special case.
func (tt_data *TtData) collect_teacher_constraints() {
	tt_shared_data := tt_data.SharedData
	ndays := tt_shared_data.NDays
	nhours := tt_shared_data.NHours
	tt_data.TeacherNotAvailable = make([][][]bool,
		len(tt_shared_data.Db.Teachers))
	for i, t := range tt_shared_data.Db.Teachers {
		// Every teacher has a blocked-slots matrix. They are ordered, so
		// they can be easily accessed.
		blocked_slots := make([][]bool, ndays)
		for d := range ndays {
			blocked_slots[d] = make([]bool, nhours)
		}
		for _, dh := range t.NotAvailable {
			if dh.Day < ndays && dh.Hour < nhours {
				blocked_slots[dh.Day][dh.Hour] = true
			}
		}
		tt_data.TeacherNotAvailable[i] = blocked_slots

		if t.MinLessonsPerDay > 0 {
			tt_data.HardConstraints[TeacherMinLessonsPerDay] = append(
				tt_data.HardConstraints[TeacherMinLessonsPerDay],
				SpecialConstraint{i, t.MinLessonsPerDay})
		}
		if t.MaxLessonsPerDay != -1 && t.MaxLessonsPerDay < nhours {
			tt_data.HardConstraints[TeacherMaxLessonsPerDay] = append(
				tt_data.HardConstraints[TeacherMaxLessonsPerDay],
				SpecialConstraint{i, t.MaxLessonsPerDay})
		}
		if t.MaxAfternoons != -1 && t.MaxAfternoons < ndays {
			tt_data.HardConstraints[TeacherMaxAfternoons] = append(
				tt_data.HardConstraints[TeacherMaxAfternoons],
				SpecialConstraint{i, t.MaxAfternoons})
		}
		if t.MaxDays != -1 && t.MaxDays < ndays {
			tt_data.HardConstraints[TeacherMaxDays] = append(
				tt_data.HardConstraints[TeacherMaxDays],
				SpecialConstraint{i, t.MaxDays})
		}
		if t.LunchBreak {
			tt_data.HardConstraints[TeacherLunchBreak] = append(
				tt_data.HardConstraints[TeacherLunchBreak],
				SpecialConstraint{i, t.LunchBreak})
		}
		if t.MaxGapsPerDay != -1 {
			tt_data.HardConstraints[TeacherMaxGapsPerDay] = append(
				tt_data.HardConstraints[TeacherMaxGapsPerDay],
				SpecialConstraint{i, t.MaxGapsPerDay})
		}
		if t.MaxGapsPerWeek != -1 {
			tt_data.HardConstraints[TeacherMaxGapsPerWeek] = append(
				tt_data.HardConstraints[TeacherMaxGapsPerWeek],
				SpecialConstraint{i, t.MaxGapsPerWeek})
		}
	}
}

// Gather the active class constraints, according to type, adding them
// to the `TtData.HardConstraints` structure. The "NotAvailable" constraints
// are a special case.
func (tt_data *TtData) collect_class_constraints() {
	tt_shared_data := tt_data.SharedData
	ndays := tt_shared_data.NDays
	nhours := tt_shared_data.NHours
	tt_data.ClassNotAvailable = make([][][]bool, len(tt_shared_data.Db.Classes))
	for i, c := range tt_shared_data.Db.Classes {
		// Every class has a blocked-slots matrix. They are ordered, so
		// they can be easily accessed.
		blocked_slots := make([][]bool, ndays)
		for d := range ndays {
			blocked_slots[d] = make([]bool, nhours)
		}
		for _, dh := range c.NotAvailable {
			if dh.Day < ndays && dh.Hour < nhours {
				blocked_slots[dh.Day][dh.Hour] = true
			}
		}
		tt_data.ClassNotAvailable[i] = blocked_slots

		if c.MinLessonsPerDay != -1 {
			tt_data.HardConstraints[ClassMinLessonsPerDay] = append(
				tt_data.HardConstraints[ClassMinLessonsPerDay],
				SpecialConstraint{i, c.MinLessonsPerDay})
		}
		if c.MaxLessonsPerDay != -1 && c.MaxLessonsPerDay < nhours {
			tt_data.HardConstraints[ClassMaxLessonsPerDay] = append(
				tt_data.HardConstraints[ClassMaxLessonsPerDay],
				SpecialConstraint{i, c.MaxLessonsPerDay})
		}
		if c.MaxAfternoons != -1 && c.MaxAfternoons < ndays {
			tt_data.HardConstraints[ClassMaxAfternoons] = append(
				tt_data.HardConstraints[ClassMaxAfternoons],
				SpecialConstraint{i, c.MaxAfternoons})
		}
		if c.ForceFirstHour {
			tt_data.HardConstraints[ClassForceFirstHour] = append(
				tt_data.HardConstraints[ClassForceFirstHour],
				SpecialConstraint{i, c.ForceFirstHour})
		}
		if c.LunchBreak {
			tt_data.HardConstraints[ClassLunchBreak] = append(
				tt_data.HardConstraints[ClassLunchBreak],
				SpecialConstraint{i, c.LunchBreak})
		}
		if c.MaxGapsPerDay != -1 {
			tt_data.HardConstraints[ClassMaxGapsPerDay] = append(
				tt_data.HardConstraints[ClassMaxGapsPerDay],
				SpecialConstraint{i, c.MaxGapsPerDay})
		}
		if c.MaxGapsPerWeek != -1 {
			tt_data.HardConstraints[ClassMaxGapsPerWeek] = append(
				tt_data.HardConstraints[ClassMaxGapsPerWeek],
				SpecialConstraint{i, c.MaxGapsPerWeek})
		}
	}
}

// Gather the active room constraints, according to type, adding them
// to the `TtData.HardConstraints` structure. The "NotAvailable" constraints
// are a special case.
func (tt_data *TtData) collect_room_constraints() {
	tt_shared_data := tt_data.SharedData
	ndays := tt_shared_data.NDays
	nhours := tt_shared_data.NHours
	tt_data.RoomNotAvailable = make([][][]bool, len(tt_shared_data.Db.Rooms))
	for i, c := range tt_shared_data.Db.Rooms {
		// Every room has a blocked-slots matrix. They are ordered, so
		// they can be easily accessed.
		blocked_slots := make([][]bool, ndays)
		for d := range ndays {
			blocked_slots[d] = make([]bool, nhours)
		}
		for _, dh := range c.NotAvailable {
			if dh.Day < ndays && dh.Hour < nhours {
				blocked_slots[dh.Day][dh.Hour] = true
			}
		}
		tt_data.RoomNotAvailable[i] = blocked_slots
	}

	// The room wishes are based on the activities
	for _, cinfo := range tt_shared_data.CourseInfoList {
		if len(cinfo.FixedRooms) != 0 || len(cinfo.RoomChoices) != 0 {
			for _, a := range cinfo.Activities {
				tt_data.HardConstraints[ActivityRooms] = append(
					tt_data.HardConstraints[ActivityRooms],
					SpecialConstraint{int(a), struct{}{}})
			}
		}
	}
}
