package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
	"slices"
	"time"
)

// The approach used here might be described as "Escalating Radicality". It
// is based on an algorithm something like Simulated Annealing, but it seems
// that the cooling process may not be very helpful. When no further
// improvements can be made, more radical steps are taken, allowing
// increasing worsening of the penalty for a limited period.

// The initial threshold seems to have little effect on the result. Even 0
// seems to produce only a minor deterioration?
const THRESHOLD0 = 1000
const N0 = 1000
const NSTEPS = 1000

// -----------------
const N1 = NSTEPS * NSTEPS
const N2 = N1 / N0

const PENALTY_UNPLACED_ACTIVITY Penalty = 1000

func PlaceLessons(
	ttinfo *ttbase.TtInfo,
	//alist []ttbase.ActivityIndex,
) bool {
	alist := CollectCourseLessons(ttinfo)

	// Seems to improve speed considerably, especially with complex data:
	slices.Reverse(alist)

	var pmon *placementMonitor
	{
		var delta int64 = 7   // This might be a reasonable value?
		var axdelta int64 = 7 // This might be a reasonable value?
		pmon = &placementMonitor{
			count:    delta,
			delta:    delta,
			added:    make([]int64, len(ttinfo.Activities)),
			axcount:  axdelta,
			axdelta:  axdelta,
			axmoved:  make([]int64, len(ttinfo.Activities)),
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

	//TODO--
	state0 := pmon.saveState()
	NR := 1
	tsum := 0.0
	for i := NR; i != 0; i-- {
		start := time.Now()

		pmon.placer()

		// calculate the exe time
		elapsed := time.Since(start)
		fmt.Printf("#### ELAPSED: %s\n", elapsed)
		tsum += elapsed.Seconds()

		if i != 1 {
			pmon.restoreState(state0)
		}
	}
	fmt.Printf("#+++ AVERAGE: %.2f seconds.\n", tsum/float64(NR))
	return false
	//--

	return pmon.placer()
}

func (pmon *placementMonitor) placer() bool {
	pmon.bestState = pmon.saveState()

	// Initial placement of unplaced lessons
	pmon.basicPlacements(THRESHOLD0)
	// state = pmon.bestState

	//++pmon.printScore("basicPlacements")

	//	return false

	for {
		//TODO: Handle optimization when all activities are placed ...
		if len(pmon.unplaced) == 0 {

			pmon.printScore("ALL PLACED (0)")

			//TODO: exit criteria ...

			if pmon.movePlace(1) {
				continue
			}
			fmt.Printf("MOVE FAILED: %+v\n", pmon.unplaced)
			break
		}
		if !pmon.placeEventually() {
			// No improvement was found by "placeEventually".
			// state = pmon.bestState

			// Mechanism to escape to other solution areas:
			// Accept a worsening step and follow its progress a while to see
			// if a better solution area can be found.
			if !pmon.breakout(1) {
				break
			}

			/* Might be useful in some form?
			//TODO: It looks like one retry can help a bit, but repeating it
			// may be unproductive.
			// Reorder unplaced activities
			i -= 1
			if i == 0 {
				break
			}
			laix := pmon.unplaced[lpu-1]
			copy(pmon.unplaced[1:], pmon.unplaced)
			pmon.unplaced[0] = laix
			copy(pmon.currentState.unplaced, pmon.unplaced)
			//++pmon.printScore("Shuffle")
			//

			continue
			*/
		}
		// state = pmon.bestState
		//++pmon.printScore("placeEventually")
	}
	fmt.Printf("§Unplaced: %d\n", len(pmon.unplaced))
	return true
}

func (pmon *placementMonitor) basicPlacements(threshold Penalty) bool {
	// Primary placement algorithm, based on "Simulated Annealing" (but see
	// cooling factor below ...).
	// Final state = bestState
	// Return true if an improvement was made.
	better := false
	for threshold != 0 {
		lcur := len(pmon.unplaced)
		if lcur == 0 {
			//TODO?
			break
		}

		if !pmon.placeTopUnplaced(threshold) {
			// No step made, premature exit
			//fmt.Printf("FAILED: %d\n", aix)

			//?
			break

			/* Something like this? It might well change bestState ...

			aix = pmon.choosePlacedActivity()
			clear(pmon.pendingPenalties)
			pmon.score += pmon.evaluate1(aix)
			// Update penalty info
			for r, p := range pmon.pendingPenalties {
				pmon.resourcePenalties[r] = p
			}
			pmon.unplaced = append(pmon.unplaced, aix)
			pmon.breakout(1)
			*/

		}
		//fmt.Printf("PLACED: %d\n", aix)

		//fmt.Printf("++ T=%d Unplaced: %d Penalty: %d\n",
		//	threshold, len(pmon.unplaced), dp)
		lcur = len(pmon.unplaced)
		lbest := len(pmon.bestState.unplaced)
		if lcur < lbest || (lcur == lbest && pmon.score < pmon.bestState.score) {
			pmon.bestState = pmon.saveState()
			better = true
		}
		// The cooling factor seems not to have a great impact, as long as
		// it's above 0.8 or so?
		// In fact it might be best with no cooling at all, i.e. without
		// the S.A. ...
		//threshold *= 9980
		//threshold /= 10000
	}
	pmon.restoreState(pmon.bestState)
	return better
}

func (pmon *placementMonitor) placeConditional() bool {
	// Force a placement of the next activity if one of the possibilities
	// leads – after a call to "placeNonColliding" – to an improved score.
	// On failure, restore entry state and return false.
	ttinfo := pmon.ttinfo
	aix := pmon.unplaced[len(pmon.unplaced)-1]
	a := ttinfo.Activities[aix]
	nposs := len(a.PossibleSlots)
	var dpen Penalty
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
		if len(clashes) != 0 {
			// Only accept slots where a replacement is necessary.
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
			//TODO: Allow more flexible acceptance.

		}
		// Restore state.
		pmon.restoreState(state0)
		i += 1
		if i == nposs {
			i = 0
		}
		if i == i0 {
			// No improved placement found

			break //??
			// The following may offer no noticeable improvement. More than
			// one repeat may slow the process down.

			threshold *= 2 // TODO??
			//++fmt.Printf("???? %d\n", threshold)
			if threshold > 9 {
				// A larger value could be counterproductive?
				break
			}
		}
	}
	return false
}

func (pmon *placementMonitor) basicLoop() {

	//TODO: This might need to be placed before the call to "basicLoop":
	pmon.placeNonColliding(-1)

	var blockslot ttbase.SlotIndex
	for {
	evaluate:
		//TODO: exit criteria

		blockslot = -1
		for len(pmon.unplaced) == 0 {
			blockslot = pmon.removeRandomActivity()
			if pmon.placeNonColliding(blockslot) {
				// score improved
				goto evaluate
			}
		}
		//TODO: Get a bit more radical – allow activities to be replaced.
		// Perform just one forced placement, followed by placeNonColliding.
		// An increased penalty may be accepted, depending on a probability
		// function.
		if pmon.placeConditional() {
			continue
		}
	}
}

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
