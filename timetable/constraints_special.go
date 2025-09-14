package timetable

type SpecialConstraint struct {
	Item  int
	Value any
}

// Gather the active teacher constraints, according to type, adding them
// to the `TtData.HardConstraints` structure.
func (tt_data *TtData) collect_teacher_constraints() {
	for i, t := range tt_data.Db.Teachers {
		if t.MinLessonsPerDay > 0 {
			tt_data.HardConstraints[TeacherMinLessonsPerDay] = append(
				tt_data.HardConstraints[TeacherMinLessonsPerDay],
				SpecialConstraint{i, t.MinLessonsPerDay})
		}
		if t.MaxLessonsPerDay != -1 && t.MaxLessonsPerDay < tt_data.NHours {
			tt_data.HardConstraints[TeacherMaxLessonsPerDay] = append(
				tt_data.HardConstraints[TeacherMaxLessonsPerDay],
				SpecialConstraint{i, t.MaxLessonsPerDay})
		}
		if t.MaxAfternoons != -1 && t.MaxAfternoons < tt_data.NDays {
			tt_data.HardConstraints[TeacherMaxAfternoons] = append(
				tt_data.HardConstraints[TeacherMaxAfternoons],
				SpecialConstraint{i, t.MaxAfternoons})
		}
		if t.MaxDays != -1 && t.MaxDays < tt_data.NDays {
			tt_data.HardConstraints[TeacherMaxDays] = append(
				tt_data.HardConstraints[TeacherMaxDays],
				SpecialConstraint{i, t.MaxDays})
		}
		if t.LunchBreak {
			tt_data.HardConstraints[TeacherLunchBreak] = append(
				tt_data.HardConstraints[TeacherLunchBreak],
				SpecialConstraint{i, TeacherLunchBreak})
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
// to the `TtData.HardConstraints` structure.
func (tt_data *TtData) collect_class_constraints() {
	for i, c := range tt_data.Db.Classes {
		if c.MinLessonsPerDay != -1 {
			tt_data.HardConstraints[ClassMinLessonsPerDay] = append(
				tt_data.HardConstraints[ClassMinLessonsPerDay],
				SpecialConstraint{i, c.MinLessonsPerDay})
		}
		if c.MaxLessonsPerDay != -1 && c.MaxLessonsPerDay < tt_data.NHours {
			tt_data.HardConstraints[ClassMaxLessonsPerDay] = append(
				tt_data.HardConstraints[ClassMaxLessonsPerDay],
				SpecialConstraint{i, c.MaxLessonsPerDay})
		}
		if c.MaxAfternoons != -1 && c.MaxAfternoons < tt_data.NDays {
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
