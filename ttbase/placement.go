package ttbase

import (
	"W365toFET/base"
	"fmt"
	"slices"
)

// initialFixedPlacement places the fixed lessons.
func (ttinfo *TtInfo) initialFixedPlacement() {
	ttplaces := ttinfo.Placements
	for _, ag := range ttplaces.ActivityGroups {
		for _, lix := range ag.LessonUnits {
			l := ttplaces.TtLessons[lix]
			if !l.Fixed {
				continue
			}

			p := l.Placement
			if p < 0 {
				base.Error.Fatalf("Fixed lesson has no placement,"+
					" course: %s\n", ttinfo.printAGCourse(ag))
			}

			if ttinfo.TestPlacement(lix, p) {
				// Perform placement
				ttinfo.PlaceLesson(lix, p)
			} else {
				t := fmt.Sprintf("%d.%d",
					p/ttinfo.DayLength, p%ttinfo.DayLength)
				base.Error.Fatalf(
					"Placement of fixed lesson @ %s failed,\n"+
						"  course: %s\n",
					t, ttinfo.printAGCourse(ag))
			}
		}
	}
}

// initialNonFixedPlacement places those non-fixed lessons which have
// placements.
func (ttinfo *TtInfo) initialNonFixedPlacement() {
	ttplaces := ttinfo.Placements
	for _, ag := range ttplaces.ActivityGroups {
		for _, lix := range ag.LessonUnits {
			l := ttplaces.TtLessons[lix]
			if l.Fixed {
				continue
			}
			if l.Placement < 0 {
				// Ensure that the room choices are empty
				for i := range l.XRooms {
					l.XRooms[i] = -1
				}
				continue
			}

			p := l.Placement
			if slices.Contains(l.PossibleSlots, p) &&
				ttinfo.TestPlacement(lix, p) {

				// Perform placement
				ttinfo.PlaceLesson(lix, p)
			} else {
				t := fmt.Sprintf("%d.%d",
					p/ttinfo.DayLength, p%ttinfo.DayLength)
				base.Error.Printf(
					"Placement of non-fixed lesson @ %s failed,\n"+
						"  course: %s\n",
					t, ttinfo.printAGCourse(ag))
				l.Placement = -1
				for i := range l.XRooms {
					l.XRooms[i] = -1
				}
			}
		}
	}
}

// initialRoomChoices tries to allocate the rooms from choice lists as
// supplied with the input data. If XRooms is incomplete, the missing
// room indexes are represented by -1.
func (ttinfo *TtInfo) initialRoomChoices() {
	ttplaces := ttinfo.Placements
	for _, ag := range ttplaces.ActivityGroups {
		for _, lix := range ag.LessonUnits {
			l := ttplaces.TtLessons[lix]
			p := l.Placement
			if p < 0 {
				continue
			}
			for i, rix := range l.XRooms {
				if rix < 0 {
					continue
				}
				slot := rix*ttinfo.SlotsPerWeek + p
				if ttplaces.TtSlots[slot] == 0 {
					ttplaces.TtSlots[slot] = lix
				} else {
					base.Warning.Printf(
						"Lesson cannot use room %s,\n  course: %s\n",
						ttinfo.Resources[rix].(*base.Room).Tag,
						ttinfo.printAGCourse(ag),
					)
					l.XRooms[i] = -1
				}
			}
		}
	}
}

// UnplaceLesson displaces a lesson from the slot in which it had been
// placed, deallocating its resources.
// TODO: Can I safely assume that no attempt will be made to unplace fixed
// lessons?
func (ttinfo *TtInfo) UnplaceLesson(lix LessonUnitIndex) {
	ttplaces := ttinfo.Placements
	l := ttplaces.TtLessons[lix]
	slot := l.Placement

	//TODO--- for testing
	if l.Fixed {
		base.Bug.Fatalf("Can't unplace %d – fixed\n", lix)
	}
	if slot < 0 {
		base.Bug.Fatalf("Can't unplace %d – not placed\n", lix)
	}
	//--

	for _, rix := range l.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < l.Duration; ix++ {
			ttplaces.TtSlots[i+ix] = 0
		}
	}
	for _, rix := range l.XRooms {
		if rix < 0 {
			continue
		}
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < l.Duration; ix++ {
			ttplaces.TtSlots[i+ix] = 0
		}
	}
	l.Placement = -1
	//--ttinfo.CheckResourceIntegrity()
}

// Note that – at present – testPlacement, findClashes and placeLesson
// don't try to place room choices. This is intentional, assuming that these
// will be placed by other functions ...

// TestPlacement is a simple boolean placement test. It assumes the slot is
// possible for the lesson – so that it will not, for example, be the last
// slot of a day if the activity duration is 2.
func (ttinfo *TtInfo) TestPlacement(lix LessonUnitIndex, slot int) bool {
	ttplaces := ttinfo.Placements
	l := ttplaces.TtLessons[lix]

	// Check for resource collisions
	for _, rix := range l.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < l.Duration; ix++ {
			if ttplaces.TtSlots[i+ix] != 0 {
				return false
			}
		}
	}

	// Check for days-between conflicts
	for _, dbc := range l.DaysBetween {
		mingap := dbc.HoursGap
		for _, xlix := range dbc.LessonUnits {
			xl := ttplaces.TtLessons[xlix]
			xp := xl.Placement
			if xp < 1 {
				continue
			}
			gap := slot - xp
			if gap < 0 {
				if -gap >= mingap {
					continue
				}
			} else if gap >= mingap {
				continue
			}
			// The gap is too small.
			if dbc.Weight == base.MAXWEIGHT {
				return false
			}

			// For soft constraints with ConsecutiveIfSameDay true, the
			// lessons must be adjacent if they are on the same day.

			if dbc.ConsecutiveIfSameDay {
				if gap < 0 {
					if -gap < ttinfo.NHours {
						// The slot is before xp on the same day
						if l.Duration+slot != xp {
							return false
						}
					}
				} else if gap < ttinfo.NHours {
					// The slot is after xp on the same day
					if xl.Duration+xp != slot {
						return false
					}
				}
			}

			// Whether this slot is acceptable depends on a random number
			// and the weight.
			if !AcceptRandom(dbc.Weight) {
				return false
			}
		}
	}

	return true
}

// PlaceLesson places a lesson in a given slot, allocating the resources.
// It assumes none of the slots are blocked, i.e. that the validity of the
// placement has been checked already.
func (ttinfo *TtInfo) PlaceLesson(lix LessonUnitIndex, slot int) {
	//--fmt.Printf("++++++++ PLACE ++++++++ %d: %d\n", aix, slot)
	ttplaces := ttinfo.Placements
	l := ttplaces.TtLessons[lix]

	//TODO-- This is for debugging
	p := l.Placement
	if p >= 0 && p != slot {
		fmt.Printf("::::: %+v\n", l)
		panic(fmt.Sprintf("Lesson %d already placed: %d\n", lix, p))
	}
	//

	for _, rix := range l.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < l.Duration; ix++ {
			ttplaces.TtSlots[i+ix] = lix
		}
	}
	l.Placement = slot
	//--ttinfo.CheckResourceIntegrity()
}

func (ttinfo *TtInfo) FindClashes(
	lix LessonUnitIndex,
	slot int,
) []LessonUnitIndex {
	// Return a list of lessons (indexes) which are in conflict with
	// the proposed placement. It assumes the slot is in principle possible –
	// so that it will not, for example, be the last slot of a day if
	// the activity duration is 2.
	clashes := []LessonUnitIndex{}
	ttplaces := ttinfo.Placements
	l := ttplaces.TtLessons[lix]

	for _, rix := range l.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < l.Duration; ix++ {
			c := ttplaces.TtSlots[i+ix]
			if c != 0 {
				clashes = append(clashes, c)
			}
		}
	}

	//TODO: Addapt different-days handling from above

	slices.Sort(clashes)
	return slices.Compact(clashes)
}
