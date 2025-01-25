package ttbase

import (
	"W365toFET/base"
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
	//TODO: Do I really need this? It might rather confuse things
	ExtendedGroups []ResourceIndex
	// Fixed specifies whether the activity must remain in its current slot
	Fixed bool
	// Placement specifies the time-slot in which this activity has been
	// placed (day * daylength + hour), or -1 if unplaced
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
// (fixed, and also non-fixed with a placement). In this way various errors
// can be checked for.
func (ttinfo *TtInfo) addActivityInfo() {
	//g2tt := ttinfo.AtomicGroupIndexes
	// Collect non-fixed activities which need placing
	toplace := []ActivityIndex{}

	//TODO??
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

	/*TODO--
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
	*/

	/////////+
	ttinfo.collectCourseResources()

	ttinfo.PrepareActivityGroups()

	//TODO: Consider the order of courses (should be same as in source)
	// and Lessons/Activities (should always be the same). Perhaps I would
	// need a course list to ensure this.

	// Complete the Activity items for each course
	r2tt := ttinfo.RoomIndexes
	for _, cinfo := range ttinfo.LessonCourses {
		// Get the room-choices. Check against room choices in course.

		//TODO: This needs testing with data that provides choice allocations.

		crooms := cinfo.Room.Rooms       // "necessary" rooms
		xrooms := cinfo.Room.RoomChoices // list of room choices
		//fmt.Printf("++ COURSE: %s\n", ttinfo.View(cinfo))
		for _, aix := range cinfo.Lessons {
			a := ttinfo.Activities[aix]
			// Check each actually allocated room
			rlist := []Ref{}
			for _, rref := range a.Lesson.Rooms {
				if slices.Contains(crooms, rref) {
					// Ignore "necessary" rooms
					continue
				}
				rlist = append(rlist, rref)
			}
			var xr []Ref
			if len(rlist) != 0 {
				// Try to match rooms in rlist to the choice list, xrooms
				xr = rclfunc(xrooms, rlist)
				if xr == nil {
					base.Error.Printf("Rooms (%s) used for lesson of"+
						" course %s:\n  rooms don't match.\n",
						ttinfo.pResources(rlist), ttinfo.View(cinfo))
				}
			}
			a.XRooms = make([]ResourceIndex, len(xrooms))
			var ri ResourceIndex
			for i := 0; i < len(xrooms); i++ {
				ri = -1
				if i < len(xr) {
					r := xr[i]
					if r != "" {
						ri = r2tt[r]
					}
				}
				a.XRooms[i] = ri
			}
			//fmt.Printf("  -- %+v\n", a.XRooms)
		}

		//

	}

	/////////-

	//TODO: How much of the following is still needed, in this or another
	// form. It should probably be in the above course loop

	//TODO: Different days should now include all days-between and be
	// specified on the activity groups as their possible slots (with weight)

	// Lessons (Activities) start at index 1!
	for aix := 1; aix < len(ttinfo.Activities); aix++ {

		a := ttinfo.Activities[aix]

		// Sort and compactify different-days activities
		ddlist, ok := diffdays[aix]
		if ok && len(ddlist) > 1 {
			slices.Sort(ddlist)
			ddlist = slices.Compact(ddlist)
		}

		/*TODO-- Sort and compactify parallel activities
		plist, ok := parallels[aix]
		if ok && len(plist) > 1 {
			slices.Sort(plist)
			plist = slices.Compact(plist)
		}
		a.Parallel = plist
		*/

		//TODO? a.ExtendedGroups = extendedGroups

		//PossibleSlots: added later (see "makePossibleSlots"),
		//DifferentDays: ddlist, // only if not fixed, see below
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
				h := p % ttinfo.DayLength
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
				if rix < 0 {
					continue
				}
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

// rclfunc uses recursion to match the rooms to the room-choice list,
// using "" where a choice has no room allocated
func rclfunc(rclist [][]Ref, rlist []Ref) []Ref {
	if len(rclist) == 0 {
		if len(rlist) == 0 {
			return []Ref{}
		} else {
			return nil
		}
	}
	rc := rclist[0]
	rlx := make([]Ref, len(rlist)-1)
	for i, r := range rlist {
		if slices.Contains(rc, r) {
			// Remove the room from the list
			rlx = rlx[:i]
			copy(rlx, rlist)
			rlx = append(rlx, rlist[i+1:]...)
			rl := rclfunc(rclist[1:], rlx)
			if rl != nil {
				return append([]Ref{r}, rl...)
			}
		}
	}
	// Assume there was no room supplied for this choice
	rl := rclfunc(rclist, rlx)
	if rl != nil {
		return append([]Ref{""}, rl...)
	}
	return nil
}

// collectCourseResources collects resources for each course
func (ttinfo *TtInfo) collectCourseResources() {
	g2tt := ttinfo.AtomicGroupIndexes
	t2tt := ttinfo.TeacherIndexes
	r2tt := ttinfo.RoomIndexes
	for _, cinfo := range ttinfo.CourseInfo {
		resources := []ResourceIndex{}

		for _, tref := range cinfo.Teachers {
			resources = append(resources, t2tt[tref])
		}

		//TODO: Is this really useful?
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
		//--?

		// Handle groups
		for _, gref := range cinfo.Groups {
			for _, agix := range g2tt[gref] {
				// Check for repetitions
				if slices.Contains(resources, agix) {
					base.Warning.Printf(
						"Lesson with repeated atomic group"+
							" in Course: %s\n", ttinfo.View(cinfo))
				} else {
					resources = append(resources, agix)
					aagmap[agix] = false
				}
			}
		}

		//TODO: What, if anything, to do with this?
		extendedGroups := []ResourceIndex{}
		for agix, ok := range aagmap {
			if ok {
				extendedGroups = append(extendedGroups, agix)
			}
		}

		//TODO--
		//fmt.Printf("COURSE: %s\n", ttinfo.View(cinfo))

		crooms := cinfo.Room.Rooms
		for _, rref := range crooms {
			// Only take the compulsory rooms here
			resources = append(resources, r2tt[rref])
		}
		cinfo.Resources = resources
	}
}
