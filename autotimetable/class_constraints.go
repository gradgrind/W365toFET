package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
)

// Each of these functions enables/disables a particular class constraint

func init() {
	cfmap[timetable.ClassMinLessonsPerDay] = func(
		instance *TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MinLessonsPerDay =
				instance.Global.TtData_0.Db.Classes[class].MinLessonsPerDay
		} else {
			instance.TtData.Db.Classes[class].MinLessonsPerDay = -1
		}
	}
	cfmap[timetable.ClassMaxLessonsPerDay] = func(
		instance *TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxLessonsPerDay =
				instance.Global.TtData_0.Db.Classes[class].MaxLessonsPerDay
		} else {
			instance.TtData.Db.Classes[class].MaxLessonsPerDay = -1
		}
	}
	cfmap[timetable.ClassMaxAfternoons] = func(
		instance *TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxAfternoons =
				instance.Global.TtData_0.Db.Classes[class].MaxAfternoons
		} else {
			instance.TtData.Db.Classes[class].MaxAfternoons = -1
		}
	}
	cfmap[timetable.ClassLunchBreak] = func(
		instance *TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].LunchBreak =
				instance.Global.TtData_0.Db.Classes[class].LunchBreak
		} else {
			instance.TtData.Db.Classes[class].LunchBreak = false
		}
	}
	cfmap[timetable.ClassForceFirstHour] = func(
		instance *TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].ForceFirstHour =
				instance.Global.TtData_0.Db.Classes[class].ForceFirstHour
		} else {
			instance.TtData.Db.Classes[class].ForceFirstHour = false
		}
	}
	cfmap[timetable.ClassMaxGapsPerDay] = func(
		instance *TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxGapsPerDay =
				instance.Global.TtData_0.Db.Classes[class].MaxGapsPerDay
		} else {
			instance.TtData.Db.Classes[class].MaxGapsPerDay = -1
		}
	}
	cfmap[timetable.ClassMaxGapsPerWeek] = func(
		instance *TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxGapsPerWeek =
				instance.Global.TtData_0.Db.Classes[class].MaxGapsPerWeek
		} else {
			instance.TtData.Db.Classes[class].MaxGapsPerWeek = -1
		}
	}

}

// Gather the active class constraints, according to type
func collect_class_constraints(
	classes []*base.Class, nDays int, nHours int,
) map[timetable.ConstraintType][]int {
	// constraint -> list of class indexes
	c_constraints := map[timetable.ConstraintType][]int{}
	for i, c := range classes {
		if c.MinLessonsPerDay != -1 {
			c_constraints[timetable.ClassMinLessonsPerDay] = append(
				c_constraints[timetable.ClassMinLessonsPerDay], i)
		}
		if c.MaxLessonsPerDay != -1 && c.MaxLessonsPerDay < nHours {
			c_constraints[timetable.ClassMaxLessonsPerDay] = append(
				c_constraints[timetable.ClassMaxLessonsPerDay], i)
		}
		if c.MaxAfternoons != -1 && c.MaxAfternoons < nDays {
			c_constraints[timetable.ClassMaxAfternoons] = append(
				c_constraints[timetable.ClassMaxAfternoons], i)
		}
		if c.ForceFirstHour {
			c_constraints[timetable.ClassForceFirstHour] = append(
				c_constraints[timetable.ClassForceFirstHour], i)
		}
		if c.LunchBreak {
			c_constraints[timetable.ClassLunchBreak] = append(
				c_constraints[timetable.ClassLunchBreak], i)
		}
		if c.MaxGapsPerDay != -1 {
			c_constraints[timetable.ClassMaxGapsPerDay] = append(
				c_constraints[timetable.ClassMaxGapsPerDay], i)
		}
		if c.MaxGapsPerWeek != -1 {
			c_constraints[timetable.ClassMaxGapsPerWeek] = append(
				c_constraints[timetable.ClassMaxGapsPerWeek], i)
		}
	}
	return c_constraints
}
