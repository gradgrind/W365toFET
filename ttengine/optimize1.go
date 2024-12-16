package ttengine

import (
	"W365toFET/base"
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
	//fmt.Printf(" furtherPlacements - aslist: %+v\n", aslist)
	//
	na0 := len(alist)
	breakout := false

	var failed []ttbase.ActivityIndex
	for {
		// Place activities from alist so long as possible without
		// making the score worse.

	newstate:
		na := len(alist)

		//TODO-- (debugging)
		// Check integrity of activity placements
		acount := 0
		for aix := 1; aix < len(ttinfo.Activities); aix++ {
			a := ttinfo.Activities[aix]
			p := a.Placement
			if p < 0 {
				acount++
				continue
			}
			// Check the resource allocation
			for _, r := range a.Resources {
				slot := r*ttinfo.SlotsPerWeek + p
				for i := 0; i < a.Duration; i++ {
					raix := ttinfo.TtSlots[slot+i]
					if raix != aix {
						base.Bug.Fatalf(
							"!!! Corrupt resource allocation, aix: %d (%d)\n",
							aix, r)
					}
				}
			}
		}
		//fmt.Printf("$Missing: %d %d\n", na, acount)
		if acount != na {
			base.Bug.Fatalf("!!! in alist: %d unplaced: %d\n", na, acount)
		}
		//--

		// Randomize list of activities
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
				//fmt.Printf("§CLASHES (%d): %+v\n", aix, clashes)
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

				if len(failed) < len(alist) {
					if pmon.score > score0 {
						if !breakout {
							goto restore
						}
						breakout = false
					}
					//n = len(failed)
					//dpn = dp
					alist = failed

					// TODO--
					fmt.Printf(" +++ %d: %d $%d\n", aix, len(alist), pmon.score)
					aslist := make([]ttbase.ActivityIndex, len(alist))
					copy(aslist, alist)
					slices.Sort(aslist)
					//fmt.Printf(" alist: %+v\n", aslist)
					//

					goto newstate // Accept this result

				}
			restore:

				//fmt.Printf(" +++ %d: %d\n", aix, len(failed))
				pmon.restoreState(state0)

				//TODO-- (debugging only)
				if pmon.score != score0 {
					panic("Score mucked up")
				}
			} // end of slot loop

			// Activity not placed

			//fmt.Printf(" +++ %d: %d $%d\n", aix, n, dpn)
		} // end of pending activity loop

		// If this point is reached normally (all activities tested),
		// no improvement was found. That doesn't mean that no improvement
		// is possible ...

		fmt.Printf(" * All Activites tested, remaining: %d -> %d\n",
			na0, len(pending))

		//

		//TODO: the question is, how might the result be improved?
		// Perhaps by placing one of the remaining activities in spite of
		// a worsening score. Then optimize and see if – in the end – an
		// improvement was made. If not revert to old state and try again.

		//TODO: More attention is needed to the score!

		breakout = true
		alist = pending
		if na0 == len(pending) {
			break
		}
		na0 = len(pending)
	}
	for _, aix := range alist {
		a := ttinfo.Activities[aix]
		fmt.Printf("==> %s\n", ttinfo.View(a.CourseInfo))
	}
	return alist
}
