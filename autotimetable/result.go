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
	Unfulfilled          map[string][]int
	TotalConstraints     int
}

// Get the result of the current instance as a `Result` structure.
// Save as JSON if debugging.
func new_current_instance(instance *TtInstance) {
	ttdata := instance.TtData
	base.Message.Printf("+++ %s @ %d\n",
		ttdata.Description, ttdata.Ticks)

	// Read placements
	alist := timetable.BACKEND.Results(ttdata)
	//if !DEBUG {
	//	timetable.BACKEND.Clear(ttdata)
	//}

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
			a2ref[i] = anode.Activity.Id
		}
	}

	// The discarded constraints
	constraints := []any{}
	nall := 0 // count all constraints
	unfulfilled := map[string][]int{}
	for ctype, clist := range instance.ConstraintEnabledMatrix {
		x := TtData_0.HardConstraints[timetable.ConstraintType(ctype)]
		ulist := []int{}
		for i, b := range clist {
			if !b {
				constraints = append(constraints, x[i])
				ulist = append(ulist, i)
			}
		}
		if len(ulist) != 0 {
			unfulfilled[timetable.ConstraintType(ctype).String()] = ulist
		}
		nall += len(clist)
	}

	LastResult = &Result{
		Time:                 ttdata.Ticks,
		Teachers:             t2ref,
		Classes:              c2ref,
		Rooms:                r2ref,
		Activities:           a2ref,
		Placements:           alist,
		DiscardedConstraints: constraints,
		Unfulfilled:          unfulfilled,
		TotalConstraints:     nall,
	}

	if DEBUG {
		//b, err := json.Marshal(LastResult)
		b, err := json.MarshalIndent(LastResult, "", "  ")
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
}
