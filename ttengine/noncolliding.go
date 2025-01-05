package ttengine

import (
	"W365toFET/ttbase"
	"math/rand/v2"
)

func (pmon *placementMonitor) placeNonColliding(
	block ttbase.SlotIndex, // Don't use this slot (-1 => none blocked)
) bool {
	// Try to place the topmost unplaced activity (repeatedly).
	// Try all possible placements until one is found that doesn't require
	// the removal of another activity. Start searching at a random slot,
	// only testing those in the activity's "PossibleSlots" list.
	// Repeat until no more activities can be placed.
	// pmon.bestState is updated if – and only if – there is an improvement.
	// Return true if pmon.bestState has been updated.
	better := false
	ttinfo := pmon.ttinfo
	lcur := len(pmon.unplaced)
	for lcur != 0 {
		// Read top activity from unplaced-stack
		aix := pmon.unplaced[lcur-1]
		a := ttinfo.Activities[aix]
		if a.Placement >= 0 {
			panic("BUG: expecting unplaced activity")
		}
		nposs := len(a.PossibleSlots)
		i0 := rand.IntN(nposs)
		// Seek a non-colliding placement
		i := i0
		var dpen Penalty
		for {
			if i != block {
				// Try one slot after the other.
				slot := a.PossibleSlots[i]
				if ttinfo.TestPlacement(aix, slot) {
					// Place and reevaluate
					dpen = pmon.place(aix, slot)

					//TODO: Perhaps there should be some consideration of dpen?

					break
				}
			}
			i += 1
			if i == nposs {
				i = 0
			}
			if i == i0 {
				// No non-colliding placement possible
				return better
			}
		}
		// Remove activity from unplaced stack
		lcur--
		pmon.unplaced = pmon.unplaced[:lcur]
		// The initially blocked slot should now be unblocked.
		block = -1
		// Update penalty info
		for r, p := range pmon.pendingPenalties {
			pmon.resourcePenalties[r] = p
		}
		pmon.score += dpen
		// Test whether the best score has been beaten.
		lbest := len(pmon.bestState.unplaced)
		if lcur < lbest || (lcur == lbest && pmon.score < pmon.bestState.score) {
			pmon.bestState = pmon.saveState()
			better = true
		}
	}
	return better
}
