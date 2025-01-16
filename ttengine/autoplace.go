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

	// Might improve speed considerably, especially with complex data:
	slices.Reverse(alist)

	//TODO-- currently not used
	// Build a map associating each non-fixed activity with its position
	// in the initial list.
	posmap := make([]int, len(ttinfo.Activities))
	for i, aix := range alist {
		posmap[aix] = i
	}

	var pmon *placementMonitor
	{
		var delta int64 = 100 // This might be a reasonable value?
		// For x01 it seems to have little effect, for Demo1 very much?
		// Though it seems very variable, suggesting that certain choices
		// might be more significant ... an escape function is probably
		// needed.

		pmon = &placementMonitor{
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
	pmon.initBreakoutData() //??

	state0 := pmon.saveState()
	pmon.bestState = state0 //??
	pmon.stateStack = append(pmon.stateStack, state0)
	//pmon.placeNonColliding(-1) //??

	for {
		//pmon.fullIntegrityCheck()
		//TODO: exit criteria

		if len(pmon.unplaced) == 0 {
			return

			//TODO ...
			pmon.removeRandomActivity()

			//break
		}

		wb := pmon.placeNextActivity()

		//TODO: Check for improvement
		// Test whether the best score has been beaten.
		lbest0 := len(state0.unplaced)
		lcur := len(pmon.unplaced)
		if lcur < lbest0 || (lcur == lbest0 && pmon.score < state0.score) {
			lbest := len(pmon.bestState.unplaced)
			bestscore := pmon.bestState.score

			state0.lives = 0 // ??

			//TODO: This might need bounding ...
			state0 = pmon.saveState()
			pmon.stateStack = append(pmon.stateStack, state0)

			if lcur < lbest || (lcur == lbest && pmon.score < bestscore) {
				pmon.bestState = state0
				//TODO--
				fmt.Printf("NEW SCORE:: %d : %d\n", lcur, pmon.score)
			}

		} else if wb {
			continue
			// Revert to a previous state?

			//if rand.IntN(100) > 50 {
			if state0.lives < 10 {
				// restore state?
				if rand.IntN(100) > 50 {
					pmon.restoreState(state0)
				}
				state0.lives++
			} else {
				lstack := len(pmon.stateStack)
				lstack--
				if lstack == 0 {
					panic("State stack empty")
				}
				state0 = pmon.stateStack[lstack-1]
				pmon.restoreState(state0)
				pmon.stateStack = pmon.stateStack[:lstack]
				//fmt.Printf("*** RESTORE *** %d / %d\n", lstack, state0.lives)
			}
		}

	}
}

type slotChoice struct {
	ptotal int
	plist  []int
	clist  [][]ttbase.ActivityIndex
	slist  []ttbase.SlotIndex
}

func (pmon *placementMonitor) placeNextActivity() bool {
	ttinfo := pmon.ttinfo
	uix := len(pmon.unplaced) - 1
	aix := pmon.unplaced[uix]

	// Find possible slots
	a := ttinfo.Activities[aix]
	slots := []ttbase.SlotIndex{}
	for _, slot := range a.PossibleSlots {
		if ttinfo.TestPlacement(aix, slot) {
			slots = append(slots, slot)
		}
	}
	nslots := len(slots)
	if nslots != 0 {
		slot := slots[rand.IntN(nslots)]
		dpen := pmon.place(aix, slot)
		// Update penalty info
		for r, p := range pmon.pendingPenalties {
			pmon.resourcePenalties[r] = p
		}
		pmon.score += dpen
		// Remove from "unplaced" list
		pmon.unplaced = pmon.unplaced[:uix]
		return true
	}

	//
	// Try the possible slots and choose (probably) one of the better
	// ones.

	//TODO: This probably shouldn't return until the number of unplaced
	// activities is down to uix – or it is established that that isn't
	// likely to be reached.

	var clashes []ttbase.ActivityIndex
	wb := false
	// Distinguish between slots which would cause removal of temporarily
	// blocked activities and those which wouldn't (preferred).
	// No blocked activity removal:
	nbslots := slotChoice{}
	// With blocked activity removal:
	wbslots := slotChoice{}

	var slot ttbase.SlotIndex
	for _, slot = range a.PossibleSlots {
		clashes = ttinfo.FindClashes(aix, slot)
		if len(clashes) == 0 {
			// There should be no slots without clashes.
			panic("Unexpectedly: no clashes")
		}

		for _, aixx := range clashes {
			if pmon.doNotRemove(aixx) {

				if nbslots.ptotal == 0 {
					wbslots.ptotal += 1000 / len(clashes)
					wbslots.plist = append(wbslots.plist, wbslots.ptotal)
					wbslots.clist = append(wbslots.clist, clashes)
					wbslots.slist = append(wbslots.slist, slot)
				}

				goto nextslot
			}
		}

		nbslots.ptotal += 1000 / len(clashes)
		nbslots.plist = append(nbslots.plist, nbslots.ptotal)
		nbslots.clist = append(nbslots.clist, clashes)
		nbslots.slist = append(nbslots.slist, slot)

	nextslot:
	}

	if nbslots.ptotal == 0 {
		nbslots = wbslots
		wb = true
	}
	i, _ := slices.BinarySearch(
		nbslots.plist, rand.IntN(nbslots.ptotal))
	slot = nbslots.slist[i]
	clashes = nbslots.clist[i]

	for _, aixx := range clashes {
		ttinfo.UnplaceActivity(aixx)
	}
	// Update "unplaced" list
	pmon.unplaced = pmon.unplaced[:uix]
	pmon.unplaced = append(pmon.unplaced, clashes...)
	dpen := pmon.place(aix, slot)
	for _, aixx := range clashes {
		dpen += pmon.evaluate1(aixx)
	}
	// Update penalty info
	for r, p := range pmon.pendingPenalties {
		pmon.resourcePenalties[r] = p
	}
	pmon.score += dpen

	return wb
}

func (pmon *placementMonitor) placeNextActivity_0() {
	aslots := pmon.nextActivity()
	aix := aslots.aix
	nslots := len(aslots.slots)
	var slot ttbase.SlotIndex
	if nslots != 0 {
		// Choose one of the slots
		slot = aslots.slots[rand.IntN(nslots)]
		dpen := pmon.place(aix, slot)
		// Update penalty info
		for r, p := range pmon.pendingPenalties {
			pmon.resourcePenalties[r] = p
		}
		pmon.score += dpen
	} else {
		//TODO: Might want to try the possible slots and choose (probably?)
		// one of the better ones?

		ttinfo := pmon.ttinfo
		a := ttinfo.Activities[aix]
		nslots = len(a.PossibleSlots)
		i0 := rand.IntN(nslots)
		i := i0
		var clashes []ttbase.ActivityIndex
		nclashes := 100
		slot0 := -1
		for {
			slot = a.PossibleSlots[i]
			clashes = ttinfo.FindClashes(aix, slot)
			if len(clashes) == 0 {
				// There should be no slots without clashes.
				panic("Unexpectedly: no clashes")
			}

			if len(clashes) < nclashes {
				slot0 = slot
				nclashes = len(clashes)
			}

			for _, aixx := range clashes {
				if pmon.doNotRemove(aixx) {
					goto nextslot
				}
			}
			break
		nextslot:
			i += 1
			if i == nslots {
				i = 0
			}
			if i == i0 {
				// If no alternative, allow removal of temporarily blocked
				// activities.
				slot = slot0
				clashes = ttinfo.FindClashes(aix, slot)
				pmon.xcount++
				if pmon.xcount%1000 == 0 {
					fmt.Printf("?xcount? %d\n", pmon.xcount)
				}
				break
			}
		}
		for _, aixx := range clashes {
			ttinfo.UnplaceActivity(aixx)
		}
		pmon.unplaced = append(pmon.unplaced, clashes...)
		dpen := pmon.place(aix, slot)
		for _, aixx := range clashes {
			dpen += pmon.evaluate1(aixx)
		}
		// Update penalty info
		for r, p := range pmon.pendingPenalties {
			pmon.resourcePenalties[r] = p
		}
		pmon.score += dpen
	}
	// Remove from "unplaced" list
	pmon.unplaced = slices.DeleteFunc(pmon.unplaced,
		func(aix1 ttbase.ActivityIndex) bool {
			return aix1 == aix
		})
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
