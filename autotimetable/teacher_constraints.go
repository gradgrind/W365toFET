package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
)

// Each of these functions enables/disables a particular teacher constraint

func init() {
	cfmap[timetable.TeacherMinLessonsPerDay] = func(
		instance *TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MinLessonsPerDay =
				instance.Global.TtData_0.Db.Teachers[teacher].MinLessonsPerDay
		} else {
			instance.TtData.Db.Teachers[teacher].MinLessonsPerDay = -1
		}
	}
	cfmap[timetable.TeacherMaxLessonsPerDay] = func(
		instance *TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MaxLessonsPerDay =
				instance.Global.TtData_0.Db.Teachers[teacher].MaxLessonsPerDay
		} else {
			instance.TtData.Db.Teachers[teacher].MaxLessonsPerDay = -1
		}
	}
	cfmap[timetable.TeacherMaxAfternoons] = func(
		instance *TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MaxAfternoons =
				instance.Global.TtData_0.Db.Teachers[teacher].MaxAfternoons
		} else {
			instance.TtData.Db.Teachers[teacher].MaxAfternoons = -1
		}
	}
	cfmap[timetable.TeacherMaxDays] = func(
		instance *TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MaxDays =
				instance.Global.TtData_0.Db.Teachers[teacher].MaxDays
		} else {
			instance.TtData.Db.Teachers[teacher].MaxDays = -1
		}
	}
	cfmap[timetable.TeacherLunchBreak] = func(
		instance *TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].LunchBreak =
				instance.Global.TtData_0.Db.Teachers[teacher].LunchBreak
		} else {
			instance.TtData.Db.Teachers[teacher].LunchBreak = false
		}
	}
	cfmap[timetable.TeacherMaxGapsPerDay] = func(
		instance *TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MaxGapsPerDay =
				instance.Global.TtData_0.Db.Teachers[teacher].MaxGapsPerDay
		} else {
			instance.TtData.Db.Teachers[teacher].MaxGapsPerDay = -1
		}
	}
	cfmap[timetable.TeacherMaxGapsPerWeek] = func(
		instance *TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MaxGapsPerWeek =
				instance.Global.TtData_0.Db.Teachers[teacher].MaxGapsPerWeek
		} else {
			instance.TtData.Db.Teachers[teacher].MaxGapsPerWeek = -1
		}
	}
}

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
