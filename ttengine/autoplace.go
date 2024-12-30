package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
)

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
	var satemp Penalty = 1000

	satemp = pmon.place1(satemp)
	var sat1 Penalty
	for len(pmon.unplaced) != 0 {
		score := pmon.bestState.score
		if score < pmon.currentState.score {
			pmon.currentState = pmon.bestState
			pmon.resetState()
			sat1 = satemp
		} else {
			satemp = sat1
		}
		pmon.place2()
		sat1 = pmon.place1(100)
	}

	return true
}

func (pmon *placementMonitor) place1(satemp Penalty) Penalty {
	for satemp != 0 {
		_, ok := pmon.step(satemp)
		if !ok {
			return -1
		}
		pmon.currentState = pmon.saveState()
		//fmt.Printf("++ T=%d Unplaced: %d Penalty: %d\n",
		//	satemp, len(pmon.unplaced), dp)
		if pmon.currentState.score < pmon.bestState.score {
			pmon.bestState = pmon.currentState
		}
		satemp *= 9980
		satemp /= 10000
	}
	return satemp
}

func (pmon *placementMonitor) place2() Penalty {
	ttinfo := pmon.ttinfo
	aix := pmon.unplaced[len(pmon.unplaced)-1]
	a := ttinfo.Activities[aix]
	nposs := len(a.PossibleSlots)
	var dpen Penalty
	slot := a.PossibleSlots[rand.IntN(nposs)]
	clashes := ttinfo.FindClashes(aix, slot)
	for _, aixx := range clashes {
		ttinfo.UnplaceActivity(aixx)
	}
	ttinfo.PlaceActivity(aix, slot)
	pmon.added[aix] = pmon.count
	pmon.count++
	dpen = pmon.evaluate1(aix) +
		PENALTY_UNPLACED_ACTIVITY*Penalty(len(clashes)-1)
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
	fmt.Printf("$$$ %d %d %d\n", aix, dpen, pmon.score)
	return dpen
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

func (pmon *placementMonitor) testPlacement(
	aix ttbase.ActivityIndex,
) []ttbase.ActivityIndex {

	return []ttbase.ActivityIndex{aix}
}

func (pmon *placementMonitor) step(temp Penalty) (Penalty, bool) {
	// Try all possible placements of the next activity, accepting one
	// if it reduces the penalty. (Start testing at random slot?)
	// Accept a worsening with a certain probability (SA?)?.
	// If all fail choose a weighted probability?
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

	clear(pmon.pendingPenalties)
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
			dpen = pmon.evaluate1(aix) - PENALTY_UNPLACED_ACTIVITY
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
		var dpen Penalty
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
		dpen = pmon.evaluate1(aix) +
			PENALTY_UNPLACED_ACTIVITY*Penalty(len(clashes)-1)
		for _, aixx := range clashes {
			dpen += pmon.evaluate1(aixx)
		}

		// Decide whether to accept
		if dpen <= 0 {
			goto accept // (not very likely!)
		} else {
			dfac := dpen / temp
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

func (pmon *placementMonitor) sa1() {
	/*
		step := 1
		accept := 0
		for step < STEP_MAX && t >= TEMP_MIN {
			proposed_neighbour := pmon.getNeighbour()
			dpen := pmon.costFunc(proposed_neighbour)

			//

			//

			t = updateT(step)
			step += 1
		}
		/*
		   	# begin optimizing

		   self.step, self.accept = 1, 0
		   while self.step < self.step_max and self.t >= self.t_min:

		   	# get neighbor
		   	proposed_neighbor = self.get_neighbor()

		   	# check energy level of neighbor
		   	E_n = self.cost_func(proposed_neighbor)
		   	dE = E_n - self.current_energy

		   	# determine if we should accept the current neighbor
		   	if random() < self.safe_exp(-dE / self.t):
		   	    self.current_energy = E_n
		   	    self.current_state = proposed_neighbor[:]
		   	    self.accept += 1

		   	# check if the current neighbor is best solution so far
		   	if E_n < self.best_energy:
		   	    self.best_energy = E_n
		   	    self.best_state = proposed_neighbor[:]

		   	# update some stuff
		   	self.t = self.update_t(self.step)
		   	self.step += 1
	*/
}
