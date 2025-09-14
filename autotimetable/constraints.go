package autotimetable

import (
	"W365toFET/timetable"
)

/*
Each constraint type has a function to switch the constraint on/off.
For the class and teacher functions these are special functions for each
type, for the general constraints, there is a single function, which has
the constraint type as parameter. The individual constraints of a given
type are also indexed, this being a parameter to the switch function.

The cumulative effects of the constraint switches are to be found in the
`TtData` supplied in the current `TtInstance`, which is used as the base
upon which the switch functions work.

Of course, only constraints which are actually specified in the source data
need to be tested, so the lists are filtered before starting the test
sequences (which may use a binary search method).

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

// Associate enable/disable functions with the constraint indexes
var cfmap [timetable.LastConstraint]func(*TtInstance, int, bool)

func start_constraints(instance *TtInstance) {
	// `instance` itself should have no constraints enabled
	tt_data := instance.Global.TtData_0
	db := tt_data.Db

	// Teacher constraints
	tcmap := collect_teacher_constraints(
		db.Teachers, tt_data.NDays, tt_data.NHours)
	//fmt.Printf("§§§tcmap: %v\n", tcmap)
	for cnx, txlist := range tcmap {
		inst := newInstance(instance, cnx.String())
		f := cfmap[cnx]
		for _, tx := range txlist {
			f(inst, tx, true)
		}
		start_constraint_trial(inst)
	}

	// Class constraints
	ccmap := collect_class_constraints(
		db.Classes, tt_data.NDays, tt_data.NHours)
	//fmt.Printf("§§§ccmap: %v\n", ccmap)
	for cnx, cxlist := range ccmap {
		if len(cxlist) == 0 {
			continue
		}
		inst := newInstance(instance, cnx.String())
		f := cfmap[cnx]
		for _, cx := range cxlist {
			f(inst, cx, true)
		}
		start_constraint_trial(inst)
	}

	// Only hard constraints for now ...
	for k, clist := range tt_data.HardConstraints {
		n := len(clist)
		if n == 0 {
			//TODO: Bug?
			panic("No constraints of type " + k.String())
		}
		inst := newInstance(instance, k.String())
		// Get all list indexes
		cilist := make([]int, n)
		for i := range n {
			cilist[i] = i
		}
		set_hard_constraint_enable_state(inst, k, cilist, true)
		start_constraint_trial(inst)
	}

	//TODO: With rooms? Fixed und choices?
}

// Enable or disable a list of indexed constraints for a particular
// constraint type in the `HardConstraints` collection, changing also
// `HardConstraintEnabled` accordingly.
func set_hard_constraint_enable_state(
	instance *TtInstance,
	constraint_type timetable.ConstraintType,
	indexes []int,
	enable bool,
) {
	if instance.HardConstraintEnabled == nil {
		if !enable {
			return
		}
		instance.HardConstraintEnabled =
			map[timetable.ConstraintType]map[int]bool{}
	}
	cmap, ok := instance.HardConstraintEnabled[constraint_type]
	if !ok {
		if !enable {
			return
		}
		cmap = map[int]bool{}
		instance.HardConstraintEnabled[constraint_type] = cmap
	}
	for _, i := range indexes {
		cmap[i] = enable
	}
	// Reconstruct the constraint list
	newlist := []any{}
	for i, c := range instance.Global.TtData_0.HardConstraints[constraint_type] {
		if cmap[i] {
			newlist = append(newlist, c)
		}
	}
	instance.TtData.HardConstraints[constraint_type] = newlist
}

func start_constraint_trial(instance *TtInstance) {
	instance.NewInstance <- instance // register with tick loop
	//fmt.Printf(" +++ %s: %v\n", instance.TtData.Description, instance.TtData.HardConstraints)
}

func disable_all_constraints(instance *TtInstance) {

	disable_class_constraints(instance)
	disable_teacher_constraints(instance)

	tt_data := instance.TtData

	// Remove general constraints
	for k := range tt_data.SoftConstraints {
		tt_data.SoftConstraints[k] = nil
	}
	for k := range tt_data.HardConstraints {
		tt_data.HardConstraints[k] = nil
	}

	// The room constraints are available in the `timetable.CourseInfo`
	// items accessible via the `CourseInfo` pointer in the individual
	// `timetable.Activity` items.
	tt_data.WITHOUT_ROOM_PLACEMENTS = true
}

// Disable all class constraints
func disable_class_constraints(instance *TtInstance) {
	n := len(instance.TtData.Db.Classes)
	for _, ci := range []timetable.ConstraintType{
		timetable.ClassMinLessonsPerDay,
		timetable.ClassMaxLessonsPerDay,
		timetable.ClassMaxAfternoons,
		timetable.ClassForceFirstHour,
		timetable.ClassLunchBreak,
		timetable.ClassMaxGapsPerDay,
		timetable.ClassMaxGapsPerWeek,
	} {
		f := cfmap[ci]
		for i := range n {
			f(instance, i, false)
		}
	}

}

// Disable all teacher constraints
func disable_teacher_constraints(instance *TtInstance) {
	n := len(instance.TtData.Db.Teachers)
	for _, ci := range []timetable.ConstraintType{
		timetable.TeacherMinLessonsPerDay,
		timetable.TeacherMaxLessonsPerDay,
		timetable.TeacherMaxAfternoons,
		timetable.TeacherMaxDays,
		timetable.TeacherLunchBreak,
		timetable.TeacherMaxGapsPerDay,
		timetable.TeacherMaxGapsPerWeek,
	} {
		f := cfmap[ci]
		for i := range n {
			f(instance, i, false)
		}
	}
}
