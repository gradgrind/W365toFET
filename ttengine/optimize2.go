package ttengine

/*

import (
	"W365toFET/ttbase"
	"fmt"
)

func (pmon *placementMonitor) furtherPlacements2(
	alist []ttbase.ActivityIndex,
) []ttbase.ActivityIndex {
	//ttinfo := pmon.ttinfo

	//TODO--
	//aslist := make([]ttbase.ActivityIndex, len(alist))
	//copy(aslist, alist)
	//slices.Sort(aslist)
	//fmt.Printf(" furtherPlacements - aslist: %+v\n", aslist)
	//

	fmt.Printf(" furtherPlacements - len(alist): %d\n", len(alist))

	pending := pmon.optimize2(alist)

	score1 := pmon.score
	fmt.Printf(" * All Activites tested, remaining: %d (%d)\n",
		len(pending), score1)

	//

	//TODO: the question is, how might the result be improved?
	// Perhaps by placing one of the remaining activities in spite of
	// a worsening score. Then optimize and see if – in the end – an
	// improvement was made. If not revert to old state and try again.

	//TODO: Probably need to give each "branch" more time to find a
	// better solution ... recurse? Or more radical breakouts?
	// or worry less about penalties than getting the activities placed?
	/*
		// Save state
		state1 := pmon.saveState()

		ocount := 0
		for {
			pending1 := pmon.optimize2(pending, true)

			if pmon.score >= score1 {
				//fmt.Printf(" * --- remaining: %d (%d)\n",
				//	len(pending1), pmon.score)
				pmon.restoreState(state1)
			} else {
				fmt.Printf(" * +++ remaining: %d (%d)\n",
					len(pending1), pmon.score)
				state1 = pmon.saveState()
				score1 = pmon.score
				pending = pending1
			}
			ocount++
			if ocount%100 == 0 {
				fmt.Printf("@ %d\n", ocount)
				if ocount/100 == 10 {
					break
				}
			}
		}
	////
	//TODO--
	fmt.Printf("\n * Return: %d (%d)\n", len(pending), score1)
	for _, aix := range pending {
		a := pmon.ttinfo.Activities[aix]
		fmt.Printf("==> %s\n", pmon.ttinfo.View(a.CourseInfo))
	}
	//--
	return pending
}

func (pmon *placementMonitor) optimize2(
	alist []ttbase.ActivityIndex,
) []ttbase.ActivityIndex {
	// Try to place the remaining activities.

	ttinfo := pmon.ttinfo
	var failed []ttbase.ActivityIndex
	for {

	newstate:
		// Save state
		score0 := pmon.score
		state0 := pmon.saveState()

		// Try each of the activities in turn.
		for _, aix := range alist {
			// Decide which slot to use
			a := ttinfo.Activities[aix]

			// The call to basicPlaceActivities needs the updated list of
			// unplaced activities. First remove the one to be placed.
			aixlist := []ttbase.ActivityIndex{}
			for _, aixp := range alist {
				if aixp != aix {
					aixlist = append(aixlist, aixp)
				}
			}
			aixlist0len := len(aixlist)

			for _, slot := range a.PossibleSlots {
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

				failed = pmon.basicPlaceActivities2(aixlist)

				if len(failed) < len(alist) {
					alist = failed

					// TODO--
					//fmt.Printf(" +++ %d: %d $%d\n", aix, len(alist), pmon.score)
					//aslist := make([]ttbase.ActivityIndex, len(alist))
					//copy(aslist, alist)
					//slices.Sort(aslist)
					//fmt.Printf(" alist: %+v\n", aslist)
					//

					goto newstate // Accept this result

				}

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
		return alist
	}
}
*/
