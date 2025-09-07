package autotimetable

import (
	"W365toFET/timetable"
)

// Each of these functions enables/disables a particular teacher constraint

func init() {
	cfmap[TMinLessonsPerDay] = func(
		instance *timetable.TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MinLessonsPerDay =
				instance.TtData_0.Db.Teachers[teacher].MinLessonsPerDay
		} else {
			instance.TtData.Db.Teachers[teacher].MinLessonsPerDay = -1
		}
	}
	cfmap[TMaxLessonsPerDay] = func(
		instance *timetable.TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MaxLessonsPerDay =
				instance.TtData_0.Db.Teachers[teacher].MaxLessonsPerDay
		} else {
			instance.TtData.Db.Teachers[teacher].MaxLessonsPerDay = -1
		}
	}
	cfmap[TMaxAfternoons] = func(
		instance *timetable.TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MaxAfternoons =
				instance.TtData_0.Db.Teachers[teacher].MaxAfternoons
		} else {
			instance.TtData.Db.Teachers[teacher].MaxAfternoons = -1
		}
	}
	cfmap[TMaxDays] = func(
		instance *timetable.TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MaxDays =
				instance.TtData_0.Db.Teachers[teacher].MaxDays
		} else {
			instance.TtData.Db.Teachers[teacher].MaxDays = -1
		}
	}
	cfmap[TLunchBreak] = func(
		instance *timetable.TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].LunchBreak =
				instance.TtData_0.Db.Teachers[teacher].LunchBreak
		} else {
			instance.TtData.Db.Teachers[teacher].LunchBreak = false
		}
	}
	cfmap[TMaxGapsPerDay] = func(
		instance *timetable.TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MaxGapsPerDay =
				instance.TtData_0.Db.Teachers[teacher].MaxGapsPerDay
		} else {
			instance.TtData.Db.Teachers[teacher].MaxGapsPerDay = -1
		}
	}
	cfmap[TMaxGapsPerWeek] = func(
		instance *timetable.TtInstance,
		teacher int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Teachers[teacher].MaxGapsPerWeek =
				instance.TtData_0.Db.Teachers[teacher].MaxGapsPerWeek
		} else {
			instance.TtData.Db.Teachers[teacher].MaxGapsPerWeek = -1
		}
	}
}

//------------------------------------------------------------------

func test_teacher_sequence(
	instance_0 *timetable.TtInstance) *timetable.TtInstance {
	// Make a new instance, reinstating all the teacher constraints

	instance := newInstance(instance_0, "TEACHER_ALL_CONSTRAINTS")
	tt_data := instance.TtData
	db := tt_data.Db

	// Get original teacher constraints
	db.Teachers = instance.TtData_0.Db.Teachers

	instance.FailurePath = timetable.TtChainedFunc{
		Delay: 0, Func: teachers_find_difficult_constraints}

	//TODO

	//instance.SuccessPath = timetable.TtChainedFunc{
	//	Delay: 0, Func: test_...}

	//TODO?
	instance.Timeout = TEST_TIMEOUT

	//TODO: Enable gaps only later?

	// Request start of instance
	instance.NewInstance <- instance
	return instance
}

func teachers_find_difficult_constraints(
	instance_0 *timetable.TtInstance) *timetable.TtInstance {
	instance := newInstance(instance_0, "TEACHER_MIN_LESSONS_PER_DAY")

	disable_teacher_constraints(instance)

	// Start with the MinLessonsPerDay constraints
	n := len(instance.TtData.Db.Teachers)
	f := cfmap[TMinLessonsPerDay]
	for i := range n {
		f(instance, i, true)
	}

	//TODO?
	instance.Timeout = TEST_TIMEOUT

	//instance.FailurePath = timetable.TtChainedFunc{
	//	Delay: 0, Func: find_teacher_min_lessons}

	instance.SuccessPath = timetable.TtChainedFunc{
		Delay: 0, Func: teacher_max_lessons_per_day}

	// Request start of instance
	instance.NewInstance <- instance
	return instance
}

func teacher_max_lessons_per_day(
	instance_0 *timetable.TtInstance) *timetable.TtInstance {
	// Add the MinLessonsPerDay constraints

	instance := newInstance(instance_0, "TEACHER_MAX_LESSONS_PER_DAY")

	n := len(instance.TtData.Db.Teachers)
	f := cfmap[TMaxLessonsPerDay]
	for i := range n {
		f(instance, i, true)
	}

	//TODO?
	instance.Timeout = TEST_TIMEOUT

	//instance.FailurePath = timetable.TtChainedFunc{
	//	Delay: 0, Func: find_teacher_max_lessons}

	//instance.SuccessPath = timetable.TtChainedFunc{
	//	Delay: 0, Func: teacher_max_days}

	// Request start of instance
	instance.NewInstance <- instance
	return instance
}

// Search for teachers having difficulties with the given constraint.
//TODO: How to propagate the search parameters?
/*
func find_difficult_teachers(
	instance_0 *timetable.TtInstance,
	constraint int,
) *timetable.TtInstance {

	instance := newInstance(instance_0, "TEACHER__"+xxx)
	f := cfmap[constraint]
	n := len(instance.TtData_0.Db.Teachers)
	if n < 4 {
		//TODO: a linear test sequence
	} else {
		lim := n / 2
		for i := range lim {
			f(instance, i, false)
		}
	}

	// Request start of instance
	instance.NewInstance <- instance
	return instance
}
*/
