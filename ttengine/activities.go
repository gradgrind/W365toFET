package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"slices"
)

type SlotIndex = int
type ResourceIndex = int

type Activity struct {
	Index         int
	Duration      int
	Resources     []ResourceIndex
	PossibleSlots []SlotIndex
	Fixed         bool
	Placement     int // day * nhours + hour, or -1 if unplaced
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
	for i, ttl := range ttinfo.TtLessons {
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

		tt.Activities[i+1] = &Activity{
			Index:     i + 1,
			Duration:  l.Duration,
			Resources: resources,
			//PossibleSlots: TODO,
			Fixed:     l.Fixed,
			Placement: p,
		}

		//TODO: Actually need to test before setting the Placement field!
		// Perhaps the placement could be revoked on failure? It seems useful
		// to have the desired data in the activities.

		// First place the fixed lessons, then build the PossibleSlots for
		// non-fixed lessons. Should Resource NotAvailable data be fed in
		// before placing fixed lessons? Almost certainly, yes!

		//TODO

		if p >= 0 {
			if tt.testPlacement(i+1, p) {
				// Perform placement
				tt.placeActivity(i+1, p)
			} else {
				base.Error.Printf(
					"Placement of Activity %d @ %d failed:\n  -- %s\n",
					i+1, p, ttinfo.View(cinfo))
			}
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
}
