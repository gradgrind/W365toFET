package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
)

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
