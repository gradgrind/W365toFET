package autotimetable

import (
	"W365toFET/timetable"
)

// Each of these functions enables/disables a particular class constraint

func init() {
	cfmap[CMinLessonsPerDay] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MinLessonsPerDay =
				instance.TtData_0.Db.Classes[class].MinLessonsPerDay
		} else {
			instance.TtData.Db.Classes[class].MinLessonsPerDay = -1
		}
	}
	cfmap[CMaxLessonsPerDay] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxLessonsPerDay =
				instance.TtData_0.Db.Classes[class].MaxLessonsPerDay
		} else {
			instance.TtData.Db.Classes[class].MaxLessonsPerDay = -1
		}
	}
	cfmap[CMaxAfternoons] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxAfternoons =
				instance.TtData_0.Db.Classes[class].MaxAfternoons
		} else {
			instance.TtData.Db.Classes[class].MaxAfternoons = -1
		}
	}
	cfmap[CLunchBreak] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].LunchBreak =
				instance.TtData_0.Db.Classes[class].LunchBreak
		} else {
			instance.TtData.Db.Classes[class].LunchBreak = false
		}
	}
	cfmap[CForceFirstHour] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].ForceFirstHour =
				instance.TtData_0.Db.Classes[class].ForceFirstHour
		} else {
			instance.TtData.Db.Classes[class].ForceFirstHour = false
		}
	}
	cfmap[CMaxGapsPerDay] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxGapsPerDay =
				instance.TtData_0.Db.Classes[class].MaxGapsPerDay
		} else {
			instance.TtData.Db.Classes[class].MaxGapsPerDay = -1
		}
	}
	cfmap[CMaxGapsPerWeek] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxGapsPerWeek =
				instance.TtData_0.Db.Classes[class].MaxGapsPerWeek
		} else {
			instance.TtData.Db.Classes[class].MaxGapsPerWeek = -1
		}
	}

}

//------------------------------------------------------------------

func test_class_sequence(
	instance_0 *timetable.TtInstance) *timetable.TtInstance {
	// Make a new instance, reinstating all the class constraints

	instance := newInstance(instance_0, "CLASS_ALL_CONSTRAINTS")
	tt_data := instance.TtData
	db := tt_data.Db

	// Get original class constraints
	db.Classes = instance.TtData_0.Db.Classes

	instance.FailurePath = timetable.TtChainedFunc{
		Delay: 0, Func: classes_find_difficult_constraints}

	instance.SuccessPath = timetable.TtChainedFunc{
		Delay: 0, Func: test_teacher_sequence}

	//TODO?
	instance.Timeout = TEST_TIMEOUT

	//TODO: Enable gaps only later?

	// Request start of instance
	instance.NewInstance <- instance
	return instance
}

func classes_find_difficult_constraints(
	instance_0 *timetable.TtInstance) *timetable.TtInstance {
	instance := newInstance(instance_0, "CLASS_MIN_LESSONS_PER_DAY")

	disable_class_constraints(instance)

	// Start with the CMinLessonsPerDay constraints
	n := len(instance.TtData.Db.Classes)
	f := cfmap[CMinLessonsPerDay]
	for i := range n {
		f(instance, i, true)
	}

	//TODO
	instance.Timeout = TEST_TIMEOUT

	//instance.FailurePath = timetable.TtChainedFunc{
	//	Delay: 0, Func: find_class...}

	instance.SuccessPath = timetable.TtChainedFunc{
		Delay: 0, Func: class_max_lessons_per_day}

	// Request start of instance
	instance.NewInstance <- instance
	return instance
}

// Add the CMaxLessonsPerDay constraints
func class_max_lessons_per_day(
	instance_0 *timetable.TtInstance) *timetable.TtInstance {

	instance := newInstance(instance_0, "CLASS_MAX_LESSONS_PER_DAY")

	n := len(instance.TtData.Db.Classes)
	f := cfmap[CMaxLessonsPerDay]
	for i := range n {
		f(instance, i, true)
	}

	//TODO
	instance.Timeout = TEST_TIMEOUT

	//instance.FailurePath = timetable.TtChainedFunc{
	//	Delay: 0, Func: find_class_max_lessons}

	//TODO: Actually it should go to the next class constraint ...
	instance.SuccessPath = timetable.TtChainedFunc{
		Delay: 0, Func: test_teacher_sequence}

	// Request start of instance
	instance.NewInstance <- instance
	return instance

}
