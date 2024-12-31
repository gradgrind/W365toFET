package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
)

const TEMPERATURE0 = 1000
const N0 = 1000
const NSTEPS = 1000

// -----------------
const N1 = NSTEPS * NSTEPS
const N2 = N1 / N0

const PENALTY_UNPLACED_ACTIVITY Penalty = 1000

func PlaceLessons3(
	ttinfo *ttbase.TtInfo,
	alist []ttbase.ActivityIndex,
) bool {
	//resourceSlotActivityMap := makeResourceSlotActivityMap(ttinfo)
	var pmon *placementMonitor
	{
		var delta int64 = 7 // This might be a reasonable value?
		pmon = &placementMonitor{
			count:    delta,
			delta:    delta,
			added:    make([]int64, len(ttinfo.Activities)),
			ttinfo:   ttinfo,
			unplaced: alist,
			//preferEarlier:           preferEarlier,
			//preferLater:             preferLater,
			//resourceSlotActivityMap: resourceSlotActivityMap,
			resourcePenalties: make([]Penalty, len(ttinfo.Resources)),
			score:             0,
			pendingPenalties:  map[ttbase.ResourceIndex]Penalty{},
		}
	}
	pmon.initConstraintData()

	// Calculate initial stage 1 penalties
	for r := 0; r < len(ttinfo.Resources); r++ {
		p := pmon.resourcePenalty1(r)
		pmon.resourcePenalties[r] = p
		pmon.score += p
		//fmt.Printf("$ PENALTY %d: %d\n", r, p)
	}

	// Add penalty for unplaced lessons
	fmt.Printf("$ PENALTY %d: %d\n", len(alist),
		pmon.score+PENALTY_UNPLACED_ACTIVITY*Penalty(len(alist)))

	pmon.currentState = pmon.saveState()
	pmon.bestState = pmon.currentState

	pmon.place1(TEMPERATURE0)

	pmon.printScore("place1")

	//	return false

	for len(pmon.unplaced) != 0 {
		if !pmon.place2() {
			break
		}
		pmon.printScore("place2")
	}
	fmt.Printf("§Unplaced: %d\n", len(pmon.currentState.unplaced))
	return false
}

func (pmon *placementMonitor) place1(satemp Penalty) bool {
	// Primary placement algorithm, based on "Simulated Annealing".
	// Final state = currentState = bestState
	// Return true if an improvement was made.
	better := false
	for satemp != 0 {
		_, ok := pmon.step(satemp)
		if !ok {
			// No step made, premature exit
			break
		}

		//fmt.Printf("++ T=%d Unplaced: %d Penalty: %d\n",
		//	satemp, len(pmon.unplaced), dp)
		lcur := len(pmon.unplaced)
		lbest := len(pmon.bestState.unplaced)
		if lcur < lbest || (lcur == lbest && pmon.score < pmon.bestState.score) {
			pmon.bestState = pmon.currentState
			better = true
		}
		satemp *= 9980
		satemp /= 10000
	}
	if pmon.currentState != pmon.bestState {
		pmon.currentState = pmon.bestState
		pmon.resetState()
	}
	return better
}

func (pmon *placementMonitor) place2() bool {
	// Force a placement of the next activity if one of the possibilities
	// leads to an improved score.
	ttinfo := pmon.ttinfo
	aix := pmon.unplaced[len(pmon.unplaced)-1]
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
		pmon.currentState = pmon.saveState()

		//TODO: Initial temperature?
		if pmon.place1(20) {
			return true
		}
		//TODO: Repeat with different start temperature?
		//if pmon.place1(100) {
		//	return true
		//}
		pmon.currentState = pmon.bestState
		pmon.resetState()

		i += 1
		if i == nposs {
			i = 0
		}
		if i == i0 {
			// No improved placement found
			break
		}
	}
	return false
}

/*
	//TODO--
	tstart := time.Now()
	//

	pending := pmon.basicPlaceActivities2(alist)

	//TODO--
	elapsed := time.Since(tstart)
	fmt.Printf("Basic Placement took %s\n", elapsed)
	//

	//TODO--
	slices.Sort(pending)
	fmt.Printf("$$$ Unplaced: %d (%d)\n  -- %+v\n",
		len(pending), pmon.score, pending)
	//

	if len(pending) != 0 {
		pmon.furtherPlacements2(pending)
	}

	//slices.Reverse(failed)
	//l0 := len(failed)
	//fmt.Printf("Remaining: %d\n", l0)

}
*/

func (pmon *placementMonitor) step(temp Penalty) (Penalty, bool) {

	//TODO:
	// Try all possible placements of the next activity, accepting one
	// if it reduces the penalty. (Start testing at random slot?)
	// Accept a worsening with a certain probability (SA?)?.
	// If all fail choose a weighted probability?

	// Assumes entry state = currentState,
	// Final state = (probably changed) currentState,
	// bestState is not affected.

	ttinfo := pmon.ttinfo
	var clashes []ttbase.ActivityIndex

	var aix ttbase.ActivityIndex
	if len(pmon.unplaced) == 0 {

		//TODO

		// Seek the activity with the highest penalty?
		// Or, based on the penalties, choose an activity at random?
		// Block activities which have only recently been placed?

		// Unplace it ...

		return 0, false

	} else {
		aix = pmon.unplaced[len(pmon.unplaced)-1]
	}
	a := ttinfo.Activities[aix]
	nposs := len(a.PossibleSlots)
	i0 := rand.IntN(nposs)

	// Start with non-colliding placements
	i := i0
	var dpen Penalty
	for {
		slot := a.PossibleSlots[i]
		if ttinfo.TestPlacement(aix, slot) {
			// Place and reevaluate
			ttinfo.PlaceActivity(aix, slot)
			pmon.added[aix] = pmon.count
			pmon.count++
			clear(pmon.pendingPenalties)
			dpen = pmon.evaluate1(aix) //- PENALTY_UNPLACED_ACTIVITY
			goto accept
		}
		i += 1
		if i == nposs {
			i = 0
		}
		if i == i0 {
			// No non-colliding placements possible
			break
		}
	}
	// As a non-colliding placement is not possible, try a colliding one.
	for {
		var dpenx Penalty
		slot := a.PossibleSlots[i]
		clashes = ttinfo.FindClashes(aix, slot)
		for _, aixx := range clashes {
			if pmon.check(aixx) {
				goto nextslot
			}
		}

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
		dpenx = dpen + PENALTY_UNPLACED_ACTIVITY*Penalty(len(clashes)-1)

		// Decide whether to accept
		if dpenx <= 0 {
			goto accept // (not very likely!)
		} else {

			//TODO: Compare with exponential function, exp(-dpenx / temp)
			dfac := dpenx / temp
			t := N1 / (dfac*dfac + N2)
			if t != 0 && Penalty(rand.IntN(N0)) < t {
				goto accept
			}
		}

		// Don't accept change, revert
		pmon.resetState()

	nextslot:
		i += 1
		if i == nposs {
			i = 0
		}
		if i == i0 {
			// No non-colliding placements possible
			return 0, false
		}
	}
accept:
	// Update penalty info
	for r, p := range pmon.pendingPenalties {
		pmon.resourcePenalties[r] = p
	}
	pmon.score += dpen
	// Remove from "unplaced" list
	pmon.unplaced = pmon.unplaced[:len(pmon.unplaced)-1]
	pmon.unplaced = append(pmon.unplaced, clashes...)
	pmon.currentState = pmon.saveState()
	return dpen, true
}

func (pmon *placementMonitor) printScore(msg string) {
	var p Penalty = 0
	for r := 0; r < len(pmon.ttinfo.Resources); r++ {
		p += pmon.resourcePenalty1(r)
	}
	fmt.Printf("§ Score: %s %d\n", msg,
		pmon.score+Penalty(len(pmon.unplaced))*PENALTY_UNPLACED_ACTIVITY)
	if p != pmon.score {
		fmt.Printf("§ ... error: %d != %d\n", p, pmon.score)
		panic("!!!")
	}
}
