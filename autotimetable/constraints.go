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

func start_basic_constraints(
	null_instance *TtInstance,
	runqueue *RunQueue,
	unconstrained_time int,
) map[*TtInstance]struct{} {
	// `null_instance` itself should have no constraints enabled
	tt_data := null_instance.Global.TtData_0

	// Start the individual constraints in the order given by the
	// ConstraintType indexes.
	instances := map[*TtInstance]struct{}{}
	counter := 0
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
		counter++
		// Get all list indexes
		cilist := make([]int, n)
		for i := range n {
			cilist[i] = i
		}
		instance := new_instance(
			null_instance, k.String(), k, cilist, unconstrained_time)
		enable_hard_constraints(instance, k, cilist)
		// Queue instance for running
		runqueue.Add(instance)
		instances[instance] = struct{}{}
	}

	//TODO: With rooms? Fixed und choices? soft constraints?

	return instances
}

/* TODO -> tick loop
	// Gather the completed instances with single constraint types
	//TODO: Find a better way to exit this loop?
	var finished *TtInstance
	var current *TtInstance = nil
	levelok := []*TtInstance{}   // collect successful runs
	levelfail := []*TtInstance{} // collect unsuccessful runs
	for {
		select {
		case finished = <-instance_done:
			break
		}
		if finished == nil {
			fmt.Printf("§ level 1: %d, level 2: %d\n", len(levelok), len(levelfail))
			break
		}

		fmt.Printf("§FINISHED: %s %d\n",
			finished.TtData.Description, finished.TtData.State)

		// Don't use failed instances until it is clear that there are no ok
		// instances available.
		counter--
		if finished.TtData.State == 1 {

			//TODO: if finished is a "COMPLETE" instance {
			//   end all other instances, making this the result,
			//   by sending on stop channel to steering? }

			levelok = append(levelok, finished)
		} else if finished.TtData.State != 5 {
			// state 5 means abandoned
			//TODO: perhaps state 5 instances shouldn't get here at all?
			levelfail = append(levelfail, finished)
		}

		// Start next level
		if current == nil {

		} else if current.TtData.State != 0 {
			// Can this fail? Or rather, what would that mean?
		}

	}
}
*/

/*
func start_constraints(
	instance *TtInstance,
	instance_done chan *TtInstance,
) {
	// `instance` itself should have no constraints enabled
	tt_data := instance.Global.TtData_0

	// Start the individual constraints in the order given by the
	// ConstraintType indexes.
	counter := 0
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
		inst := newInstance(instance, k.String(), int(k))
		counter++
		// Get all list indexes
		cilist := make([]int, n)
		for i := range n {
			cilist[i] = i
		}
		//TODO? set_hard_constraint_enable_state(inst, k, cilist, true)
		enable_hard_constraints(inst, k, cilist)
		start_constraint_trial(inst)
	}

	//TODO: With rooms? Fixed und choices? soft constraints?

	// Gather the completed instances with single constraint types
	//TODO: Find a better way to exit this loop?
	var finished *TtInstance
	var current *TtInstance = nil
	levelok := []*TtInstance{}   // collect successful runs
	levelfail := []*TtInstance{} // collect unsuccessful runs
	for {
		select {
		case finished = <-instance_done:
			break
		}
		if finished == nil {
			fmt.Printf("§ level 1: %d, level 2: %d\n", len(levelok), len(levelfail))
			break
		}

		fmt.Printf("§FINISHED: %s %d\n",
			finished.TtData.Description, finished.TtData.State)

		// Don't use failed instances until it is clear that there are no ok
		// instances available.
		counter--
		if finished.TtData.State == 1 {

			//TODO: if finished is a "COMPLETE" instance {
			//   end all other instances, making this the result,
			//   by sending on stop channel to steering? }

			levelok = append(levelok, finished)
		} else if finished.TtData.State != 5 {
			// state 5 means abandoned
			//TODO: perhaps state 5 instances shouldn't get here at all?
			levelfail = append(levelfail, finished)
		}

		// Start next level
		if current == nil {

		} else if current.TtData.State != 0 {
			// Can this fail? Or rather, what would that mean?
		}

	}
}
*/

// Enable or disable a list of indexed constraints for a particular
// constraint type in the `TtData.HardConstraints` collection, changing also
// `TtInstance.HardConstraintEnabled` accordingly. The constraints are kept
// in the original order.
func enable_hard_constraints(
	instance *TtInstance,
	constraint_type timetable.ConstraintType,
	indexes []int,
) {
	//fmt.Printf("§ENABLE %s: %v\n", constraint_type.String(), indexes)

	// Mark the constraints in the matrix
	cmap := instance.HardConstraintEnabled[constraint_type]
	for _, i := range indexes {
		cmap[i] = true
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

/*
func start_constraint_trial(instance *TtInstance) {
	instance.Global.NewInstance <- instance // register with tick loop
	//fmt.Printf(" >>>>>> %s\n", instance.TtData.Description)
}
*/

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
