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
	// The first Activity has index 1. Index 0 is kept empty, 0 being an
	// invalid ActivityIndex. To accommodate this, the Ttlessons also start
	// at index 0.
	tt.Activities = make([]*Activity, len(ttinfo.TtLessons))
	warned := []*ttbase.CourseInfo{} // used to warn only once per course
	// Collect non-fixed activities which need placing
	toplace := []ActivityIndex{}

	// Collect the hard-different-days lessons (gap = 1) for each lesson.
	diffdays := map[ActivityIndex][]ActivityIndex{}
	for _, dbc := range ttinfo.MinDaysBetweenLessons {
		if dbc.Weight == base.MAXWEIGHT && dbc.MinDays == 1 {
			alist := dbc.Lessons
			for _, aix := range alist {
				for _, aix2 := range alist {
					if aix2 != aix {
						diffdays[aix] = append(diffdays[aix], aix2)
					}
				}
			}
		}
	}

	// Collect the hard-parallel lessons for each lesson.
	parallels := map[ActivityIndex][]ActivityIndex{}
	for _, pl := range ttinfo.ParallelLessons {
		if pl.Weight == base.MAXWEIGHT {
			// Hard constraint – prepare for Activities
			for _, alist := range pl.LessonGroups {
				for _, aix := range alist {
					for _, aix2 := range alist {
						if aix2 != aix {
							parallels[aix] = append(parallels[aix], aix2)
						}
					}
				}
			}
		}
	}

	// Lessons start at index 1!
	for aix := 1; aix < len(ttinfo.TtLessons); aix++ {
		ttl := ttinfo.TtLessons[aix]
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

		// Sort and compactify different-days activities
		ddlist, ok := diffdays[aix]
		if ok && len(ddlist) > 1 {
			slices.Sort(ddlist)
			ddlist = slices.Compact(ddlist)
		}

		// Sort and compactify parallel activities
		plist, ok := parallels[aix]
		if ok && len(plist) > 1 {
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
			//DifferentDays: ddlist, // only if not fixed, see below
			Parallel: plist,
		}
		if !l.Fixed {
			a.DifferentDays = ddlist
		}
		tt.Activities[aix] = a

		// The placement has not yet been tested, so the Placement field
		// may still need to be revoked!
	}

	//TODO: Move to ttbase?
	// Check parallel lessons for compatibility, etc.
	for aix := 1; aix < len(ttinfo.TtLessons); aix++ {
		a := tt.Activities[aix]
		if len(a.Parallel) != 0 {
			continue
		}
		p := a.Placement
		if a.Fixed {
			if p < 0 {
				base.Bug.Fatalf("Fixed activity with no time slot: %d\n", aix)
			}
			for _, paix := range a.Parallel {
				pa := tt.Activities[paix]
				pp := pa.Placement
				if pa.Fixed {
					base.Warning.Printf("Parallel fixed lessons:\n"+
						"  -- %d: %s\n  -- %d: %s\n",
						aix,
						ttinfo.View(ttinfo.TtLessons[aix].CourseInfo),
						paix,
						ttinfo.View(ttinfo.TtLessons[paix].CourseInfo),
					)
					if pp != p {
						base.Error.Fatalln("Parallel fixed lessons have" +
							" different times")
					}
				} else {
					if pp != p {
						if pp >= 0 {
							base.Warning.Printf("Parallel lessons with"+
								" different times:\n  -- %d: %s\n  -- %d: %s\n",
								aix,
								ttinfo.View(ttinfo.TtLessons[aix].CourseInfo),
								paix,
								ttinfo.View(ttinfo.TtLessons[paix].CourseInfo),
							)
						}
						pa.Placement = p
						pa.Fixed = true
					}
				}
			}
		} else {
			if p < 0 {
				continue
			}
			for _, paix := range a.Parallel {
				pa := tt.Activities[paix]
				pp := pa.Placement
				if pp >= 0 && pp != p {
					//TODO: Warn and set ALL to -1?

				}
			}
		}
	}
	//TODO: How to avoid multiple placement of parallels? Perhaps with a map/set
	// of already placed ones?

	// First place the fixed lessons, then build the PossibleSlots for
	// non-fixed lessons.
	for aix := 1; aix < len(ttinfo.TtLessons); aix++ {
		a := tt.Activities[aix]
		p := a.Placement

		if p >= 0 {
			if a.Fixed {
				if tt.testPlacement(aix, p) {
					// Perform placement
					tt.placeActivity(aix, p)
				} else {
					base.Error.Fatalf(
						"Placement of Fixed Activity %d @ %d failed:\n"+
							"  -- %s\n",
						aix, p, ttinfo.View(ttinfo.TtLessons[aix].CourseInfo))
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

func (tt *TtCore) testPlacement(aix ActivityIndex, slot int) bool {
	// Simple boolean placement test. It assumes the slot is possible –
	// so that it will not, for example, be the last slot of a day if
	// the activity duration is 2.
	a := tt.Activities[aix]
	day := slot / tt.NHours
	for _, addix := range a.DifferentDays {
		add := tt.Activities[addix]
		if add.Placement >= 0 && add.Placement/tt.NHours == day {
			return false
		}
	}
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
		for _, addix := range a.DifferentDays {
			add := tt.Activities[addix]
			if add.Placement >= 0 && add.Placement/tt.NHours == day {
				return false
			}
		}
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
	a.Placement = slot
	for _, aixp := range a.Parallel {
		a := tt.Activities[aixp]
		for _, rix := range a.Resources {
			i := rix*tt.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				tt.TtSlots[i+ix] = aixp
			}
		}
		a.Placement = slot
	}
}
