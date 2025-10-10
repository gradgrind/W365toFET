package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"fmt"
)

func new_current_instance(instance *TtInstance) {
	ttdata := instance.TtData
	base.Message.Printf("+++ %s\n",
		ttdata.Description)

	// Read placements
	alist := timetable.BACKEND.Results(ttdata)
	if REMOVE_OLD_DATA {
		timetable.BACKEND.Clear(ttdata)
	}

	//TODO

	//TODO: It might be better to return all activities as their integer
	// Id, supplying the mapping separately.

	//TODO: At least rooms are returned as resource indexes – they should
	// be simple room indexes (and a room index to room reference map
	// would be needed.

	for _, a := range alist {
		_ = a
		//fmt.Printf("§§§ %+v\n", a)
	}

	for ctype, clist := range instance.HardConstraintEnabled {
		x := TtData_0.HardConstraints[timetable.ConstraintType(ctype)]
		for i, b := range clist {
			if !b {
				fmt.Printf("$ -- %+v\n", x[i])
			}
		}
	}
}
