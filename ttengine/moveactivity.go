package ttengine

import (
	"W365toFET/ttbase"
	"math/rand/v2"
	"slices"
)

func (pmon *placementMonitor) removeRandomActivity() ttbase.SlotIndex {
	// Select an activity and "unplace" it.
	// pmon.bestState is not changed.
	// Return the slot from which the activity was removed.
	ttinfo := pmon.ttinfo
	// Construct a cumulative array of the penalties for each activity.
	// This allows an activity to be chosen with a probability proportional
	// to its penalty. Fixed and unplaced activities, as well as those which
	// have only recently been placed, are given no penalty, so that they
	// cannot be chosen.
	var total Penalty = -1
	pvec := make([]Penalty, len(ttinfo.Activities))
	pvec[0] = -1
	for aix := 1; aix < len(ttinfo.Activities); aix += 1 {
		a := ttinfo.Activities[aix]
		if a.Placement >= 0 && !a.Fixed && !pmon.doNotRemove(aix) {
			for _, r := range a.Resources {
				total += pmon.resourcePenalties[r]
			}
		}
		pvec[aix] = total
	}
	// Choose an activity.
	aix, _ := slices.BinarySearch(pvec, Penalty(rand.IntN(int(total)+1)))
	// Displace the activity.
	slot := ttinfo.Activities[aix].Placement
	ttinfo.UnplaceActivity(aix)
	pmon.unplaced = append(pmon.unplaced, aix)
	// Update penalty info
	clear(pmon.pendingPenalties)
	pmon.score += pmon.evaluate1(aix)
	for r, p := range pmon.pendingPenalties {
		pmon.resourcePenalties[r] = p
	}
	return ttbase.SlotIndex(slot)
}
