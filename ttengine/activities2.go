package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"slices"
	"strings"
)

func (tt *TtCore) addActivities2(
	ttinfo *ttbase.TtInfo,
	t2tt map[Ref]ResourceIndex,
	r2tt map[Ref]ResourceIndex,
	g2tt map[Ref][]ResourceIndex,
) {
	// Construct the Activities from the ttinfo.TtLessons.
	// The first Activity has index 1. Index 0 is kept empty, 0 being an
	// invalid ActivityIndex. Note that the Ttlessons start at index 0,
	// so their indexes must be shifted by 1.
	tt.Activities = make([]*Activity, len(ttinfo.TtLessons)+1)
	warned := []*ttbase.CourseInfo{} // used to warn only once per course
	// Collect non-fixed activities which need placing
	toplace := []ActivityIndex{}

	// The place to get custom different-days constraints is the
	// course, which provides links to all the lessons of the course.
	// However, there is also the possibility of a constraint modifying
	// the default behaviour.
	autoDifferentDays := true
	cadd, ok := ttinfo.Constraints["AutomaticDifferentDays"]
	if ok {
		if len(cadd) > 1 {
			base.Error.Fatalf("Constraint AutomaticDifferentDays exists"+
				" %d times\n", len(cadd))
		}
		if cadd[0].(*base.AutomaticDifferentDays).Weight != base.MAXWEIGHT {
			autoDifferentDays = false
		}
	}

	differentDays := map[Ref]bool{}
	for _, c := range ttinfo.Constraints["DaysBetween"] {
		cc := c.(*base.DaysBetween)
		if cc.DaysBetween == 1 {
			for _, cref := range cc.Courses {
				differentDays[cref] = cc.Weight == base.MAXWEIGHT
			}
		}
	}

	differentDaysJoin := map[Ref][]Ref{}
	for _, c := range ttinfo.Constraints["DaysBetweenJoin"] {
		cc := c.(*base.DaysBetweenJoin)
		if cc.Weight == base.MAXWEIGHT && cc.DaysBetween == 1 {
			differentDaysJoin[cc.Course1] = append(
				differentDaysJoin[cc.Course1], cc.Course2)
			differentDaysJoin[cc.Course2] = append(
				differentDaysJoin[cc.Course2], cc.Course1)
		}
	}

	// All other such constraints are not handled at this stage.

	// Initialize parallel courses data.
	parallels := map[ActivityIndex][]ActivityIndex{}
	for _, pc := range ttinfo.Constraints["ParallelCourses"] {
		//TODO
		pcc := pc.(*base.ParallelCourses)

		// The courses must have the same number of lessons and the
		// lengths of the corresponding lessons must also be the same.

		// Check lesson lengths
		footprint := []int{} // lesson sizes
		ll := 0              // number of lessons in each course
		var alists [][]int   // collect the parallel activities
		for i, cref := range pcc.Courses {
			cinfo := ttinfo.CourseInfo[cref]
			if i == 0 {
				ll = len(cinfo.Lessons)
				alists = make([][]int, ll)
			} else if len(cinfo.Lessons) != ll {
				clist := []string{}
				for _, cr := range pcc.Courses {
					clist = append(clist, string(cr))
				}
				base.Error.Fatalf("Parallel courses have different"+
					" lessons: %s\n",
					strings.Join(clist, ","))
			}
			for j, lix := range cinfo.Lessons {
				l := ttinfo.TtLessons[lix].Lesson
				if i == 0 {
					footprint = append(footprint, l.Duration)
				} else if l.Duration != footprint[j] {
					clist := []string{}
					for _, cr := range pcc.Courses {
						clist = append(clist, string(cr))
					}
					base.Error.Fatalf("Parallel courses have lesson"+
						" mismatch: %s\n",
						strings.Join(clist, ","))
				}
				alists[j] = append(alists[j], lix+1)
			}
		}

		// alists is now a list of lists of parallel Activity indexes.
		pcc.Activities = alists

		if pcc.Weight == base.MAXWEIGHT {
			// Hard constraint – prepare for Activities
			for _, alist := range alists {
				for _, aix := range alist {
					for _, aixp := range alist {
						if aixp != aix {
							parallels[aix] = append(parallels[aix], aixp)
						}
					}
				}
			}
		}
	}

	for i, ttl := range ttinfo.TtLessons {
		aix := i + 1
		l := ttl.Lesson
		p := -1
		if l.Day >= 0 {
			p = l.Day*tt.NHours + l.Hour
		}
		cinfo := ttl.CourseInfo
		resources := []ResourceIndex{}

		for _, tref := range cinfo.Teachers {
			resources = append(resources, t2tt[tref])
		}

		for _, gref := range cinfo.Groups {
			for _, ag := range g2tt[gref] {
				// Check for repetitions
				if slices.Contains(resources, ag) {
					if !slices.Contains(warned, cinfo) {
						base.Warning.Printf(
							"Lesson with repeated atomic group"+
								" in Course: %s\n", ttinfo.View(cinfo))
						warned = append(warned, cinfo)
					}
				} else {
					resources = append(resources, ag)
				}
			}
		}

		for _, rref := range cinfo.Room.Rooms {
			// Only take the compulsory rooms here
			resources = append(resources, r2tt[rref])
		}

		// Prepare the DifferentDays field
		ddlist := []ActivityIndex{}
		// Get different-days info for the course
		dd, ok := differentDays[cinfo.Id]
		if !ok {
			dd = autoDifferentDays
		}
		if dd {
			for _, l := range cinfo.Lessons {
				if l != i {
					ddlist = append(ddlist, l+1) // add the activity index
				}
			}
		}
		for _, cj := range differentDaysJoin[cinfo.Id] {
			cjinfo := ttinfo.CourseInfo[cj]
			for _, l := range cjinfo.Lessons {
				ddlist = append(ddlist, l+1) // add the activity index
			}
		}

		// Sort and compactify parallel activities
		plist, ok := parallels[aix]
		if ok {
			slices.Sort(plist)
			plist = slices.Compact(plist)
		}

		a := &Activity{
			Index:     aix,
			Duration:  l.Duration,
			Resources: resources,
			Fixed:     l.Fixed,
			Placement: p,
			//PossibleSlots: added later (see "makePossibleSlots"),
			DifferentDays: ddlist,
			Parallel:      plist,
		}
		tt.Activities[aix] = a

		// The placement has not yet been tested, so the Placement field
		// may still need to be revoked!

		// First place the fixed lessons, then build the PossibleSlots for
		// non-fixed lessons.

		if p >= 0 {
			if a.Fixed {
				if tt.testPlacement(aix, p) {
					// Perform placement
					tt.placeActivity(aix, p)
				} else {
					//TODO: Maybe this shoud be fatal?
					base.Error.Printf(
						"Placement of Fixed Activity %d @ %d failed:\n"+
							"  -- %s\n",
						aix, p, ttinfo.View(cinfo))
					a.Placement = -1
					a.Fixed = false
				}
			} else {
				toplace = append(toplace, aix)
			}
		}
	}

	//TODO: What about non-fixed lessons parallel to fixed ones? Should they
	// be made fixed? They wouldn't need to be placed a second time ...

	// Build PossibleSlots
	tt.makePossibleSlots()

	// Place non-fixed lessons
	for _, aix := range toplace {
		a := tt.Activities[aix]
		p := a.Placement
		if tt.testPlacement(aix, p) {
			// Perform placement
			tt.placeActivity(aix, p)
		} else {
			// Need CourseInfo for reporting details
			ttl := ttinfo.TtLessons[aix-1]
			cinfo := ttl.CourseInfo
			//
			base.Warning.Printf(
				"Placement of Activity %d @ %d failed:\n"+
					"  -- %s\n",
				aix, p, ttinfo.View(cinfo))
			a.Placement = -1
		}
	}
}

func (tt *TtCore) findClashes(aix ActivityIndex, slot int) []ActivityIndex {
	// Return a list of activities (indexes) which are in conflict with
	// the proposed placement. It assumes the slot is in principle possible –
	// so that it will not, for example, be the last slot of a day if
	// the activity duration is 2.
	clashes := []ActivityIndex{}
	a := tt.Activities[aix]
	day := slot / tt.NHours
	for _, addix := range a.DifferentDays {
		add := tt.Activities[addix]
		if add.Placement >= 0 && add.Placement/tt.NHours == day {
			clashes = append(clashes, addix)
		}
	}
	for _, rix := range a.Resources {
		i := rix*tt.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			c := tt.TtSlots[i+ix]
			if c != 0 {
				clashes = append(clashes, c)
			}
		}
	}
	for _, aixp := range a.Parallel {
		a := tt.Activities[aixp]
		for _, addix := range a.DifferentDays {
			add := tt.Activities[addix]
			if add.Placement >= 0 && add.Placement/tt.NHours == day {
				clashes = append(clashes, addix)
			}
		}
		for _, rix := range a.Resources {
			i := rix*tt.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				c := tt.TtSlots[i+ix]
				if c != 0 {
					clashes = append(clashes, c)
				}
			}
		}
	}
	slices.Sort(clashes)
	return slices.Compact(clashes)
}

// TODO: Can I safely assume that no attempt will be made to unplace fixed
// Activities?
func (tt *TtCore) unplaceActivity(aix ActivityIndex) {
	a := tt.Activities[aix]
	slot := a.Placement
	for _, rix := range a.Resources {
		i := rix*tt.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			tt.TtSlots[i+ix] = 0
		}
	}
	a.Placement = -1
	for _, aixp := range a.Parallel {
		a := tt.Activities[aixp]
		for _, rix := range a.Resources {
			i := rix*tt.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				tt.TtSlots[i+ix] = 0
			}
		}
		a.Placement = -1
	}

}
