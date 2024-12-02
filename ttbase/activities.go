package ttbase

import (
	"W365toFET/base"
	"slices"
)

type SlotIndex = TtIndex
type ResourceIndex = TtIndex
type ActivityIndex = TtIndex

type Activity struct {
	Index         ActivityIndex
	Duration      int
	Resources     []ResourceIndex
	Fixed         bool
	Placement     int // day * nhours + hour, or -1 if unplaced
	PossibleSlots []SlotIndex
	DifferentDays []ActivityIndex // hard constraint only
	Parallel      []ActivityIndex // hard constraint only

	// Access to basic information
	CourseInfo *CourseInfo
	Lesson     *base.Lesson
}

func (ttinfo *TtInfo) addActivityInfo(
	t2tt map[Ref]ResourceIndex,
	r2tt map[Ref]ResourceIndex,
	g2tt map[Ref][]ResourceIndex,
) {
	// Complete the initialization of the Activities.
	warned := []*CourseInfo{} // used to warn only once per course
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
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		ttl := ttinfo.Activities[aix]
		l := ttl.Lesson
		p := -1
		if l.Day >= 0 {
			p = l.Day*ttinfo.NHours + l.Hour
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

		a := ttinfo.Activities[aix]
		a.Duration = l.Duration
		a.Resources = resources
		a.Fixed = l.Fixed
		a.Placement = p
		//PossibleSlots: added later (see "makePossibleSlots"),
		//DifferentDays: ddlist, // only if not fixed, see below
		a.Parallel = plist
		if !l.Fixed {
			a.DifferentDays = ddlist
		}

		// The placement has not yet been tested, so it may still need to be
		// revoked!
	}

	// Check parallel lessons for compatibility, etc.
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		a := ttinfo.Activities[aix]
		if len(a.Parallel) != 0 {
			continue
		}
		p := a.Placement
		if a.Fixed {
			if p < 0 {
				base.Bug.Fatalf("Fixed activity with no time slot: %d\n", aix)
			}
			for _, paix := range a.Parallel {
				pa := ttinfo.Activities[paix]
				pp := pa.Placement
				if pa.Fixed {
					base.Warning.Printf("Parallel fixed lessons:\n"+
						"  -- %d: %s\n  -- %d: %s\n",
						aix,
						ttinfo.View(ttinfo.Activities[aix].CourseInfo),
						paix,
						ttinfo.View(ttinfo.Activities[paix].CourseInfo),
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
								ttinfo.View(ttinfo.Activities[aix].CourseInfo),
								paix,
								ttinfo.View(ttinfo.Activities[paix].CourseInfo),
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
				pa := ttinfo.Activities[paix]
				pp := pa.Placement
				if pp >= 0 && pp != p {
					// Warn and set ALL to -1
					base.Warning.Printf("Parallel lessons with different"+
						" times (placements revoked):\n  -- %d: %s\n",
						aix,
						ttinfo.View(ttinfo.Activities[aix].CourseInfo))
					a.Placement = -1
					for _, paix := range a.Parallel {
						pa := ttinfo.Activities[paix]
						pa.Placement = -1
					}
					break
				}
			}
		}
	}
	// To avoid multiple placement of parallels, mark Activities which have
	// been placed.
	placed := make([]bool, len(ttinfo.Activities))

	// First place the fixed lessons, then build the PossibleSlots for
	// non-fixed lessons.
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		a := ttinfo.Activities[aix]
		p := a.Placement

		if p >= 0 {
			if placed[aix] {
				continue
			}
			if a.Fixed {
				if ttinfo.testPlacement(aix, p) {
					// Perform placement
					ttinfo.placeActivity(aix, p)
					placed[aix] = true
					for _, paix := range a.Parallel {
						placed[paix] = true
					}
				} else {
					base.Error.Fatalf(
						"Placement of Fixed Activity %d @ %d failed:\n"+
							"  -- %s\n",
						aix, p, ttinfo.View(ttinfo.Activities[aix].CourseInfo))
				}
			} else {
				toplace = append(toplace, aix)
			}
		}
	}

	// Build PossibleSlots
	ttinfo.makePossibleSlots()

	// Place non-fixed lessons
	for _, aix := range toplace {
		if placed[aix] {
			continue
		}
		a := ttinfo.Activities[aix]
		p := a.Placement
		if ttinfo.testPlacement(aix, p) {
			// Perform placement
			ttinfo.placeActivity(aix, p)
			placed[aix] = true
			for _, paix := range a.Parallel {
				placed[paix] = true
			}
		} else {
			// Need CourseInfo for reporting details
			ttl := ttinfo.Activities[aix-1]
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

func (ttinfo *TtInfo) findClashes(aix ActivityIndex, slot int) []ActivityIndex {
	// Return a list of activities (indexes) which are in conflict with
	// the proposed placement. It assumes the slot is in principle possible –
	// so that it will not, for example, be the last slot of a day if
	// the activity duration is 2.
	clashes := []ActivityIndex{}
	a := ttinfo.Activities[aix]
	day := slot / ttinfo.NHours
	for _, addix := range a.DifferentDays {
		add := ttinfo.Activities[addix]
		if add.Placement >= 0 && add.Placement/ttinfo.NHours == day {
			clashes = append(clashes, addix)
		}
	}
	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			c := ttinfo.TtSlots[i+ix]
			if c != 0 {
				clashes = append(clashes, c)
			}
		}
	}
	for _, aixp := range a.Parallel {
		a := ttinfo.Activities[aixp]
		for _, addix := range a.DifferentDays {
			add := ttinfo.Activities[addix]
			if add.Placement >= 0 && add.Placement/ttinfo.NHours == day {
				clashes = append(clashes, addix)
			}
		}
		for _, rix := range a.Resources {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				c := ttinfo.TtSlots[i+ix]
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
func (ttinfo *TtInfo) unplaceActivity(aix ActivityIndex) {
	a := ttinfo.Activities[aix]
	slot := a.Placement
	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			ttinfo.TtSlots[i+ix] = 0
		}
	}
	a.Placement = -1
	for _, aixp := range a.Parallel {
		a := ttinfo.Activities[aixp]
		for _, rix := range a.Resources {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				ttinfo.TtSlots[i+ix] = 0
			}
		}
		a.Placement = -1
	}

}

func (ttinfo *TtInfo) testPlacement(aix ActivityIndex, slot int) bool {
	// Simple boolean placement test. It assumes the slot is possible –
	// so that it will not, for example, be the last slot of a day if
	// the activity duration is 2.
	a := ttinfo.Activities[aix]
	day := slot / ttinfo.NHours
	for _, addix := range a.DifferentDays {
		add := ttinfo.Activities[addix]
		if add.Placement >= 0 && add.Placement/ttinfo.NHours == day {
			return false
		}
	}
	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			if ttinfo.TtSlots[i+ix] != 0 {
				return false
			}
		}
	}
	for _, aixp := range a.Parallel {
		a := ttinfo.Activities[aixp]
		for _, addix := range a.DifferentDays {
			add := ttinfo.Activities[addix]
			if add.Placement >= 0 && add.Placement/ttinfo.NHours == day {
				return false
			}
		}
		for _, rix := range a.Resources {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				if ttinfo.TtSlots[i+ix] != 0 {
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
	a := ttinfo.Activities[aix]
	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			acx := ttinfo.TtSlots[i+ix]
			if acx != 0 {
				return acx, rix
			}
		}
	}
	return 0, 0
}
*/

func (ttinfo *TtInfo) placeActivity(aix ActivityIndex, slot int) {
	// Allocate the resources, assuming none of the slots are blocked!
	a := ttinfo.Activities[aix]
	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			ttinfo.TtSlots[i+ix] = aix
		}
	}
	a.Placement = slot
	for _, aixp := range a.Parallel {
		a := ttinfo.Activities[aixp]
		for _, rix := range a.Resources {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				ttinfo.TtSlots[i+ix] = aixp
			}
		}
		a.Placement = slot
	}
}
