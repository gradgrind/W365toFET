package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
	"slices"
)

func (pmon *placementMonitor) furtherPlacements(
	alist []ttbase.ActivityIndex,
) []ttbase.ActivityIndex {
	ttinfo := pmon.ttinfo

	//TODO--
	aslist := make([]ttbase.ActivityIndex, len(alist))
	copy(aslist, alist)
	slices.Sort(aslist)
	fmt.Printf(" furtherPlacements - aslist: %+v\n", aslist)
	//

	var failed []ttbase.ActivityIndex
	for {
		// Place activities from alist so long as possible without
		// making the score worse.

	newstate:
		// Randomize list of activities
		na := len(alist)
		pending := make([]int, 0, na)
		if na > 1 {
			for _, i := range rand.Perm(na) {
				pending = append(pending, alist[i])
			}
		} else {
			pending = append(pending, alist...)
		}

		// Save state
		score0 := pmon.score
		state0 := pmon.saveState()

		// Try each of the activities in turn.
		for _, aix := range pending {
			//n := len(pending)
			//dpn := 0

			// Decide which slot to use
			a := ttinfo.Activities[aix]

			// Randomize list of slots
			nslots := len(a.PossibleSlots)
			var slots []int = nil
			if nslots > 1 {
				slots = make([]int, nslots)
				for i, j := range rand.Perm(nslots) {
					slots[i] = a.PossibleSlots[j]
				}
			} else {
				slots = a.PossibleSlots
			}

			// The call to basicPlaceActivities needs the updated list of
			// unplaced activities. First remove the one to be placed.
			aixlist := []ttbase.ActivityIndex{}
			for _, aixp := range pending {
				if aixp != aix {
					aixlist = append(aixlist, aixp)
				}
			}
			aixlist0len := len(aixlist)

			for _, slot := range slots {
				// Remove added elements from aixlist
				aixlist = aixlist[:aixlist0len]

				// Prepare for the placement by removing clashing activities.
				clashes := ttinfo.FindClashes(aix, slot)
				for _, aixc := range clashes {
					ttinfo.UnplaceActivity(aixc)
					aixlist = append(aixlist, aixc)
				}

				//TODO--
				fmt.Printf("§CLASHES (%d): %+v\n", aix, clashes)
				//

				ttinfo.PlaceActivity(aix, slot)

				// Adjust the score by recalculating the resources for
				// aix and clashes.
				rmap := map[ttbase.ResourceIndex]bool{}
				dp := 0
				for _, r := range a.Resources {
					rmap[r] = true
					rp := pmon.resourcePenalty1(r)
					rp0 := pmon.resourcePenalties[r]
					if rp != rp0 {
						dp += rp - rp0
						pmon.resourcePenalties[r] = rp
					}
				}
				for _, aixc := range clashes {
					ap := ttinfo.Activities[aixc]
					for _, r := range ap.Resources {
						if rmap[r] {
							continue
						}
						rmap[r] = true
						rp := pmon.resourcePenalty1(r)
						rp0 := pmon.resourcePenalties[r]
						if rp != rp0 {
							dp += rp - rp0
							pmon.resourcePenalties[r] = rp
						}
					}
				}
				pmon.score += dp

				failed = pmon.basicPlaceActivities(aixlist)

				if pmon.score <= score0 && len(failed) < len(alist) {
					//n = len(failed)
					//dpn = dp
					alist = failed

					// TODO--
					fmt.Printf(" +++ %d: %d $%d\n", aix, len(alist), dp)
					aslist := make([]ttbase.ActivityIndex, len(alist))
					copy(aslist, alist)
					slices.Sort(aslist)
					fmt.Printf(" alist: %+v\n", aslist)
					//

					goto newstate // Accept this result
				}

				//fmt.Printf(" +++ %d: %d\n", aix, len(failed))
				pmon.restoreState(state0)

				//TODO-- (debugging only)
				if pmon.score != score0 {
					panic("Score mucked up")
				}
			}
			//fmt.Printf(" +++ %d: %d $%d\n", aix, n, dpn)
		} // end of pending activity loop

		// If this point is reached normally (all activities tested),
		// no improvement was found. That doesn't mean that no improvement
		// is possible ...

		//TODO: This is very experimental!
		if len(failed) < len(alist) {
			alist = failed

			//TODO--
			fmt.Printf(" §REPEAT %+v\n", alist)

		} else {
			//TODO: Is this unnecessary (has it been restored already)?
			pmon.restoreState(state0)
			break
		}
	}
	return failed
}
