package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
	"slices"
	"time"
)

// The initial temperature seems to have little effect on the result. Even 0
// seems to produce only a minor deterioration?
const TEMPERATURE0 = 1000
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

	//TODO--
	state0 := pmon.saveState()
	NR := 50
	tsum := 0.0
	for i := 0; i < NR; i++ {
		start := time.Now()

		pmon.Placer()

		// calculate the exe time
		elapsed := time.Since(start)
		fmt.Printf("#### ELAPSED: %s\n", elapsed)
		tsum += elapsed.Seconds()

		pmon.restoreState(state0)
	}
	fmt.Printf("#+++ AVERAGE: %.2f seconds.\n", tsum/float64(NR))
	return false
	//--

	return pmon.Placer()
}

func (pmon *placementMonitor) Placer() bool {
	pmon.bestState = pmon.saveState()

	pmon.place1(TEMPERATURE0)

	//++pmon.printScore("place1")

	//	return false

	for len(pmon.unplaced) != 0 {
		if !pmon.place2() {

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
		//++pmon.printScore("place2")
	}
	fmt.Printf("§Unplaced: %d\n", len(pmon.unplaced))
	return true
}

// TODO: Is there an optimal limit? Too small and it may get trapped too
// easily. Larger values may use a bit more memory and seem a bit slower.
// Around 5 – 10 seems reasonable.
const MAX_BREAKOUT_LEVEL = 5

func (pmon *placementMonitor) breakout(level int) bool {
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
			if pmon.place2() {
				//++pmon.printScore(fmt.Sprintf("place2 (%d)", level))
				continue
			}
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
	pmon.bestState = best
	pmon.restoreState(best)

	return false
}

func (pmon *placementMonitor) place1(satemp Penalty) bool {
	// Primary placement algorithm, based on "Simulated Annealing" (but see
	// cooling factor below ...).
	// Final state = bestState
	// Return true if an improvement was made.
	better := false
	for satemp != 0 {
		lcur := len(pmon.unplaced)
		if lcur == 0 {
			//TODO?
			break
		}

		if !pmon.tryPlacing(satemp) {
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
		//	satemp, len(pmon.unplaced), dp)
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
		//satemp *= 9980
		//satemp /= 10000
	}
	pmon.restoreState(pmon.bestState)
	return better
}

func (pmon *placementMonitor) place2() bool {
	// Force a placement of the next activity if one of the possibilities
	// leads to an improved score.
	// If it fails (return false) the state will be unchanged.
	ttinfo := pmon.ttinfo
	aix := pmon.unplaced[len(pmon.unplaced)-1]
	a := ttinfo.Activities[aix]
	nposs := len(a.PossibleSlots)
	var dpen Penalty
	i0 := rand.IntN(nposs)
	i := i0
	//TODO: Initial temperature = ?
	var temp Penalty = 5
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

		if pmon.place1(temp) {
			return true
		}

		i += 1
		if i == nposs {
			i = 0
		}
		if i == i0 {
			// No improved placement found
			temp *= 2 // TODO??
			//++fmt.Printf("???? %d\n", temp)
			if temp > 9 {
				// A larger value could be counterproductive?
				//if temp > TEMPERATURE0 {
				break
			}
		}
	}
	return false
}

func (pmon *placementMonitor) tryPlacing(
	threshold Penalty,
) bool {
	// Try to place the topmost unplaced activity.
	// Try all possible placements until one is found that reduces the
	// penalty. However, a placement may also be accepted – with a certain
	// probability, based on the penalty increase – if it increases the
	// penalty. Start searching at a random slot, only testing those in the
	// activity's "PossibleSlots" list.

	// If no placement is found, fail and leave state as on entry.
	// pmon.bestState is not affected.

	ttinfo := pmon.ttinfo
	lcur := len(pmon.unplaced)
	if lcur == 0 {
		panic("BUG: not expecting empty unplaced list")
	}
	// Pop activity from "unplaced" list
	lcur--
	aix := pmon.unplaced[lcur]
	pmon.unplaced = pmon.unplaced[:lcur]

	var state0 *ttState
	var clashes []ttbase.ActivityIndex
	a := ttinfo.Activities[aix]
	if a.Placement >= 0 {
		panic("BUG: expecting unplaced activity")
	}
	nposs := len(a.PossibleSlots)
	i0 := rand.IntN(nposs)
	// Start with non-colliding placements
	i := i0
	var dpen Penalty
	for {
		// Try one slot after the other.
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
	state0 = pmon.saveState()
	for {
		var dpenx Penalty
		slot := a.PossibleSlots[i]
		clashes = ttinfo.FindClashes(aix, slot)
		for _, aixx := range clashes {
			if pmon.check(aixx) {
				// Reject if too recently placed.
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
			dfac := dpenx / threshold
			// The traditional exponential function seems no better,
			// this function may be a little faster?
			t := N1 / (dfac*dfac + N2)
			//t := Penalty(math.Exp(float64(-dfac)) * float64(N0))
			if t != 0 && Penalty(rand.IntN(N0)) < t {
				goto accept
			}
		}

		// Don't accept change, revert
		pmon.restoreState(state0)

	nextslot:
		i += 1
		if i == nposs {
			i = 0
		}
		if i == i0 {
			// No further placements possible
			return false
		}
	}
accept:
	// Update penalty info
	for r, p := range pmon.pendingPenalties {
		pmon.resourcePenalties[r] = p
	}
	pmon.score += dpen
	pmon.unplaced = append(pmon.unplaced, clashes...)
	return true
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
