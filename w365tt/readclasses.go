package w365tt

import (
	"W365toFET/logging"
	"strconv"
)

func (dbp *DbTopLevel) readClasses() {
	// Every Class-Group must be within one – and only one – Class-Division.
	// To handle that, the Group references are first gathered here. Then,
	// when a Group is "used" it is flagged. At the end, any unused Groups
	// can be found and reported.
	pregroups := map[Ref]bool{}
	for _, n := range dbp.Groups {
		pregroups[n.Id] = false
	}

	for i := 0; i < len(dbp.Classes); i++ {
		n := &dbp.Classes[i]

		if n.NotAvailable == nil {
			// Avoid a null value
			n.NotAvailable = []TimeSlot{}
		}
		// MaxAfternoons = 0 has a special meaning (all blocked)
		if n.MaxAfternoons == 0 {
			n.MaxAfternoons = -1
			dbp.handleZeroAfternoons(&n.NotAvailable)
		}

		// Handle special "classes" for stand-ins
		//TODO: Specification still unclear.
		if n.Year <= 0 { //TODO: should be < 0
			// Pending clarification of what exactly this should do, disregard
			// any divisions – there shouldn't be any anyway.
			n.Tag = "" // no students – should already be empty
			continue
		}

		// Get the divisions and flag their Groups.
		for i, wdiv := range n.Divisions {
			if wdiv.Name == "" {
				n.Divisions[i].Name = "#div" + strconv.Itoa(i+1)
			}
			glist := []Ref{}
			for _, g := range wdiv.Groups {
				// get Tag
				flag, ok := pregroups[g]
				if ok {
					if flag {
						logging.Error.Fatalf("Group Defined in"+
							" multiple Divisions:\n  -- %s\n", g)
					}
					// Flag Group and add to division's group list
					pregroups[g] = true
					glist = append(glist, g)
				} else {
					logging.Error.Printf("Unknown Group in Class %s,"+
						" Division %s:\n  %s\n", n.Tag, wdiv.Name, g)
				}
			}
			// Accept Divisions which have too few Groups at this stage.
			if len(glist) < 2 {
				logging.Warning.Printf("In Class %s,"+
					" not enough valid Groups (>1) in Division %s\n",
					n.Tag, wdiv.Name)
			}
			n.Divisions[i].Groups = glist
		}
	}
	for g, used := range pregroups {
		if !used {
			logging.Error.Printf("Group not in Division, removing:\n  %s,", g)
			delete(dbp.Elements, g)
			//TODO: Also remove from Groups list?
		}
	}
}
