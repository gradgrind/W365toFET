package ttengine

import (
	"W365toFET/ttbase"
	"math/rand/v2"
)

//TODO: make a function for handling no unplaceds?

// TODO?
func (pmon *placementMonitor) breakout2(
	aix ttbase.ActivityIndex,
	level int,
) bool {
	// Suspend the current search, saving its pmon.bestState.
	// Allow an unconditional placement of the given unplaced activity.
	// Then follow this line of development until a penalty is reached
	// which is less than the suspended best. This function can be called
	// recursively to allow more radical jumps in the search space. But the
	// depth of recursion is limited.

	if level > MAX_BREAKOUT_LEVEL {
		return false
	}

	// Remember current best state
	best := pmon.bestState

	ttinfo := pmon.ttinfo
	a := ttinfo.Activities[aix]
	nposs := len(a.PossibleSlots)
	var dpen Penalty
	i0 := rand.IntN(nposs)
	i := i0
	for {
		slot := a.PossibleSlots[i]
		clashes := ttinfo.FindClashes(aix, slot)
		for _, aixx := range clashes {
			ttinfo.UnplaceActivity(aixx)
		}
		ttinfo.PlaceActivity(aix, slot)
		pmon.added[aix] = pmon.count
		pmon.count++
		clear(pmon.pendingPenalties)
		dpen = pmon.evaluate1(aix)
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

		for {
			if len(pmon.unplaced) == 0 {
				//TODO

				break
			}

			// First seek an improvement within this search frame.
			if pmon.placeEventually(-1) {
				//++pmon.printScore(fmt.Sprintf("placeEventually (%d)", level))
				continue
			}
			// If not successful, recurse, thus taking a more radical step.
			if !pmon.breakout(level + 1) {
				break
			}
			//++pmon.printScore(fmt.Sprintf("breakout (%d)", level))
		}
		// state = currentState = bestState, but probably not the same as
		// before the loop ...

		// If better than "best" return true
		lcur := len(pmon.unplaced)
		lbest := len(best.unplaced)
		if lcur < lbest || (lcur == lbest && pmon.score < best.score) {
			//++pmon.printScore(fmt.Sprintf("return true (%d)", level))
			return true
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

	//TODO-- superfluous?
	//pmon.bestState = best
	//pmon.restoreState(best)

	return false
}
