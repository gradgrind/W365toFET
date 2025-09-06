package autotimetable

import "W365toFET/timetable"

/*
The idea is for each constraint to have a function to switch the constraint
on/off. The constraint type is indexed, the individual instances of a
constraint type also have indexes, so two indexes are needed to refer to
a single constraint.

The cumulative effects of the constraint switches are to be found in the
`TtData` supplied in the current `TtInstance`, which is used as the base
upon which the switch functions work.

Of course, only constraints which are actually specified in the source data
need to be tested. So it might be worth filtering the lists before starting
the test sequences (which may use a binary search method).

//TODO: This is a more general comment:
Although a binary search can be relatively efficient, this efficiency will
be reduced if more than one of the components is difficult, especially if the
difficulty arises from the combination, which they mostly do. On the other
hand, checking all combinations is not feasible (because of the enormous
number). The hope is that by offering some assistance in narrowing down the
difficult constraints to particular types, and perhaps individuals within
those types, that the user can find a way to adjust the constraints to make
the construction of the timetable possible.
*/

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// TODO: This probably needs to encompass ALL constraints ...
// TODO: Map constraint indexes to constraint names?
const (
	TMinLessonsPerDay int = iota
	TMaxLessonsPerDay
	TMaxAfternoons
	TMaxDays
	TLunchBreak
	TMaxGapsPerDay
	TMaxGapsPerWeek

	CMinLessonsPerDay
	CMaxLessonsPerDay
	CMaxAfternoons
	CLunchBreak
	CForceFirstHour
	CMaxGapsPerDay
	CMaxGapsPerWeek

	// ...

	LastConstraint // can be used for sizing function map, etc.
)

var cfmap [LastConstraint]func(*timetable.TtInstance, int, bool)

func disable_all_constraints(instance *timetable.TtInstance) {

	disable_class_constraints(instance)
	disable_teacher_constraints(instance)

	tt_data := instance.TtData

	// Remove general constraints
	for k := range tt_data.Constraints {
		tt_data.Constraints[k] = nil
	}

	// ... and special ones
	tt_data.MinDaysBetweenLessons = nil
	tt_data.ParallelLessons = nil

	// The room constraints are available in the `timetable.CourseInfo`
	// items accessible via the `CourseInfo` pointer in the individual
	// `timetable.Activity` items.
	tt_data.WITHOUT_ROOM_PLACEMENTS = true
}

// Disable all class constraints
func disable_class_constraints(instance *timetable.TtInstance) {
	n := len(instance.TtData.Db.Classes)
	for _, ci := range []int{
		CMinLessonsPerDay,
		CMaxLessonsPerDay,
		CMaxAfternoons,
		CLunchBreak,
		CForceFirstHour,
		CMaxGapsPerDay,
		CMaxGapsPerWeek,
	} {
		f := cfmap[ci]
		for i := range n {
			f(instance, i, false)
		}
	}

}

// Disable all teacher constraints
func disable_teacher_constraints(instance *timetable.TtInstance) {
	n := len(instance.TtData.Db.Teachers)
	for _, ci := range []int{
		TMinLessonsPerDay,
		TMaxLessonsPerDay,
		TMaxAfternoons,
		TMaxDays,
		TLunchBreak,
		TMaxGapsPerDay,
		TMaxGapsPerWeek,
	} {
		f := cfmap[ci]
		for i := range n {
			f(instance, i, false)
		}
	}
}

// TODO???
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
