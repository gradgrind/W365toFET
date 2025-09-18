package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
)

// Gather the active teacher constraints, according to type
func collect_teacher_constraints(
	teachers []*base.Teacher, nDays int, nHours int,
) map[timetable.ConstraintType][]int {
	// constraint -> list of teacher indexes
	t_constraints := map[timetable.ConstraintType][]int{}
	for i, t := range teachers {
		if t.MinLessonsPerDay > 0 {
			t_constraints[timetable.TeacherMinLessonsPerDay] = append(
				t_constraints[timetable.TeacherMinLessonsPerDay], i)
		}
		if t.MaxLessonsPerDay != -1 && t.MaxLessonsPerDay < nHours {
			t_constraints[timetable.TeacherMaxLessonsPerDay] = append(
				t_constraints[timetable.TeacherMaxLessonsPerDay], i)
		}
		if t.MaxAfternoons != -1 && t.MaxAfternoons < nDays {
			t_constraints[timetable.TeacherMaxAfternoons] = append(
				t_constraints[timetable.TeacherMaxAfternoons], i)
		}
		if t.MaxDays != -1 && t.MaxDays < nDays {
			t_constraints[timetable.TeacherMaxDays] = append(
				t_constraints[timetable.TeacherMaxDays], i)
		}
		if t.LunchBreak {
			t_constraints[timetable.TeacherLunchBreak] = append(
				t_constraints[timetable.TeacherLunchBreak], i)
		}
		if t.MaxGapsPerDay != -1 {
			t_constraints[timetable.TeacherMaxGapsPerDay] = append(
				t_constraints[timetable.TeacherMaxGapsPerDay], i)
		}
		if t.MaxGapsPerWeek != -1 {
			t_constraints[timetable.TeacherMaxGapsPerWeek] = append(
				t_constraints[timetable.TeacherMaxGapsPerWeek], i)
		}
	}
	return t_constraints
}
