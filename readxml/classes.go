package readxml

import (
	"W365toFET/base"
	"fmt"
	"slices"
	"strconv"
)

/* TODO-- see readClasses
func (cdata *conversionData) readGroups() {
	for i := 0; i < len(cdata.xmlin.Groups); i++ {
		n := &cdata.xmlin.Groups[i]
		e := cdata.db.NewGroup(n.Id)
		e.Tag = n.Shortcut
	}
}
*/

func (cdata *conversionData) readDivisions() {
	for i := 0; i < len(cdata.xmlin.Divisions); i++ {
		n := &cdata.xmlin.Divisions[i]
		cdata.divisions[n.Id] = n
	}
}

func (cdata *conversionData) readClasses() {
	// Every Class-Group must be within one – and only one – Class-Division.
	// To handle that, the Group references are first gathered here. Then,
	// when a Group is "used" it is initialized. At the end, any unused Groups
	// can be found and reported.
	// Groups which belong to a class, but not a division are added to a
	// special division with no name.
	db := cdata.db
	pregroups := map[Ref]*base.Class{}
	for _, n := range cdata.xmlin.Groups {
		pregroups[n.Id] = nil
	}

	//TODO???
	db.GroupRefMap = map[base.Ref]base.Ref{}

	slices.SortFunc(cdata.xmlin.Classes, func(a, b Class) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	ndays := len(db.Days)
	nhours := len(db.Hours)
	for i := 0; i < len(cdata.xmlin.Classes); i++ {
		n := &cdata.xmlin.Classes[i]
		e := db.NewClass(n.Id)
		e.Name = n.Name
		e.Year = n.Level
		e.Letter = n.Letter
		e.Tag = strconv.Itoa(n.Level) + n.Letter

		notAvailable := cdata.getAbsences(n.Absences,
			fmt.Sprintf("In Class %s (Absences)", n.Id))
		// MaxAfternoons = 0 has a special meaning (all blocked)
		maxpm := n.MaxAfternoons
		if maxpm >= ndays {
			maxpm = -1
		}
		e.NotAvailable = handleZeroAfternoons(db, notAvailable, maxpm)
		if maxpm == 0 {
			maxpm = -1
		}

		// Add a Group for the whole class (not provided by W365).
		classGroup := db.NewGroup("")
		classGroup.Tag = ""
		db.GroupRefMap[e.Id] = classGroup.Id
		e.ClassGroup = classGroup.Id

		if cdata.isStandIns(n.Categories, n.Id) {
			e.Year = -1
			e.Letter = ""
			e.Tag = ""
			e.Divisions = []base.Division{}
			e.MinLessonsPerDay = -1
			e.MaxLessonsPerDay = -1
			e.MaxGapsPerDay = -1
			e.MaxGapsPerWeek = -1
			e.MaxAfternoons = -1
			e.LunchBreak = false
			e.ForceFirstHour = false
			continue
		}

		// Handle no-lunch-break flag on class.
		lb := cdata.withLunchBreak(n.Categories, n.Id)
		maxlpd := n.MaxLessonsPerDay
		if lb {
			if maxlpd >= nhours-1 {
				maxlpd = -1
			}
		} else if maxlpd >= nhours {
			maxlpd = -1
		}

		// Get the Divisions and their Groups.
		divs := []base.Division{}
		for i, wdivref := range splitRefList(n.Divisions) {
			wdiv, ok := cdata.divisions[wdivref]
			if !ok {
				base.Error.Fatalf("In Class %s:\n  -- Invalid Division: %s\n",
					n.Id, wdivref)
			}
			dname := wdiv.Name
			if dname == "" {
				dname = "#div" + strconv.Itoa(i+1)
			}
			glist := []Ref{}
			for _, gref := range splitRefList(wdiv.Groups) {
				c, ok := pregroups[gref]
				if ok {
					if c != nil {
						base.Error.Fatalf("Group Defined in"+
							" multiple Divisions:\n  -- %s\n", gref)
					}
					// Flag Group and add to division's group list
					pregroups[gref] = e
					glist = append(glist, gref)
				} else {
					base.Error.Fatalf("Unknown Group in Class %s,"+
						" Division %s:\n  %s\n", e.Tag, wdiv.Name, gref)
				}
			}
			if len(glist) < 2 {
				base.Error.Fatalf("In Class %s,"+
					" not enough valid Groups (>1) in Division %s\n",
					e.Tag, wdiv.Name)
			}
			divs = append(divs, base.Division{
				Name:   dname,
				Groups: glist,
			})
		}

		// Check for Groups not in a Division
		xdiv := []Ref{} // pseudodivision for groups not in a division
		for _, gref := range splitRefList(n.Groups) {
			c, ok := pregroups[gref]
			if ok {
				if c == nil {
					xdiv = append(xdiv, gref)
					pregroups[gref] = e
				}
			} else {
				base.Error.Fatalf("Unknown Group in Class %s,"+
					" no Division:\n  %s\n", e.Tag, gref)
			}
		}
		if len(xdiv) != 0 {
			divs = append(divs, base.Division{
				Name:   "",
				Groups: xdiv,
			})
		}

		e.Divisions = divs
		e.MinLessonsPerDay = n.MinLessonsPerDay
		e.MaxLessonsPerDay = maxlpd
		e.MaxGapsPerDay = -1
		e.MaxGapsPerWeek = 0
		e.MaxAfternoons = maxpm
		e.LunchBreak = lb
		e.ForceFirstHour = n.ForceFirstHour
	}

	// Copy Groups.
	for _, n := range cdata.xmlin.Groups {
		if pregroups[n.Id] == nil {
			base.Error.Printf("Group not attached to Class, removing:\n"+
				"  -- %s\n", n.Id)
			continue
		}
		g := db.NewGroup(n.Id)
		g.Tag = n.Shortcut
		db.GroupRefMap[n.Id] = n.Id // mapping to itself is correct!
	}
}

func (cdata *conversionData) getCourseGroups(c *Course) []Ref {
	//
	// Deal with the Groups field of a Course – in W365 the entries can
	// be either a Group or a Class. For the base db they must all be
	// base.Groups.
	//
	glist := []Ref{}
	for _, ref := range splitRefList(c.Groups) {
		s, ok := cdata.db.Elements[ref]
		if ok {
			_, ok := s.(*base.Group)
			if ok {
				glist = append(glist, ref)
				continue
			}
			c, ok := s.(*base.Class)
			if ok {
				glist = append(glist, c.ClassGroup)
				continue
			}
		}
		base.Error.Fatalf("In Course %s:\n  -- Invalid Course Group: %s\n",
			c.Id, ref)
	}
	return glist
}
