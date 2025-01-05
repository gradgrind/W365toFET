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
		var delta int64 = 7 // This might be a reasonable value?
		//var axdelta int64 = 7 // This might be a reasonable value?
		pmon = &placementMonitor{
			count: delta,
			delta: delta,
			added: make([]int64, len(ttinfo.Activities)),
			//axcount:  axdelta,
			//axdelta:  axdelta,
			//axmoved:  make([]int64, len(ttinfo.Activities)),
			ttinfo:            ttinfo,
			unplaced:          alist,
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

		pmon.basicLoop()

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

	pmon.basicLoop()
	return false
}

func (pmon *placementMonitor) basicLoop() {

	//TODO: This might need to be placed before the call to "basicLoop":
	pmon.bestState = pmon.saveState()
	pmon.placeNonColliding(-1)

	var blockslot ttbase.SlotIndex
	levels := []*breakoutLevel{}
	var best *ttState
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
		// Get a bit more radical – allow activities to be replaced.
		// Perform just one forced placement, followed by placeNonColliding.
		// An increased penalty may be accepted, depending on a probability
		// function.
		if pmon.placeConditional() {
			continue
		}
		// Mechanism to escape to other solution areas:
		// Accept a worsening step and follow its progress a while to see
		// if a better solution area can be found.

		// If no improvement has been made, go to the next level, if there is
		// one. If not, choose the next possibility from this level. Once
		// these have all failed to bring an improvement, end this level.

		if pmon.bestState != best {
			pmon.printScore("Best")
			best = pmon.bestState
			// Clear the level stack.
			levels = levels[:0]
		}

		if !pmon.radicalStep(&levels) {
			//TODO: Revert to bestState?
			pmon.printScore("Revert")
			pmon.restoreState(pmon.bestState)
			pmon.fullIntegrityCheck()
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
