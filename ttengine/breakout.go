package ttengine

import (
	"math/rand/v2"
)

// TODO: Is there an optimal limit? Too small and it may get trapped too
// easily. Larger values may use a bit more memory and seem a bit slower.
// Around 5 – 10 seems reasonable.
const MAX_BREAKOUT_LEVEL = 5

func (pmon *placementMonitor) breakout(level int) bool {
	// Suspend the current search, saving its pmon.bestState.
	// Allow an unconditional placement of the topmost unplaced activity.
	// Then follow this line of development until a penalty is reached
	// which is less than the suspended best. This function can be called
	// recursively to allow more radical jumps in the search space. But the
	// depth of recursion is limited.

	// If a placement attempt is not accepted, the state will revert to
	// the pmon.bestState as it was on entry. So on entry to this function
	// state = pmon.bestState is necessary.
	// On exit, either the pmon.bestState will have been updated to the
	// improved version, or the entry version will be reinstated, but in
	// either case, state = pmon.bestState.

	if level > MAX_BREAKOUT_LEVEL {
		return false
	}

	// Remember current best state
	best := pmon.bestState

	ttinfo := pmon.ttinfo
	aix := pmon.unplaced[len(pmon.unplaced)-1]
	a := ttinfo.Activities[aix]
	nposs := len(a.PossibleSlots)
	var dpen Penalty
	i0 := rand.IntN(nposs)
	i := i0
	for {
		// Place the activity.
		slot := a.PossibleSlots[i]
		clashes := ttinfo.FindClashes(aix, slot)
		for _, aixx := range clashes {
			ttinfo.UnplaceActivity(aixx)
		}
		dpen = pmon.place(aix, slot)
		for _, aixx := range clashes {
			dpen += pmon.evaluate1(aixx)
		}
		// Update penalty info
		for r, p := range pmon.pendingPenalties {
			pmon.resourcePenalties[r] = p
		}
		pmon.score += dpen
		// Remove from "unplaced" list
		pmon.unplaced = pmon.unplaced[:len(pmon.unplaced)-1]
		pmon.unplaced = append(pmon.unplaced, clashes...)
		pmon.bestState = pmon.saveState()

		//TODO: Would it be sensible to return as soon as something better
		// than "best" has been found, rather than trying to improve it at
		// this level?

		for {
			if len(pmon.unplaced) == 0 {
				//--pmon.printScore(fmt.Sprintf("ALL PLACED (%d)", level))

				//TODO: exit criteria ...

				if pmon.movePlace(level) {
					// If state is better than "best" return true
					lcur := len(pmon.unplaced)
					lbest := len(best.unplaced)
					if lcur < lbest || (lcur == lbest && pmon.score < best.score) {
						//++pmon.printScore(fmt.Sprintf("return true (%d)", level))
						return true
					}
					continue
				}
				break
			}
			// Seek an improvement within this search frame.
			if pmon.placeEventually() {
				//++pmon.printScore(fmt.Sprintf("placeEventually (%d)", level))
				// If state is better than "best" return true
				lcur := len(pmon.unplaced)
				lbest := len(best.unplaced)
				if lcur < lbest || (lcur == lbest && pmon.score < best.score) {
					//++pmon.printScore(fmt.Sprintf("return true (%d)", level))
					return true
				}
				//TODO?? continue
			}
			// If not successful, recurse, thus taking a more radical step.
			if pmon.breakout(level + 1) {
				//++pmon.printScore(fmt.Sprintf("breakout (%d)", level))
				// If state is better than "best" return true
				lcur := len(pmon.unplaced)
				lbest := len(best.unplaced)
				if lcur < lbest || (lcur == lbest && pmon.score < best.score) {
					//++pmon.printScore(fmt.Sprintf("return true (%d)", level))
					return true
				}
				//TODO?? continue
			}
			//++pmon.printScore(fmt.Sprintf("breakout (%d)", level))
			break
		}

		pmon.bestState = best
		pmon.restoreState(best)

		i += 1
		if i == nposs {
			i = 0
		}
		if i == i0 {
			break
		}
	}
	return false
}
