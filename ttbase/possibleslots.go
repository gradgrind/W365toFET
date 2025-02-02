package ttbase

import (
	"W365toFET/base"
)

// makePossibleSlots tests all slots for all (non-fixed) lessons. It should
// be called just after the fixed lessons have been placed. Each [TtLesson]
// gets a list of potentially available slots.
func (ttinfo *TtInfo) makePossibleSlots() {
	ttplaces := ttinfo.Placements

	for _, ag := range ttplaces.ActivityGroups {
		dmap := map[int][]int{} // remember per-duration possible-slots
		for _, lix := range ag.LessonUnits {
			l := ttplaces.TtLessons[lix]
			if l.Fixed {
				// Fixed lessons don't need possible slots.
				continue
			}
			plist, ok := dmap[l.Duration]
			if !ok {
				hlimit := ttinfo.NHours - l.Duration
				for d := 0; d < ttinfo.NDays; d++ {
					dayslot0 := d * ttinfo.DayLength
					for h := 0; h <= hlimit; h++ {
						p := dayslot0 + h
						if ttinfo.TestPlacement(lix, p) {
							plist = append(plist, p)
						}
					}
				}
				if len(plist) == 0 {
					base.Error.Fatalf("Lesson %d has no available time"+
						" slots\n  -- Course: %s\n",
						lix, ttinfo.printAGCourse(ag))
				}
				dmap[l.Duration] = plist
			}
			l.PossibleSlots = plist
		}
	}
}
