package ttbase

import (
	"W365toFET/base"
	"fmt"
	"slices"
)

// These types are defined primarily for documentation, to make the types
// of the corresponding items a bit clearer.
type SlotIndex = TtIndex     // index of a time-slot within a week
type ResourceIndex = TtIndex // index into [TtInfo.Resources]
type ActivityIndex = TtIndex // index into [TtInfo.Activities]

type Activity struct {
	// Index is the offset of this activity in [TtInfo.Activities]
	Index ActivityIndex
	// Duration specifies the length in timetable-hours of this activity
	Duration int
	// Resources are teachers, student (atomic!) groups and rooms. They are
	// referenced using indexes into [TtInfo.Resources].
	// [Activity.Resources] lists those required by the Activity.
	Resources []ResourceIndex
	// XRooms is a list of chosen rooms, distinct from the required rooms
	XRooms []ResourceIndex
	// ExtendedGroups is a list of atomic group indexes for those groups
	// in the activity's class(es) which are NOT involved in the activity.
	ExtendedGroups []ResourceIndex
	// Fixed specifies whether the activity must remain in its current slot
	Fixed bool
	// Placement specifies the time-slot in which this activity has been
	// placed (day * nhours + hour), or -1 if unplaced
	Placement SlotIndex
	// PossibleSlots is a list of non-blocked time-slots for this activity
	PossibleSlots []SlotIndex
	// DifferentDays lists activities which must definitely be placed on a
	// different day to the present activity
	DifferentDays []ActivityIndex // hard constraint only
	// Parallel lists activities which must start at the same time as the
	// present activity.
	Parallel []ActivityIndex // hard constraint only

	// Access to basic information about this activity
	CourseInfo *CourseInfo
	Lesson     *base.Lesson
}

// addActivityInfo completes the initialization of the Activities. This
// includes the placement in the timetable structures of all the lessons
// (fixed and non-fixed) which have a placement specified. In this way
// various errors can be checked for.
func (ttinfo *TtInfo) addActivityInfo(
	t2tt map[Ref]ResourceIndex,
	r2tt map[Ref]ResourceIndex,
	g2tt map[Ref][]ResourceIndex,
) {
	warned := []*CourseInfo{} // used to warn only once per course
	// Collect non-fixed activities which need placing
	toplace := []ActivityIndex{}
	//--fmt.Printf("=== %+v\n\n", ttinfo.MinDaysBetweenLessons)
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

	// Lessons (Activities) start at index 1!
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		ttl := ttinfo.Activities[aix]
		cinfo := ttl.CourseInfo
		resources := []ResourceIndex{}

		for _, tref := range cinfo.Teachers {
			resources = append(resources, t2tt[tref])
		}

		// Get class(es) ... and atomic groups
		// This is for finding the "extended groups" – in the activity's
		// class(es) but not involved in the activity. This list may help
		// finding activities which can be placed parallel.
		cagmap := map[base.Ref][]ResourceIndex{}
		for _, gref := range cinfo.Groups {
			cagmap[ttinfo.Db.Elements[gref].(*base.Group).Class] = nil
		}
		aagmap := map[ResourceIndex]bool{}
		for cref := range cagmap {
			c := ttinfo.Db.Elements[cref].(*base.Class)
			aglist := g2tt[c.ClassGroup]
			//fmt.Printf("???????? %s: %+v\n", c.Tag, aglist)
			for _, agix := range aglist {
				aagmap[agix] = true
			}
		}
		// Handle groups
		for _, gref := range cinfo.Groups {
			for _, agix := range g2tt[gref] {
				// Check for repetitions
				if slices.Contains(resources, agix) {
					if !slices.Contains(warned, cinfo) {
						base.Warning.Printf(
							"Lesson with repeated atomic group"+
								" in Course: %s\n", ttinfo.View(cinfo))
						warned = append(warned, cinfo)
					}
				} else {
					resources = append(resources, agix)
					aagmap[agix] = false
				}
			}
		}
		extendedGroups := []ResourceIndex{}
		for agix, ok := range aagmap {
			if ok {
				extendedGroups = append(extendedGroups, agix)
			}
		}

		crooms := cinfo.Room.Rooms
		for _, rref := range crooms {
			// Only take the compulsory rooms here
			resources = append(resources, r2tt[rref])
		}
		a := ttinfo.Activities[aix]
		a.Resources = resources
		// Now add room-choices. Check against room choices in course.
		// Keep it simple, even though this will miss some errors.
		rchoices := map[base.Ref]bool{}
		for _, rlist := range cinfo.Room.RoomChoices {
			for _, rref := range rlist {
				rchoices[rref] = true
			}
		}
		nrooms := len(crooms) + len(cinfo.Room.RoomChoices)
		a.XRooms = make([]ResourceIndex, 0, nrooms)
		for _, rref := range ttl.Lesson.Rooms {
			if slices.Contains(crooms, rref) {
				continue
			}
			if rchoices[rref] {
				a.XRooms = append(a.XRooms, r2tt[rref])
			} else {
				base.Error.Printf("Room (%s) used for lesson of"+
					" course %s:\n  Room not specified for course.\n",
					ttinfo.Ref2Tag[rref], ttinfo.View(cinfo))
			}
		}
		if len(a.XRooms) > nrooms {
			base.Warning.Printf("Lesson in course %s uses more rooms"+
				" than specified for course.\n", ttinfo.View(cinfo))
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

		a.ExtendedGroups = extendedGroups
		//PossibleSlots: added later (see "makePossibleSlots"),
		//DifferentDays: ddlist, // only if not fixed, see below
		a.Parallel = plist
		if !a.Fixed {
			a.DifferentDays = ddlist
		}
		//--fmt.Printf("  ((%d)) %+v\n", aix, a.DifferentDays)
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
				// Check for end-of-day problems when duration > 1
				h := p % ttinfo.NHours
				if h+a.Duration > ttinfo.NHours {
					base.Error.Fatalf(
						"Placement for Fixed Activity %d @ %d invalid:\n"+
							"  -- %s\n",
						aix, p, ttinfo.View(a.CourseInfo))
					//a.XRooms = a.XRooms[:0]
				}
				if ttinfo.TestPlacement(aix, p) {
					// Perform placement
					ttinfo.PlaceActivity(aix, p)
					placed[aix] = true
					for _, paix := range a.Parallel {
						placed[paix] = true
					}
				} else {
					base.Error.Fatalf(
						"Placement of Fixed Activity %d @ %d failed:\n"+
							"  -- %s\n",
						aix, p, ttinfo.View(a.CourseInfo))
					//a.XRooms = a.XRooms[:0]
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
		if slices.Contains(a.PossibleSlots, p) &&
			ttinfo.TestPlacement(aix, p) {

			// Perform placement
			ttinfo.PlaceActivity(aix, p)
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
			a.XRooms = a.XRooms[:0]
		}
	}

	// Add room choices where possible.
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		a := ttinfo.Activities[aix]
		if len(a.XRooms) != 0 {
			var rnew []ResourceIndex
			p := a.Placement
			for _, rix := range a.XRooms {
				slot := rix*ttinfo.SlotsPerWeek + p
				if ttinfo.TtSlots[slot] == 0 {
					ttinfo.TtSlots[slot] = aix
				} else {
					base.Warning.Printf(
						"Lesson in course %s cannot use room %s\n",
						ttinfo.View(a.CourseInfo),
						ttinfo.Resources[rix].(*base.Room).Tag)
					rnew = append(rnew, rix)
				}
			}
			if len(rnew) != 0 {
				a.XRooms = a.XRooms[:len(rnew)]
				copy(a.XRooms, rnew)
			}
		}
	}
}

func (ttinfo *TtInfo) FindClashes(aix ActivityIndex, slot int) []ActivityIndex {
	// Return a list of activities (indexes) which are in conflict with
	// the proposed placement. It assumes the slot is in principle possible –
	// so that it will not, for example, be the last slot of a day if
	// the activity duration is 2.
	clashes := []ActivityIndex{}
	a := ttinfo.Activities[aix]
	day := slot / ttinfo.NHours
	//--fmt.Printf("????0 aix: %d slot %d\n", aix, slot)
	for _, addix := range a.DifferentDays {
		add := ttinfo.Activities[addix]
		if add.Placement >= 0 && add.Placement/ttinfo.NHours == day {
			clashes = append(clashes, addix)
			//--fmt.Printf("????1 %d\n", addix)
		}
	}
	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			c := ttinfo.TtSlots[i+ix]
			if c != 0 {
				//--xxx := ttinfo.Activities[c].Placement
				clashes = append(clashes, c)
				//--fmt.Printf("????2 %d %d r: %d p: %d\n", c, ix, rix, xxx)
			}
		}
	}
	for _, aixp := range a.Parallel {
		a := ttinfo.Activities[aixp]
		for _, addix := range a.DifferentDays {
			add := ttinfo.Activities[addix]
			if add.Placement >= 0 && add.Placement/ttinfo.NHours == day {
				clashes = append(clashes, addix)
				//--fmt.Printf("????3 %d\n", addix)
			}
		}
		for _, rix := range a.Resources {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				c := ttinfo.TtSlots[i+ix]
				if c != 0 {
					clashes = append(clashes, c)
					//--fmt.Printf("????4 %d %d\n", c, ix)
				}
			}
		}
	}
	slices.Sort(clashes)
	return slices.Compact(clashes)
}

// TODO: Can I safely assume that no attempt will be made to unplace fixed
// Activities?
func (ttinfo *TtInfo) UnplaceActivity(aix ActivityIndex) {
	a := ttinfo.Activities[aix]
	slot := a.Placement

	//TODO--- for testing
	if a.Fixed {
		base.Bug.Fatalf("Can't unplace %d – fixed\n", aix)
	}
	if slot < 0 {
		base.Bug.Printf("Can't unplace %d – not placed\n", aix)
		panic(1)
		return
	}

	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			ttinfo.TtSlots[i+ix] = 0
		}
	}
	for _, rix := range a.XRooms {
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
		for _, rix := range a.XRooms {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				ttinfo.TtSlots[i+ix] = 0
			}
		}
		a.Placement = -1
	}
	//--ttinfo.CheckResourceIntegrity()
}

// Note that – at present – testPlacement, findClashes and placeActivity
// don't try to place room choices. This is intentional, assuming that these
// will be placed by other functions ...

func (ttinfo *TtInfo) TestPlacement(aix ActivityIndex, slot int) bool {
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

func (ttinfo *TtInfo) PlaceActivity(aix ActivityIndex, slot int) {
	// Allocate the resources, assuming none of the slots are blocked!
	//--fmt.Printf("++++++++ PLACE ++++++++ %d: %d\n", aix, slot)
	a := ttinfo.Activities[aix]

	//TODO-- This is for debugging
	p := a.Placement
	if p >= 0 && p != slot {
		fmt.Printf("::::: %+v\n", a)
		panic(fmt.Sprintf("Activity %d already placed: %d\n", aix, p))
	}
	//

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
	//--ttinfo.CheckResourceIntegrity()
}

// DEBUGGING only
func (ttinfo *TtInfo) CheckResourceIntegrity() {
	for rix := 0; rix < len(ttinfo.Resources); rix++ {
		slot0 := rix * ttinfo.SlotsPerWeek
		for p := 0; p < ttinfo.SlotsPerWeek; p++ {
			aix := ttinfo.TtSlots[slot0+p]
			if aix <= 0 {
				continue
			}
			a := ttinfo.Activities[aix]
			ap := a.Placement
			if ap < 0 {
				panic(fmt.Sprintf("Resource (%d) of unplaced Activity (%d)"+
					" at position %d:", rix, aix, ap))
			}
			for i := 0; i < a.Duration; i++ {
				if ap+i == p {
					goto pok
				}
			}
			panic(fmt.Sprintf("Resource (%d) of Activity (%d)"+
				" at wrong position %d (should be %d):",
				rix, aix, p, ap))
		pok:
		}
	}
}
