package w365tt

import (
	"W365toFET/base"
	"strconv"
)

func (db *DbTopLevel) readClasses(newdb *base.DbTopLevel) {
	// Every Class-Group must be within one – and only one – Class-Division.
	// To handle that, the Group references are first gathered here. Then,
	// when a Group is "used" it is flagged. At the end, any unused Groups
	// can be found and reported.
	pregroups := map[Ref]bool{}
	for _, n := range db.Groups {
		pregroups[n.Id] = false
	}

	db.GroupRefMap = map[base.Ref]base.Ref{}
	for _, e := range db.Classes {
		// MaxAfternoons = 0 has a special meaning (all blocked)
		amax := e.MaxAfternoons
		tsl := db.handleZeroAfternoons(e.NotAvailable, amax)
		if amax == 0 {
			amax = -1
		}

		// Get the divisions and flag their Groups.
		divs := []base.Division{}
		for i, wdiv := range e.Divisions {
			dname := wdiv.Name
			if dname == "" {
				dname = "#div" + strconv.Itoa(i+1)
			}
			glist := []Ref{}
			for _, g := range wdiv.Groups {
				// get Tag
				flag, ok := pregroups[g]
				if ok {
					if flag {
						base.Error.Fatalf("Group Defined in"+
							" multiple Divisions:\n  -- %s\n", g)
					}
					// Flag Group and add to division's group list
					pregroups[g] = true
					glist = append(glist, g)
				} else {
					base.Error.Printf("Unknown Group in Class %s,"+
						" Division %s:\n  %s\n", e.Tag, wdiv.Name, g)
				}
			}
			// Accept Divisions which have too few Groups at this stage.
			if len(glist) < 2 {
				base.Warning.Printf("In Class %s,"+
					" not enough valid Groups (>1) in Division %s\n",
					e.Tag, wdiv.Name)
			}
			divs = append(divs, base.Division{
				Name:   dname,
				Groups: glist,
			})
		}

		// Add a Group for the whole class (not provided by W365).
		classGroup := newdb.NewGroup("")
		classGroup.Tag = ""
		db.GroupRefMap[e.Id] = classGroup.Id

		n := newdb.NewClass(e.Id)
		n.Tag = e.Tag
		n.Year = e.Year
		n.Letter = e.Letter
		n.Name = e.Name
		n.NotAvailable = tsl
		n.Divisions = divs
		n.MinLessonsPerDay = e.MinLessonsPerDay
		n.MaxLessonsPerDay = e.MaxLessonsPerDay
		n.MaxGapsPerDay = e.MaxGapsPerDay
		n.MaxGapsPerWeek = e.MaxGapsPerWeek
		n.MaxAfternoons = e.MaxAfternoons
		n.LunchBreak = e.LunchBreak
		n.ForceFirstHour = e.ForceFirstHour
		n.ClassGroup = classGroup.Id
	}

	// Copy Groups.
	for _, n := range db.Groups {
		if pregroups[n.Id] {
			g := newdb.NewGroup(n.Id)
			g.Tag = n.Tag
			db.GroupRefMap[n.Id] = n.Id // mapping to itself is correct!
		} else {
			base.Error.Printf("Group not in Division, removing:\n  %s,",
				n.Id)
		}
	}
}
