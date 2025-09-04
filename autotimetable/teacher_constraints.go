package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
)

/*
The idea is to provide for each constraint type a function to switch the
constraint on. As it must be possible to combine the switches, the `TtData`
of the input instance needs to be modified, rather than starting from
scratch with the `TtData_0`.

A further search, when a constraint type has been identified as potentially
difficult, would be for individual teachers.

Of course, only constraints which are actually specified need to be activated.
At least at the individual teachers level, it might be worth filtering the
list before starting the binary search.

//TODO: This is a more general comment:
Although a binary search can be relatively efficient, this efficiency will
be reduced if more than one of the components is difficult, especially if the
difficulty arises from the combination, which they mostly do. On the other
hand, checking all combinations is not feasible (because of the enormous
number). The hope is that by offering some assistance in narrowing down the
difficult constraints to particular types, and perhaps individuals within
those types, that the user can find a way to adjust the constraints to make
the construction of the timetable possible.

	t.MinLessonsPerDay = -1 // unconstrained
	t.MaxLessonsPerDay = -1 // unconstrained
	t.MaxDays = -1          // unconstrained
	t.MaxGapsPerDay = -1    // unconstrained
	t.MaxGapsPerWeek = -1   // unconstrained
	t.MaxAfternoons = -1    // unconstrained
	t.LunchBreak = false
*/

type tConstraintSwitch struct {
	MinLessonsPerDay bool
	MaxLessonsPerDay bool
	MaxAfternoons    bool
	MaxDays          bool
	LunchBreak       bool
	MaxGapsPerDay    bool
	MaxGapsPerWeek   bool
}

// Create a new Teacher node, (shallow) copying from the supplied one,
// which should be from the original data.
// Select its active constraints from the `selection` argument.
func set_teacher_constraints(
	t0 *base.Teacher, selection tConstraintSwitch,
) *base.Teacher {
	t := *t0
	if selection.MinLessonsPerDay {
		t.MinLessonsPerDay = t0.MinLessonsPerDay
	} else {
		t.MinLessonsPerDay = -1 // unconstrained
	}
	if selection.MaxLessonsPerDay {
		t.MaxLessonsPerDay = t0.MaxLessonsPerDay
	} else {
		t.MaxLessonsPerDay = -1 // unconstrained
	}
	if selection.MaxAfternoons {
		t.MaxAfternoons = t0.MaxAfternoons
	} else {
		t.MaxAfternoons = -1 // unconstrained
	}
	if selection.MaxDays {
		t.MaxDays = t0.MaxDays
	} else {
		t.MaxDays = -1 // unconstrained
	}
	if selection.LunchBreak {
		t.LunchBreak = t0.LunchBreak
	} else {
		t.LunchBreak = false // not required
	}
	if selection.MaxGapsPerDay {
		t.MaxGapsPerDay = t0.MaxGapsPerDay
	} else {
		t.MaxGapsPerDay = -1 // unconstrained
	}
	if selection.MaxGapsPerWeek {
		t.MaxGapsPerWeek = t0.MaxGapsPerWeek
	} else {
		t.MaxGapsPerWeek = -1 // unconstrained
	}
	return &t
}

func test_teacher_sequence(
	instance_0 *timetable.TtInstance) *timetable.TtInstance {
	// Make a new instane, reinstating all the teacher constraints

	instance := newInstance(instance_0, "TEACHER_ALL_CONSTRAINTS")
	tt_data := instance.TtData
	db := tt_data.Db

	// Get original teacher constraints
	db.Teachers = instance.TtData_0.Db.Teachers

	instance.FailurePath = timetable.TtChainedFunc{
		Delay: 0, Func: teacher_min_lessons_per_day}

	//TODO

	//instance.SuccessPath = timetable.TtChainedFunc{
	//	Delay: 0, Func: test_class_constraints}

	//TODO: Set Timeout?
	//TODO: Enyble gaps only later?

	// Request start of instance
	instance.NewInstance <- instance
	return instance
}

//TODO: It might be better to save the constraint enablements (perhaps
// with those for individual teachers, too ...) in the instance.

func teacher_min_lessons_per_day(
	instance_0 *timetable.TtInstance) *timetable.TtInstance {
	// Start with the MinLessonsPerDay constraints
	cs := tConstraintSwitch{
		MinLessonsPerDay: true,
	}

	instance := newInstance(instance_0, "TEACHER_MIN_LESSONS_PER_DAY")
	tt_data := instance.TtData
	db := tt_data.Db

	tlist := []*base.Teacher{}
	for _, t := range instance.TtData_0.Db.Teachers {
		tlist = append(tlist, set_teacher_constraints(t, cs))
	}
	db.Teachers = tlist

	//TODO

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
	// Start with the MinLessonsPerDay constraints
	cs := tConstraintSwitch{
		MinLessonsPerDay: true,
		MaxLessonsPerDay: true,
	}

	instance := newInstance(instance_0, "TEACHER_MAX_LESSONS_PER_DAY")
	tt_data := instance.TtData
	db := tt_data.Db

	tlist := []*base.Teacher{}
	for _, t := range instance.TtData_0.Db.Teachers {
		tlist = append(tlist, set_teacher_constraints(t, cs))
	}
	db.Teachers = tlist

	//TODO

	//instance.FailurePath = timetable.TtChainedFunc{
	//	Delay: 0, Func: find_teacher_min_lessons}

	//instance.SuccessPath = timetable.TtChainedFunc{
	//	Delay: 0, Func: teacher_max_days}

	// Request start of instance
	instance.NewInstance <- instance
	return instance

}

//TODO: Need to keep records of exactly which constraints are (dis)abled!

func binary_filter(
	instance_0 *timetable.TtInstance,
	tag string,
	flist []func(*timetable.TtInstance) *timetable.TtData,
) {
	//TODO: It is probably not so good to start follow-ons here until their
	// path has been confirmed correct.

	//TODO: How to pass the activation lists to follow-ons?! I suppose it has
	// to be in the TtInstance.

	llen := len(flist)
	if llen < 3 {
		// special, linear, treatment
	} else {
		h := llen / 2

		inst0 := newInstance(instance_0, tag) // TODO: tag ...
		for _, f := range flist[:h] {
			f(inst0)
		}
		//TODO: start it ...

		// if fail: split further until succeed (or pass on unchanged)

		// if succeed:

		inst1 := newInstance(instance_0, tag) // TODO: tag ...
		for _, f := range flist[h:] {
			f(inst1)
		}
		//TODO: start it ...

		// evaluate ...
	}
}
