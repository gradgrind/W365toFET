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

	// Build a map associating each non-fixed activity with its position
	// in the initial list.
	posmap := make([]int, len(ttinfo.Activities))
	for i, aix := range alist {
		posmap[aix] = i
	}

	var pmon *placementMonitor
	{
		var delta int64 = 7 // This might be a reasonable value?
		pmon = &placementMonitor{
			stateStack:    make([]*ttState, len(alist)),
			unplacedIndex: 0,
			notFixed:      alist,

			//--? Which are still needed?

			count:                 delta,
			delta:                 delta,
			added:                 make([]int64, len(ttinfo.Activities)),
			ttinfo:                ttinfo,
			activityPlacementList: posmap,
			unplaced:              alist,
			resourcePenalties:     make([]Penalty, len(ttinfo.Resources)),
			score:                 0,
			pendingPenalties:      map[ttbase.ResourceIndex]Penalty{},
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

	//TODO--
	fmt.Printf("$ UNPLACED: %d ... PENALTY: %d\n", len(alist), pmon.score)

	//TODO--
	state0 := pmon.saveState()
	NR := 1
	tsum := 0.0
	tmax := 0.0
	var tlist []float64
	for i := NR; i != 0; i-- {
		start := time.Now()

		pmon.basicLoop()

		// calculate the exe time
		elapsed := time.Since(start)
		fmt.Printf("\n#### ELAPSED: %s\n\n", elapsed)
		telapsed := elapsed.Seconds()
		if telapsed > tmax {
			tmax = telapsed
		}
		tsum += telapsed
		tlist = append(tlist, telapsed)

		if i != 1 {
			pmon.restoreState(state0)
		}
	}
	tmean := tsum / float64(NR)
	slices.Sort(tlist)
	NR2 := NR / 2
	tmedian := tlist[NR2]
	if NR%2 == 0 {
		tmedian = (tmedian + tlist[NR2+1]) / 2
	}
	fmt.Printf("#+++ MEAN: %.2f s, MEDIAN: %.2f s, MAX: %.2f s.\n",
		tmean, tmedian, tmax)
	return false
	//--

	pmon.basicLoop()
	return false
}

func (pmon *placementMonitor) basicLoop() {

	//TODO: This might need to be placed before the call to "basicLoop":
	pmon.initBreakoutData()           //??
	pmon.bestState = pmon.saveState() //??
	//pmon.placeNonColliding(-1) //??

	//

	//TODO--
	//var blockslot ttbase.SlotIndex
	var bestscore Penalty
	var bestUnplacedIndex int
	//var end0 Penalty = 0
	//--

	revertx := 1000
	sleeping := false

	for {
		//pmon.fullIntegrityCheck()
		//TODO: exit criteria

		if bestscore != pmon.bestState.score ||
			bestUnplacedIndex != pmon.bestState.unplacedIndex {
			bestscore = pmon.bestState.score
			bestUnplacedIndex = pmon.bestState.unplacedIndex

			//TODO--
			revertx = 1000

			//TODO--
			fmt.Printf("NEW SCORE:: %d / %d : %d\n",
				bestUnplacedIndex, len(pmon.notFixed), bestscore)
		}

		/* ???
		if bestunplaced == 0 {
			//return // to exit when all activities have been placed
			if end0 == 0 {
				end0 = bestscore / 10
			}
			if bestscore <= end0 {
				return
			}
		}
		*/

		if pmon.unplacedIndex == len(pmon.notFixed) {
			pmon.unplacedIndex--
		}

		//TODO--
		if sleeping {
			time.Sleep(100 * time.Millisecond)
		}
		//--

		if pmon.placeNonColliding() {
			pmon.unplacedIndex++
			//fmt.Printf(" -- Step: %d\n", pmon.unplacedIndex)
			continue
		}
		// No further possibilities with this activity
		pmon.unplacedIndex--
		//TODO--
		if pmon.unplacedIndex < revertx {
			revertx = pmon.unplacedIndex
			//sleeping = true
			fmt.Printf(" -- Revert: %d\n", revertx)
		}

		if pmon.unplacedIndex < 0 {
			break
		}

	}
}

func (pmon *placementMonitor) placeNextActivity() {
	aslots := pmon.nextActivity()
	nslots := len(aslots.slots)
	if nslots != 0 {
		// Choose one of the slots
		slot := aslots.slots[rand.IntN(nslots)]
		dpen := pmon.place(aslots.aix, slot)
		// Update penalty info
		for r, p := range pmon.pendingPenalties {
			pmon.resourcePenalties[r] = p
		}
		pmon.score += dpen
	} else {
		//TODO: Might want to try the possible slots and choose (probably?)
		// one of the better ones?
	}
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

func (pmon *placementMonitor) printStateScore(msg string, state *ttState) {
	fmt.Printf("§ StateScore: %s %d\n", msg,
		state.score+Penalty(len(state.unplaced))*PENALTY_UNPLACED_ACTIVITY)
}
