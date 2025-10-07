package autotimetable

import (
	"W365toFET/timetable"
)

/* TODO?
Function `set_hard_constraint_enable_state` activates or deactivates a list
of individual constraints of a given type. The individual constraints are
indexed, the index being in a range determined by the enabled constraints
of that type in the original data.

The cumulative effects of the constraint switches are to be found in the
`TtData` supplied in the current `TtInstance`, which is used as the base
upon which the switch functions work. The current state of all constraints
of the original data is available as a boolean matrix, field
`ConstraintEnableMatrix` of the instance.

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

// Set up the `TtInstance.HardConstraintEnabled` matrix based on the initial
// hard constraint map. Initially all constraints are disabled in this map.
func setup_hard_constraint_map(
	constraints map[timetable.ConstraintType][]any,
) [][]bool {
	cmap := make([][]bool, timetable.LastConstraint)
	for cx, clist := range constraints {
		l := len(clist)
		if l == 0 {
			panic("Bug: Empty constraint list")
		}
		cmap[cx] = make([]bool, l) // default: all entries false
	}
	return cmap
}

func get_basic_constraints(
	null_instance *TtInstance,
	unconstrained_time int,
) []*TtInstance {
	// `null_instance` itself should have no constraints enabled
	tt_data := TtData_0

	// Start the individual constraints in the order given by the
	// ConstraintType indexes.
	instances := []*TtInstance{}
	//counter := 0
	for k := range timetable.LastConstraint {
		// Only hard constraints for now ...
		clist, ok := tt_data.HardConstraints[k]
		if !ok {
			continue
		}
		n := len(clist)
		if n == 0 {
			//TODO: Bug?
			panic("No constraints of type " + k.String())
		}
		//counter++
		// Get all list indexes
		cilist := make([]int, n)
		for i := range n {
			cilist[i] = i
		}
		instance := new_instance(
			null_instance,
			k.String(),
			k,
			cilist,
			max(unconstrained_time*NEXT_STAGE_TIMEOUT_FACTOR/10,
				NEXT_STAGE_TIMEOUT_MIN))
		instances = append(instances, instance)
	}

	//TODO: With rooms? Fixed und choices? soft constraints?

	return instances
}

func disable_all_constraints(ttdata *timetable.TtData) {
	// Remove general constraints
	for k := range ttdata.SoftConstraints {
		ttdata.SoftConstraints[k] = nil
	}
	for k := range ttdata.HardConstraints {
		ttdata.HardConstraints[k] = nil
	}

	// The room constraints are available in the `timetable.CourseInfo`
	// items accessible via the `CourseInfo` pointer in the individual
	// `timetable.Activity` items.
	ttdata.WITHOUT_ROOM_PLACEMENTS = true
}
