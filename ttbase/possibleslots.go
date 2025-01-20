package ttbase

import (
	"W365toFET/base"
)

// By trying all slots for all (non-fixed) activities just after the fixed
// activities have been placed, each activity can get a list of potentially
// available slots.
func (ttinfo *TtInfo) makePossibleSlots() {
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		a := ttinfo.Activities[aix]
		if a.Fixed {
			// Fixed activities don't need possible slots.
			continue
		}
		plist := []int{}
		hlimit := ttinfo.NHours - a.Duration
		for d := 0; d < ttinfo.NDays; d++ {
			dayslot0 := d * ttinfo.DayLength
			for h := 0; h <= hlimit; h++ {
				p := dayslot0 + h
				if ttinfo.TestPlacement(aix, p) {
					plist = append(plist, p)
				}
			}
		}
		if len(plist) == 0 {
			base.Error.Fatalf("Activity %d has no available time slots\n"+
				"  -- Course: %s\n",
				aix, ttinfo.View(a.CourseInfo))
		}
		a.PossibleSlots = plist
	}
}
