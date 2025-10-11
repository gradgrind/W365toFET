package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"encoding/json"
	"os"
	"path/filepath"
)

type RefTag struct {
	Tag string
	Ref timetable.NodeRef
}

type Result struct {
	Time                 int
	Teachers             []RefTag
	Classes              []RefTag
	Rooms                []RefTag
	Activities           []timetable.NodeRef
	Placements           []timetable.ActivityPlacement
	DiscardedConstraints []any
	TotalConstraints     int
}

// Save the result of the current instance as a JSON file.
func new_current_instance(instance *TtInstance) {
	ttdata := instance.TtData
	base.Message.Printf("+++ %s\n",
		ttdata.Description)

	// Read placements
	alist := timetable.BACKEND.Results(ttdata)
	if REMOVE_OLD_DATA {
		timetable.BACKEND.Clear(ttdata)
	}

	// Collect teachers, classes and rooms
	db := ttdata.SharedData.Db
	t2ref := make([]RefTag, len(db.Teachers))
	for i, tnode := range db.Teachers {
		t2ref[i] = RefTag{tnode.GetTag(), tnode.GetRef()}
	}
	c2ref := make([]RefTag, len(db.Classes))
	for i, cnode := range db.Classes {
		c2ref[i] = RefTag{cnode.GetTag(), cnode.GetRef()}
	}
	r2ref := make([]RefTag, len(db.Rooms))
	for i, rnode := range db.Rooms {
		r2ref[i] = RefTag{rnode.GetTag(), rnode.GetRef()}
	}

	// The activity "references"
	a2ref := make([]timetable.NodeRef, len(ttdata.SharedData.Activities))
	for i, anode := range ttdata.SharedData.Activities {
		if i != 0 {
			a2ref[i] = anode.Lesson.Id
		}
	}

	// The discarded constraints
	constraints := []any{}
	n := 0 // count all constraints
	for ctype, clist := range instance.HardConstraintEnabled {
		x := TtData_0.HardConstraints[timetable.ConstraintType(ctype)]
		for i, b := range clist {
			if !b {
				constraints = append(constraints, x[i])
			}
		}
		n += len(clist)
	}

	result := Result{
		Time:                 ttdata.Ticks,
		Teachers:             t2ref,
		Classes:              c2ref,
		Rooms:                r2ref,
		Activities:           a2ref,
		Placements:           alist,
		DiscardedConstraints: constraints,
		TotalConstraints:     n,
	}

	//b, err := json.Marshal(result)
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	fpath := filepath.Join(ttdata.SharedData.WorkingDir,
		ttdata.Description+".json")
	f, err := os.Create(fpath)
	if err != nil {
		panic("Couldn't open output file: " + fpath)
	}
	defer f.Close()
	_, err = f.Write(b)
	if err != nil {
		panic("Couldn't write result to: " + fpath)
	}
}
