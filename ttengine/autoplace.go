package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
	"slices"
	"time"
)

func PlaceLessons(
	ttinfo *ttbase.TtInfo,
	//alist []ttbase.ActivityIndex,
) bool {
	alist := CollectCourseLessons(ttinfo)

	// Might improve speed considerably, especially with complex data:
	slices.Reverse(alist)

	var pmon *placementMonitor
	{
		pmon = &placementMonitor{
			usedSlots:         make([][]ttbase.SlotIndex, len(ttinfo.Activities)),
			maxdepth:          2,
			ttinfo:            ttinfo,
			unplaced:          alist,
			resourcePenalties: make([]Penalty, len(ttinfo.Resources)),
			score:             0,
		}
	}
	pmon.initConstraintData()

	// Calculate initial stage 1 penalties
	pmon.setResourcePenalties()

	//TODO--
	fmt.Printf("$ UNPLACED: %d ... PENALTY: %d\n", len(alist), pmon.score)

	//TODO--
	state0 := pmon.saveState()
	NR := 10
	tsum := 0.0
	tmax := 0.0
	var tlist []float64
	for i := NR; i != 0; i-- {
		pmon.maxdepth = 2
	repeat:
		start := time.Now()

		pmon.usedSlots = make([][]ttbase.SlotIndex, len(ttinfo.Activities))
		pmon.stateStack = []*ttState{}
		if !pmon.basicLoop(0, 0) {
			pmon.maxdepth++
			fmt.Printf("$$$$$$$$$$$$$$$$$$ %d\n", len(pmon.unplaced))
			goto repeat
		}

		pmon.setResourcePenalties()

		fmt.Printf("==> start optimizing: %d (%d)\n",
			len(pmon.unplaced), len(pmon.stateStack))
		pmon.optimize()

		// calculate the exe time
		elapsed := time.Since(start)
		fmt.Printf("\n#### ELAPSED: %s\n\n", elapsed)
		telapsed := elapsed.Seconds()
		if telapsed > tmax {
			tmax = telapsed
		}
		tsum += telapsed
		tlist = append(tlist, telapsed)

		//TODO--
		fmt.Printf("$ NEW PENALTY: %d\n", pmon.score)

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

	//pmon.optimize()

	return false
	//return true
	//--

	pmon.stateStack = []*ttState{}
	pmon.basicLoop(0, 0)
	return false
}

func (pmon *placementMonitor) basicLoop(startlevel int, depth int) bool {
	//TODO--
	//n0 := len(pmon.unplaced)

	stacklevel0 := len(pmon.stateStack)
	for {
		state0 := pmon.saveState()
		//pmon.bestState = state0 //??
		pmon.stateStack = append(pmon.stateStack, state0)
		//pmon.placeNonColliding(-1) //??

		level := len(pmon.unplaced)
		if level == startlevel {
			//TODO: exit criteria
			return true

			//TODO ...
			//pmon.removeRandomActivity()
		}

		//pmon.fullIntegrityCheck()

		aix, slot := pmon.placeNextActivity()
		if aix != 0 {
			if depth == 0 {
				pmon.usedSlots[aix] = append(pmon.usedSlots[aix], slot)
			}

			//fmt.Printf("==> placed: %d (%d) @ %d\n",
			//	n0-len(pmon.unplaced), len(pmon.stateStack), depth)
			//time.Sleep(100 * time.Millisecond)
			continue
		}

		//fmt.Printf("==> unplaced: %d (%d) @ %d\n",
		//	n0-len(pmon.unplaced), len(pmon.stateStack), depth)

		if depth < pmon.maxdepth {

			//TODO: Maybe use a dynamic maximum depth?
			// A lower level seems to run faster, but may not place all
			// activities. It depends on the data. What about incrementing
			// the limit on each fail? Up to a certain limit, then ...?

			stacklevel := len(pmon.stateStack)
			aix, slot := pmon.forceNextActivity(depth)
			if aix != 0 {
				if depth == 0 {
					pmon.usedSlots[aix] = append(pmon.usedSlots[aix], slot)
				}

				pmon.stateStack = pmon.stateStack[:stacklevel]

				//fmt.Printf("==> placed (forced): %d (%d) @ %d\n",
				//	n0-len(pmon.unplaced), len(pmon.stateStack), depth)
				//time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		//fmt.Printf("==> not placed: %d (%d) @ %d\n",
		//	n0-len(pmon.unplaced), len(pmon.stateStack), depth)
		//time.Sleep(100 * time.Millisecond)

		pmon.restoreState(pmon.stateStack[stacklevel0])
		pmon.stateStack = pmon.stateStack[:stacklevel0]
		return false
	}
}

type slotChoice struct {
	ptotal int
	plist  []int
	clist  [][]ttbase.ActivityIndex
	slist  []ttbase.SlotIndex
}

func (pmon *placementMonitor) placeNextActivity() (
	ttbase.ActivityIndex, ttbase.SlotIndex,
) {
	ttinfo := pmon.ttinfo
	uix := len(pmon.unplaced) - 1
	aix := pmon.unplaced[uix]

	usedSlots := pmon.usedSlots[aix]

	// Find possible slots
	a := ttinfo.Activities[aix]
	pslots := a.PossibleSlots
	nslots := len(pslots)
	// Seek one (at random) which can accept the placement.
	i0 := rand.IntN(nslots)
	i := i0
	for {
		slot := pslots[i]
		if !slices.Contains(usedSlots, slot) {
			if ttinfo.TestPlacement(aix, slot) {
				ttinfo.PlaceActivity(aix, slot)
				// Remove from "unplaced" list
				pmon.unplaced = pmon.unplaced[:uix]
				return aix, slot
			}
		}
		i++
		if i == nslots {
			i = 0
		}
		if i == i0 {
			break
		}
	}
	return 0, 0
}

func (pmon *placementMonitor) forceNextActivity(depth int) (
	ttbase.ActivityIndex, ttbase.SlotIndex,
) {
	//fmt.Printf("??? %d @ %d\n", len(pmon.unplaced), depth)
	//time.Sleep(100 * time.Millisecond)

	//
	// Try the possible slots and choose (probably) one of the better
	// ones.
	ttinfo := pmon.ttinfo
	uix := len(pmon.unplaced) - 1
	aix := pmon.unplaced[uix]
	a := ttinfo.Activities[aix]

	usedSlots := pmon.usedSlots[aix]

	//TODO: This probably shouldn't return until the number of unplaced
	// activities is down to uix – or it is established that that isn't
	// likely to be reached.

	var clashes []ttbase.ActivityIndex
	nbslots := slotChoice{}

	var slot ttbase.SlotIndex
	for _, slot = range a.PossibleSlots {
		if slices.Contains(usedSlots, slot) {
			continue
		}
		clashes = ttinfo.FindClashes(aix, slot)
		if len(clashes) == 0 {
			// There should be no slots without clashes.
			panic("Unexpectedly: no clashes")
		}

		// Filter out if one of the activities is fixed
		for _, aixc := range clashes {
			if ttinfo.Activities[aixc].Fixed {
				goto skip
			}
		}

		nbslots.ptotal += 1000 / len(clashes)
		nbslots.plist = append(nbslots.plist, nbslots.ptotal)
		nbslots.clist = append(nbslots.clist, clashes)
		nbslots.slist = append(nbslots.slist, slot)

	skip:
	}

	if len(nbslots.plist) == 0 {
		fmt.Printf("HELP! %d: %+v\n", aix, usedSlots)
		pmon.usedSlots[aix] = pmon.usedSlots[aix][:0]
		return 0, 0
	}

	count := 0
	//??
	state := pmon.stateStack[len(pmon.stateStack)-1]
try_again:
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
	ttinfo.PlaceActivity(aix, slot)

	//TODO ...

	if !pmon.basicLoop(uix, depth+1) {
		if count < 10 {
			// It looks like there should be some repetition allowed, but very
			// low levels can be compensated by increasing depth limit. < 0
			// is not so good, but even < 1 may work. Maybe around < 10 is
			// optimal?
			count++
			pmon.restoreState(state)
			//fmt.Printf("     ))) count: %d // %d\n", count, len(pmon.stateStack))
			goto try_again
		}
		return 0, 0
	}
	return aix, slot
}
