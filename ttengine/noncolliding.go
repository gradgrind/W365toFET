package ttengine

import (
	"W365toFET/ttbase"
	"slices"
)

type resourcePenalty struct {
	resourceIndex ttbase.ResourceIndex
	penalty       Penalty
}

type slotPenalties struct {
	slot              ttbase.SlotIndex
	deltaPenalty      Penalty
	resourcePenalties []resourcePenalty
}

func (pmon *placementMonitor) placeNonColliding() bool {

	//TODO: rewrite:
	// Try to place the topmost unplaced activity (repeatedly).
	// Try all possible placements until one is found that doesn't require
	// the removal of another activity. Start searching at a random slot,
	// only testing those in the activity's "PossibleSlots" list.
	// Repeat until no more activities can be placed.
	// pmon.bestState is updated if – and only if – there is an improvement.
	// Return true if pmon.bestState has been updated.

	ttinfo := pmon.ttinfo

	aix := pmon.notFixed[pmon.unplacedIndex]
	a := ttinfo.Activities[aix]

	var state0 *ttState
	if pmon.stateStack[pmon.unplacedIndex] == nil {
		// Start handling activity
		state0 = pmon.saveState()
		//state0.unplacedIndex = pmon.unplacedIndex
		pmon.stateStack[pmon.unplacedIndex] = state0
		// Sort possible slots
		slots := []slotPenalties{}
		for _, slot := range a.PossibleSlots {
			if ttinfo.TestPlacement(aix, slot) {
				// Place and reevaluate
				dpen := pmon.place(aix, slot)
				pplist := make([]resourcePenalty, 0, len(pmon.pendingPenalties))
				for r, rp := range pmon.pendingPenalties {
					pplist = append(pplist, resourcePenalty{r, rp})
				}
				slots = append(slots, slotPenalties{slot, dpen, pplist})
				// Unplace the activity
				ttinfo.UnplaceActivity(aix)
			}
		}
		slices.SortFunc(slots, func(a, b slotPenalties) int {
			if a.deltaPenalty < b.deltaPenalty {
				// Sort in decreasing penalty order
				return 1
			}
			if a.deltaPenalty == b.deltaPenalty {
				if len(a.resourcePenalties) > len(b.resourcePenalties) {
					// Prefer more resources
					return 1
				}
			}
			return -1
		})
		state0.slots = slots

	} else {
		state0 = pmon.stateStack[pmon.unplacedIndex]
		pmon.restoreState(state0)
	}

	// Choose next available slot
	if len(state0.slots) == 0 {
		pmon.stateStack[pmon.unplacedIndex] = nil
		return false
	}

	// Get data for last slot
	i := len(state0.slots) - 1
	slotData := state0.slots[i]
	ttinfo.PlaceActivity(aix, slotData.slot)

	// Update penalty info
	for r, p := range pmon.pendingPenalties {
		pmon.resourcePenalties[r] = p
	}
	pmon.score += slotData.deltaPenalty

	// Remove last slot from list
	state0.slots = state0.slots[:i]

	// Test whether the best score has been beaten.

	//TODO: The comparison of unplacedIndexes is not correct!

	lbest := pmon.bestState.unplacedIndex
	lcur := pmon.unplacedIndex
	if lcur > lbest || (lcur == lbest && pmon.score < pmon.bestState.score) {
		pmon.bestState = pmon.saveState()
		pmon.bestState.unplacedIndex = lcur + 1 // ???
	}
	return true
}
