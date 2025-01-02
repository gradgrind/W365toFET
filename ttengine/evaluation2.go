package ttengine

import (
	"W365toFET/ttbase"
	"math/rand/v2"
	"slices"
)

func (pmon *placementMonitor) choosePlacedActivity() ttbase.ActivityIndex {
	// Find the scores for all activities, choose one at random, preferring
	// those with large penalties.

	ttinfo := pmon.ttinfo
	var total Penalty = -1
	pvec := make([]Penalty, len(ttinfo.Activities))
	pvec[0] = -1
	for aix := 1; aix < len(ttinfo.Activities); aix += 1 {
		a := ttinfo.Activities[aix]
		if a.Placement >= 0 && !a.Fixed {
			for _, r := range a.Resources {
				total += pmon.resourcePenalties[r]
			}
		}
		pvec[aix] = total
	}
	aix, _ := slices.BinarySearch(pvec, Penalty(rand.IntN(int(total)+1)))
	return ttbase.ActivityIndex(aix)
}
