package ttengine

import (
	"W365toFET/ttbase"
	"math/rand/v2"
)

func (pmon *placementMonitor) placeConditional() bool {
	// Force a placement of the next activity if one of the possibilities
	// leads – after a call to "placeNonColliding" – to an improved score.
	// Depending an a probability function a worsened state might be accepted.
	// On failure (non-acceptance), restore entry state and return false.
	// Note that pmon.bestState is not necessarily changed, even if true
	// is returned.
	ttinfo := pmon.ttinfo
	aix := pmon.unplaced[len(pmon.unplaced)-1]
	a := ttinfo.Activities[aix]
	nposs := len(a.PossibleSlots)
	var dpen Penalty
	var dpenx Penalty
	i0 := rand.IntN(nposs)
	i := i0
	// Save entry state.
	state0 := pmon.saveState()

	//TODO: Initial threshold = ?
	var threshold Penalty = 5

	var clashes []ttbase.ActivityIndex
	for {
		slot := a.PossibleSlots[i]
		clashes = ttinfo.FindClashes(aix, slot)
		if len(clashes) == 0 {
			// Only accept slots where a replacement is necessary.
			goto nextslot
		}
		for _, aixx := range clashes {
			if pmon.doNotRemove(aixx) {
				goto nextslot
			}
		}
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
		// ... and add removed activities
		pmon.unplaced = append(pmon.unplaced, clashes...)

		if pmon.placeNonColliding(-1) {
			return true
		}
		// Allow more flexible acceptance.
		dpenx = dpen + PENALTY_UNPLACED_ACTIVITY*Penalty(
			len(pmon.unplaced)-len(state0.unplaced))
		// Decide whether to accept.
		if dpenx <= 0 {
			return true // (just in case ...)
		} else {
			dfac := dpenx / threshold
			// The traditional exponential function seems no better,
			// this function may be a little faster?
			t := N1 / (dfac*dfac + N2)
			//t := Penalty(math.Exp(float64(-dfac)) * float64(N0))
			if t != 0 && Penalty(rand.IntN(N0)) < t {
				//TODO: A different probability of acceptance (< t*N, with N
				// <1 or >1) may be helpful under some circumstances.
				// Perhaps it should be variable?
				return true
			}
		}

		// Restore state.
		pmon.restoreState(state0)
	nextslot:
		i += 1
		if i == nposs {
			i = 0
		}
		if i == i0 {
			// All slots have been tested.
			break
		}
	}
	return false
}
