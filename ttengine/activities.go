package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"slices"
)

type SlotIndex int
type ResourceIndex = ttbase.ResourceIndex

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
			p = l.Day*tt.NDays + l.Hour
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
	}
}
