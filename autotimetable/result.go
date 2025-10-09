package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
)

func new_current_instance(instance *TtInstance) {
	ttdata := instance.TtData
	base.Message.Printf("+++ %s\n",
		ttdata.Description)

	// Read placements
	alist := timetable.BACKEND.Results(ttdata)
	timetable.BACKEND.Clear(ttdata)

	//TODO
	for _, a := range alist {
		_ = a
		//fmt.Printf("§§§ %+v\n", a)
	}

}
