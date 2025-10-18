package autotimetable

import (
	"W365toFET/timetable"
)

// Return a constraint-enabled matrix based on the initial hard or soft
// constraint map. Initially all constraints are disabled in this map.
func setup_constraint_map(
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

// Collect the individual constraints in the order given by the
// ConstraintType indexes.
func get_basic_constraints(instance0 *TtInstance, soft bool,
) ([]*TtInstance, int) {
	instances := []*TtInstance{} // one instance per constraint type
	nconstraints := 0            // count constraints
	emap := instance0.ConstraintEnabledMatrix
	var cmap map[timetable.ConstraintType][]any
	if soft {
		cmap = TtData_0.SoftConstraints
	} else {
		cmap = TtData_0.HardConstraints
	}
	for ctype := range timetable.LastConstraint {
		blist := emap[ctype]
		cixlist := []int{}
		for i, b := range blist {
			if !b {
				cixlist = append(cixlist, i)
			}
		}
		if len(cixlist) == 0 {
			continue
		}
		nconstraints += len(cixlist)

		clist, ok := cmap[ctype]
		if !ok {
			continue
		}
		n := len(clist)
		if n == 0 {
			//TODO: Bug?
			panic("No constraints of type " + ctype.String())
		}
		instance := new_instance(
			instance0,
			ctype.String(),
			ctype,
			cixlist,
			CYCLE_TIMEOUT,
			soft)
		instances = append(instances, instance)
	}
	return instances, nconstraints
}

func disable_all_constraints(ttdata *timetable.TtData) {
	// Remove general constraints
	for k := range ttdata.SoftConstraints {
		ttdata.SoftConstraints[k] = nil
	}
	for k := range ttdata.HardConstraints {
		ttdata.HardConstraints[k] = nil
	}
}

func disable_soft_constraints(ttdata *timetable.TtData) {
	// Remove soft constraints
	for k := range ttdata.SoftConstraints {
		ttdata.SoftConstraints[k] = nil
	}
}
