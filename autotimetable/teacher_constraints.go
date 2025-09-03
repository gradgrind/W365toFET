package autotimetable

import (
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

func teacher_min_lessons_per_day(
	instance_0 *timetable.TtInstance) *timetable.TtInstance {
	// Add the MinLessonsPerDay constraints

	instance := newInstance(instance_0, "TEACHER_MIN_LESSONS_PER_DAY")
	tt_data := instance.TtData
	db := tt_data.Db

	db0 := instance.TtData_0.Db
	base_teachers := db0.Teachers
	for i, tp := range db.Teachers {
		tp.MinLessonsPerDay = base_teachers[i].MinLessonsPerDay
	}

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

	instance := newInstance(instance_0, "TEACHER_MAX_LESSONS_PER_DAY")
	tt_data := instance.TtData
	db := tt_data.Db

	db0 := instance.TtData_0.Db
	base_teachers := db0.Teachers
	for i, tp := range db.Teachers {
		tp.MaxLessonsPerDay = base_teachers[i].MaxLessonsPerDay
	}

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
