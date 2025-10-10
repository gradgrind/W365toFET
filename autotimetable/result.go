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

	//TODO: Can I make the constraints a bit more self-explanatory? Perhaps
	// by using structures which indicate what their fields are.

	for _, a := range alist {
		_ = a
		//fmt.Printf("§§§ %+v\n", a)
	}

	for ctype, clist := range instance.HardConstraintEnabled {
		ctname := timetable.ConstraintType(ctype).String()
		x := TtData_0.HardConstraints[timetable.ConstraintType(ctype)]
		for i, b := range clist {
			if !b {
				fmt.Printf("$ -- %s: %v\n", ctname, x[i])
			}
		}
	}
}
