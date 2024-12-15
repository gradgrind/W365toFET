package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
	"time"
)

// Use a penalty-weighting approach.

func PlaceLessons2(ttinfo *ttbase.TtInfo, alist []ttbase.ActivityIndex) {
	//resourceSlotActivityMap := makeResourceSlotActivityMap(ttinfo)
	var pmon placementMonitor
	{
		//var delta int64 = 7 // This might be a reasonable value?
		pmon = placementMonitor{
			//count:                   delta,
			//delta:                   delta,
			//added:                   make([]int64, len(ttinfo.Activities)),
			ttinfo: ttinfo,
			//preferEarlier:           preferEarlier,
			//preferLater:             preferLater,
			//resourceSlotActivityMap: resourceSlotActivityMap,
			resourcePenalties: make([]int, len(ttinfo.Resources)),
			score:             0,
			pendingPenalties:  []resourcePenalty{},
		}
	}
	pmon.initConstraintData()

	// Calculate initial stage 1 penalties
	for r := 0; r < len(ttinfo.Resources); r++ {
		p := pmon.resourcePenalty1(r)
		pmon.resourcePenalties[r] = p
		pmon.score += p
		//fmt.Printf("$ PENALTY %d: %d\n", r, p)
	}

	//TODO--
	tstart := time.Now()
	//

	pending := pmon.basicPlaceActivities(alist)

	//TODO--
	elapsed := time.Since(tstart)
	fmt.Printf("Outer Loop took %s\n", elapsed)
	//

	fmt.Printf("$$$ Unplaced: %d\n", len(pending))

	//slices.Reverse(failed)
	//l0 := len(failed)
	//fmt.Printf("Remaining: %d\n", l0)

}

func (pmon *placementMonitor) basicPlaceActivities(
	alist []ttbase.ActivityIndex,
) []ttbase.ActivityIndex {
	// Place the activities as far as is possible without increasing the
	// penalty and without displacing already placed activities.

	ttinfo := pmon.ttinfo
	npending := len(alist)
	pending := []ttbase.ActivityIndex{}

	for {
		for _, aix := range alist {
			// Place the activity in one of the available slots.
			// Choose a slot such that no additional penalty arises.
			// If there is no suitable slot, add to pending list.

			// First get the atomic groups for this activity
			// TODO: This could be pre-calculated.
			a := ttinfo.Activities[aix]
			aglist := []ttbase.ResourceIndex{}
			for _, agix := range a.Resources {
				if agix < ttinfo.NAtomicGroups {
					aglist = append(aglist, agix)
				}
			}

			// Get free slots for this activity
			nslots := []int{}
			pslots := []int{} // "priority" slots
			for _, slot := range a.PossibleSlots {
				if !ttinfo.TestPlacement(aix, slot) {
					continue
				}

				//+++
				// Prefer a slot with something parallel in the class.
				for _, agix := range a.ExtendedGroups {
					if ttinfo.TtSlots[agix*ttinfo.SlotsPerWeek+slot] != 0 {
						// There is a parallel lesson
						pslots = append(pslots, slot)
						goto nextslot
					}
				}
				//++-

				nslots = append(nslots, slot)
			nextslot:
			} // end of slot loop

			nl := len(nslots)
			pl := len(pslots)

			//fmt.Printf("§FROM %+v / %+v\n", pslots, slots)
			if pl == 0 && nl == 0 {
				// There are currently no suitable free slots for this
				// activity, so add it to the pending list.
				goto notplaced
			}

			{
				slots := make([]int, 0, nl+pl)
				if len(pslots) > 1 {
					for _, i := range rand.Perm(pl) {
						slots = append(slots, pslots[i])
					}
				} else {
					slots = append(slots, pslots...)
				}
				if len(nslots) > 1 {
					for _, i := range rand.Perm(nl) {
						slots = append(slots, nslots[i])
					}
				} else {
					slots = append(slots, nslots...)
				}

				for _, slot := range slots {
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
					// Evaluate
					dp := pmon.evaluate1(aix)
					if dp > 0 {
						// Reject the change
						ttinfo.UnplaceActivity(aix)
						continue // -> next slot
					}
					// Accept the change
					for _, pp := range pmon.pendingPenalties {
						pmon.resourcePenalties[pp.resource] = pp.penalty
					}
					pmon.score += dp
					goto adone
				}
			} // end of slot-place loop
		notplaced: // activity not placed
			pending = append(pending, aix)
		adone:
		} // end of activity loop

		// Repeat with initially rejected activities
		np := len(pending)
		if np == 0 {
			break
		}
		if np == npending {
			// got stuck
			break
		}
		npending = np
		alist = pending
		pending = nil
		fmt.Printf("  --- Pending: %d\n", len(alist))
	}
	return pending
}
