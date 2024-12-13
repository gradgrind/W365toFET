package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
)

// Try a gap-aware approach with backtracking.

func PlaceLessons(ttinfo *ttbase.TtInfo, alist []ttbase.ActivityIndex) {
	npending := len(alist)
	pending := []ttbase.ActivityIndex{}
	for {
		for count, aix := range alist {
			//TODO-- This counter is only for debugging
			if count == 1000 {
				return
			}

			// Place the activity in one of the available slots.
			// Choose a slot such that no additional gap arises.
			// If there is no suitable slot, add to pending list.

			// First get the atomic groups for this activity
			a := ttinfo.Activities[aix]
			aglist := []ttbase.ResourceIndex{}
			for _, agix := range a.Resources {
				if agix < ttinfo.NAtomicGroups {
					aglist = append(aglist, agix)
				}
			}
			// Get free slots for this activity
			slots := []int{}
			for _, slot := range a.PossibleSlots {
				// Filter to include only slots which don't create gaps for a
				// group of students.
				h := slot % ttinfo.NHours
				if h == 0 {
					if ttinfo.TestPlacement(aix, slot) {
						slots = append(slots, slot)
					}
					continue
				}
				d := slot / ttinfo.NHours
				end := (d + 1) * ttinfo.NHours

				// For each atomic group
				for _, agix := range aglist {
					// Fail if the slot before is empty AND there are only
					// empty slots afterwards on that day.

					//TODO: Consider lunch beaks ... if the previous slot is
					// at lunch time, check the slot before that, too.

					slot0 := agix * ttinfo.SlotsPerWeek
					if ttinfo.TtSlots[slot0+slot-1] == 0 {
						// Check subsequent slots

						for sl := slot + 1; sl < end; sl++ {
							if ttinfo.TtSlots[slot0+sl] > 0 {
								// Reducing an existing gap
								goto slotok
							}
						}
						// Not OK, goto next slot
						goto slotfail
					}
				slotok:
					// OK, check next ag
				}
				if ttinfo.TestPlacement(aix, slot) {
					slots = append(slots, slot)
				}
			slotfail:
			}

			if len(slots) == 0 {
				// There are currently no suitable free slots for this
				// activity, add it to the pending list.
				pending = append(pending, aix)
				continue
			}

			var slot int
			if len(slots) == 1 {
				slot = slots[0]
			} else {
				slot = slots[rand.IntN(len(slots))]
			}

			/*
				cinfo := a.CourseInfo
				fmt.Printf(" §PLACE @%d.%d %s -- %+v\n",
					slot/ttinfo.NHours,
					slot%ttinfo.NHours,
					ttinfo.View(cinfo),
					a.DifferentDays,
				)
			*/

			ttinfo.PlaceActivity(aix, slot)
		} // end of activity loop
		// Repeat with initially rejected activities
		np := len(pending)
		if np == 0 {
			break
		}
		if np == npending {
			break
		}
		npending = np
		alist = pending
		pending = nil
		fmt.Printf("  --- Pending: %d\n", len(alist))
	}
	fmt.Printf("$$$ Unplaced: %d\n", len(pending))

	//slices.Reverse(failed)
	//l0 := len(failed)
	//fmt.Printf("Remaining: %d\n", l0)

}
