package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
	"slices"
)

func (pmon *placementMonitor) optimize() {
	ttinfo := pmon.ttinfo
	n_activities := len(ttinfo.Activities)

	// Choose activity to replace randomly, based on penalty.
	p_activities := make([]int, n_activities)
	p_activities[0] = -1

	score_0 := pmon.score

	for pmon.score > score_0/5 {
		p_total := -1
		for aix := 1; aix < n_activities; aix++ {
			p := pmon.getActivityPenalty(aix)
			if p > 0 {
				//fmt.Printf("? PENALTY: %d - %d\n", aix, p)
				p_total += int(p)
			}
			p_activities[aix] = p_total
		}

		aix, _ := slices.BinarySearch(p_activities, rand.IntN(p_total))

		if pmon.move(aix) {
			fmt.Printf("?? PENALTY: %d  (%d)\n", pmon.score, pmon.scoreCount)
			pmon.scoreCount = 0
		} else {
			pmon.scoreCount++
			if pmon.scoreCount%10 == 0 {
				fmt.Printf("?? ++++ %d\n", pmon.scoreCount)
			}
		}
	}

	/* Check distribution
	collect := map[int]int{}
	for i := 0; i < 10000; i++ {
		aix, _ := slices.BinarySearch(p_activities, rand.IntN(p_total))
		collect[aix]++
	}
	type ac struct{ a, c int }
	clist := make([]ac, 0, len(collect))
	for a, c := range collect {
		clist = append(clist, ac{a, c})
	}
	slices.SortFunc(clist, func(a, b ac) int {
		return cmp.Compare(b.c, a.c)
	})
	fmt.Printf("? CHOSEN: %+v\n", clist)
	*/
}

func (pmon *placementMonitor) move(aix ttbase.ActivityIndex) bool {
	ttinfo := pmon.ttinfo
	a := ttinfo.Activities[aix]
	slot0 := a.Placement
	state0 := pmon.saveState()
	score0 := pmon.score
	var best *ttState
	for _, slot := range a.PossibleSlots {
		if slot == slot0 {
			continue
		}
		ttinfo.UnplaceActivity(aix)
		//fmt.Printf("$ Try %d@%d (%d)\n", aix, slot, slot0)
		if len(pmon.unplaced) != 0 {
			panic("unplaced!?!?")
		}
		clashes := ttinfo.FindClashes(aix, slot)
		for _, aixc := range clashes {
			ttinfo.UnplaceActivity(aixc)
		}
		ttinfo.PlaceActivity(aix, slot)
		if len(clashes) != 0 {
			pmon.unplaced = clashes
			a.Fixed = true
			pmon.stateStack = pmon.stateStack[:0]
			if !pmon.basicLoop(0, 0) {
				//fmt.Printf("$ MOVE FAILED: %d@%d\n", aix, slot)
				goto next
			}
		}
		pmon.setResourcePenalties()
		//fmt.Printf("$ MOVED PENALTY %d@%d: %d\n", aix, slot, pmon.score)
		if pmon.score < score0 {
			//TODO: Accept the first improvement? ... or seek the best?
			{
				a.Fixed = false
				return true
			}
			// There may not be much of a difference, in fact the simple one
			// (above) might even be a bit faster!
			{
				score0 = pmon.score
				best = pmon.saveState()
			}
		}
	next:
		pmon.restoreState(state0)
		a.Fixed = false
	}
	if best != nil {
		pmon.restoreState(best)
		return true
	}
	return false
}
