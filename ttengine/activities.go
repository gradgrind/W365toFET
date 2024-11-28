package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"slices"
)

type SlotIndex = ttbase.TtIndex
type ResourceIndex = ttbase.TtIndex
type ActivityIndex = ttbase.TtIndex

type Activity struct {
	Index         ActivityIndex
	Duration      int
	Resources     []ResourceIndex
	Fixed         bool
	Placement     int // day * nhours + hour, or -1 if unplaced
	PossibleSlots []SlotIndex
	DifferentDays []ActivityIndex // hard constraint only
	Parallel      []ActivityIndex // hard constraint only
}

func (tt *TtCore) addActivities(
	ttinfo *ttbase.TtInfo,
	t2tt map[Ref]ResourceIndex,
	r2tt map[Ref]ResourceIndex,
	g2tt map[Ref][]ResourceIndex,
) {
	// Construct the Activities from the ttinfo.TtLessons.
	// The first element (index 0) is kept empty, 0 being an
	// invalid ActivityIndex.
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

		a := &Activity{
			Index:     aix,
			Duration:  l.Duration,
			Resources: resources,
			Fixed:     l.Fixed,
			Placement: p,
			//PossibleSlots: added later (see "makePossibleSlots"),
			DifferentDays: ddlist,
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
					//TODO: MAybe this shoud be fatal?
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

func (tt *TtCore) testPlacement(aix ActivityIndex, slot int) bool {
	// Simple boolean placement test. It assumes the slot is possible –
	// so that it will not, for example, be the last slot of a day if
	// the activity duration is 2.
	a := tt.Activities[aix]
	for _, rix := range a.Resources {
		i := rix*tt.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			if tt.TtSlots[i+ix] != 0 {
				return false
			}
		}
	}
	for _, aixp := range a.Parallel {
		a := tt.Activities[aixp]
		for _, rix := range a.Resources {
			i := rix*tt.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				if tt.TtSlots[i+ix] != 0 {
					return false
				}
			}
		}
	}
	return true
}

/* For testing?
func (tt *TtCore) testPlacement2(aix ActivityIndex, slot int) (int, int) {
	// Simple boolean placement test. It assumes the slot is possible –
	// so that it will not, for example, be the last slot of a day if
	// the activity duration is 2.
	a := tt.Activities[aix]
	for _, rix := range a.Resources {
		i := rix*tt.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			acx := tt.TtSlots[i+ix]
			if acx != 0 {
				return acx, rix
			}
		}
	}
	return 0, 0
}
*/

func (tt *TtCore) placeActivity(aix ActivityIndex, slot int) {
	// Allocate the resources, assuming none of the slots are blocked!
	a := tt.Activities[aix]
	for _, rix := range a.Resources {
		i := rix*tt.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			tt.TtSlots[i+ix] = aix
		}
	}
	for _, aixp := range a.Parallel {
		a := tt.Activities[aixp]
		for _, rix := range a.Resources {
			i := rix*tt.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				tt.TtSlots[i+ix] = aixp
			}
		}
	}
}
