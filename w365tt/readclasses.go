package w365tt

import (
	"W365toFET/base"
	"W365toFET/logging"
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
						logging.Error.Fatalf("Group Defined in"+
							" multiple Divisions:\n  -- %s\n", g)
					}
					// Flag Group and add to division's group list
					pregroups[g] = true
					glist = append(glist, g)
				} else {
					logging.Error.Printf("Unknown Group in Class %s,"+
						" Division %s:\n  %s\n", e.Tag, wdiv.Name, g)
				}
			}
			// Accept Divisions which have too few Groups at this stage.
			if len(glist) < 2 {
				logging.Warning.Printf("In Class %s,"+
					" not enough valid Groups (>1) in Division %s\n",
					e.Tag, wdiv.Name)
			}
			divs = append(divs, base.Division{
				Name:   dname,
				Groups: glist,
			})
		}
		newdb.Classes = append(newdb.Classes, &base.Class{
			Id:               e.Id,
			Tag:              e.Tag,
			Year:             e.Year,
			Letter:           e.Letter,
			Name:             e.Name,
			NotAvailable:     tsl,
			Divisions:        divs,
			MinLessonsPerDay: e.MinLessonsPerDay,
			MaxLessonsPerDay: e.MaxLessonsPerDay,
			MaxGapsPerDay:    e.MaxGapsPerDay,
			MaxGapsPerWeek:   e.MaxGapsPerWeek,
			MaxAfternoons:    e.MaxAfternoons,
			LunchBreak:       e.LunchBreak,
			ForceFirstHour:   e.ForceFirstHour,
		})
	}

	// Copy Groups.
	newdb.Groups = []*base.Group{}
	for _, n := range db.Groups {
		if pregroups[n.Id] {
			newdb.Groups = append(newdb.Groups, &base.Group{
				Id:  n.Id,
				Tag: n.Tag,
			})
		} else {
			logging.Error.Printf("Group not in Division, removing:\n  %s,",
				n.Id)
		}
	}
}
