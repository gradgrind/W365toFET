package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
)

const PENALTY_UNPLACED_ACTIVITY = 1000

func PlaceLessons3(
	ttinfo *ttbase.TtInfo,
	alist []ttbase.ActivityIndex,
) bool {
	//resourceSlotActivityMap := makeResourceSlotActivityMap(ttinfo)
	var pmon placementMonitor
	{
		//var delta int64 = 7 // This might be a reasonable value?
		pmon = placementMonitor{
			//count:                   delta,
			//delta:                   delta,
			//added:                   make([]int64, len(ttinfo.Activities)),
			ttinfo: ttinfo,
			//preferEarlier:           preferEarlier,
			//preferLater:             preferLater,
			//resourceSlotActivityMap: resourceSlotActivityMap,
			resourcePenalties: make([]int, len(ttinfo.Resources)),
			score:             0,
			pendingPenalties:  []resourcePenalty{},
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
		pmon.score+PENALTY_UNPLACED_ACTIVITY*len(alist))

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

func (pmon placementMonitor) testPlacement(
	aix ttbase.ActivityIndex,
) []ttbase.ActivityIndex {

	return []ttbase.ActivityIndex{aix}
}

func (pmon placementMonitor) getNeighbour() int {
	// Try all possible placements of the next activity, accepting one
	// if it reduces the penalty. (Start testing at random slot?)
	// Accept a worsening with a certain probability (SA?)?.
	// If all fail choose a weighted probability?

	delta := 0

	return delta
}

func (pmon placementMonitor) sa1() {
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
