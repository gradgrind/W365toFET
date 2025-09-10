package autotimetable

import (
	"W365toFET/timetable"
)

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

	LastConstraint // not a real constraint, it can be used as the total
	// number of constraints.
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

//TODO: To assist in reporting the difficult constraints, I suggest
// maintaining a map (or list) of lists where the activation state of
// each constraint is registered, once the test for a constraint type
// are done.

// TODO: Another attempt at binary (?) search patterns ...
// Return a "working" subset
// TODO: Do I still need SearchInfo?
func bs2(
	instance *timetable.TtInstance,
	constraint int,
	index0 int,
	number int,
	tag string,
) *timetable.TtInstance {
	f := cfmap[constraint]
	if number == 1 {
		inst := newInstance(instance, tag)
		//??
		f(inst, index0, true)

		//TODO ...
		// run trial, waiting ...
		// -> completed

		//TODO: inst.State may be wrong here, because that might be
		// set later in tick loop ...

		switch inst.State {
		case 1:
			// Success: Update "enabled" matrix
			inst.ConstraintEnableMatrix[constraint][index0] = true
			return inst //???
		case 5:
			// Process cancelled
			return nil //???
		default:
			// failed, divide first half
			return instance //???
		}
	}

	// Test the first half
	h := number / 2

	inst := newInstance(instance, tag+"_0")

	i := index0
	for range h {
		f(inst, i, true)
		i++
	}

	//TODO ...
	// run trial, waiting ...
	// -> completed

	switch inst.State {
	case 1:
		// Success: Update "enabled" matrix
		i = index0
		for range h {
			inst.ConstraintEnableMatrix[constraint][i] = true
			i++
		}
	case 5:
		// Process cancelled
		return nil //???
	default:
		// Failed, divide first half
		inst = bs2(instance, constraint, index0, h, inst.Description)
		if inst == nil {
			return nil
		}
	}

	// Second half
	inst2 := newInstance(inst, instance.Description+"_1")

	i = h
	for range number - h {
		f(inst, i, true)
		i++
	}

	//TODO ...
	// run trial, waiting ...
	// -> completed

	switch inst2.State {
	case 1:
		// Success: Update "enabled" matrix
		i = h
		for range number - h {
			inst2.ConstraintEnableMatrix[constraint][i] = true
			i++
		}
	case 5:
		// Process cancelled
		return nil //???
	default:
		// Failed, divide second half
		inst2 = bs2(inst, constraint, h, number-h, inst2.Description)
	}

	return inst2 //??
}

var TIMEOUT_1 int = 30
var TIMEOUT_2 int = 10

func bs3(
	instance *timetable.TtInstance,
	constraint int,
	index0 int,
	number int,
	tag string,
) *timetable.TtInstance {
	f := cfmap[constraint]

	// Run with all constraints enabled
	inst0 := newInstance(instance, tag+"_0")
	i := index0
	for range number {
		f(inst0, i, true)
		i++
	}
	inst0.Timeout = TIMEOUT_1
	//TODO ...
	// run trial in goroutine, don't wait here
	// -> completed
	// cc = 0 -> success: cancel subroutines? set flags, return inst0
	// cc = 1 -> failed, do nothing except trigger untriggered fail process
	// cc = -1 -> cancelled, return nil

	//???

	if number == 1 {
		inst := newInstance(instance, tag)
		//??
		f(inst, index0, true)

		//TODO ...
		// run trial, waiting ...
		// -> completed

		//TODO: inst.State may be wrong here, because that might be
		// set later in tick loop ...

		switch inst.State {
		case 1:
			// Success: Update "enabled" matrix
			inst.ConstraintEnableMatrix[constraint][index0] = true
			return inst //???
		case 5:
			// Process cancelled
			return nil //???
		default:
			// failed, divide first half
			return instance //???
		}
	}

	// Test the first half
	h := number / 2

	inst := newInstance(instance, tag+"_0")

	i = index0
	for range h {
		f(inst, i, true)
		i++
	}

	//TODO ...
	// run trial, waiting ...
	// -> completed

	switch inst.State {
	case 1:
		// Success: Update "enabled" matrix
		i = index0
		for range h {
			inst.ConstraintEnableMatrix[constraint][i] = true
			i++
		}
	case 5:
		// Process cancelled
		return nil //???
	default:
		// Failed, divide first half
		inst = bs2(instance, constraint, index0, h, inst.Description)
		if inst == nil {
			return nil
		}
	}

	// Second half
	inst2 := newInstance(inst, instance.Description+"_1")

	i = h
	for range number - h {
		f(inst, i, true)
		i++
	}

	//TODO ...
	// run trial, waiting ...
	// -> completed

	switch inst2.State {
	case 1:
		// Success: Update "enabled" matrix
		i = h
		for range number - h {
			inst2.ConstraintEnableMatrix[constraint][i] = true
			i++
		}
	case 5:
		// Process cancelled
		return nil //???
	default:
		// Failed, divide second half
		inst2 = bs2(inst, constraint, h, number-h, inst2.Description)
	}

	return inst2 //??
}

// TODO???
/*
func search_constraint_difficulties(
	instance *timetable.TtInstance,
	constraint int,
	index0 int,
	number int,

	//? tag string,
) {
	//TODO: It is probably not so good to start follow-ons here until their
	// path has been confirmed correct.

	//TODO: How to pass the activation lists to follow-ons?! I suppose it has
	// to be in the TtInstance.

	if number < 4 {
		// special, linear, treatment
	} else {
		h := number / 2

		{
			inst0 := newInstance(instance, tag) // TODO: tag ...

			si := &timetable.SearchInfo{
				Constraint: constraint, Index0: index0, Enabled: make([]bool, h),
			}
			for i := range h {
				si.Enabled[i] = true
			}
			inst0.SearchInfo = si

			//TODO ...
		}
		{
			inst1 := newInstance(instance, tag) // TODO: tag ...

			si := &timetable.SearchInfo{
				Constraint: constraint, Index0: index0, Enabled: make([]bool, h),
			}
			for i := h; i < number; i++ {
				si.Enabled[i] = true
			}
			inst1.SearchInfo = si

			//TODO: start it ...
		}

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

func search_instance_succeeded(
	instance_0 *timetable.TtInstance,
) *timetable.TtInstance {
	si := instance_0.SearchInfo
	i := si.Index0
	for range len(si.Enabled) {
		si.Enabled[i] = true
		i++
	}
	si.Done |= si.Part
	if si.Done == 3 {
		//TODO: Test together.
		// There is a problem, though. What if it fails?
		// Maybe the second half of the test should only be run when the
		// first half has completed?
	}
	return nil
}

func search_instance_failed(
	instance_0 *timetable.TtInstance,
) *timetable.TtInstance {

}
*/
