package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"math/rand/v2"
	"slices"
)

type placementMonitor struct {
	count int64
	delta int64
	added []int64
	//
	ttinfo                  *ttbase.TtInfo
	preferEarlier           []int
	preferLater             []int
	resourceSlotActivityMap map[ttbase.ResourceIndex]map[int][]ttbase.ActivityIndex
}

func (pm *placementMonitor) check(aix ttbase.ActivityIndex) bool {
	// Return true if only fixed or "recently" placed.
	aixc := pm.added[aix]
	return aixc < 0 || pm.count-aixc < pm.delta
}

func possibleSlots(
	ttinfo *ttbase.TtInfo,
	aix ttbase.ActivityIndex,
) []int {
	// Get possible slots for the given activity
	a := ttinfo.Activities[aix]
	poss := []int{}
	for _, slot := range a.PossibleSlots {
		day := slot / ttinfo.NHours
		for _, addix := range a.DifferentDays {
			add := ttinfo.Activities[addix]
			if add.Placement >= 0 && add.Placement/ttinfo.NHours == day {
				//TODO: Maybe it's OK if the course is different?
				// Could try accepting these later, if there are
				// otherwise no possible slots?
				goto fail
			}
		}
		poss = append(poss, slot)
	fail:
	}
	return poss
}

func tryToPlace(ttinfo *ttbase.TtInfo, aix ttbase.ActivityIndex) bool {
	a := ttinfo.Activities[aix]
	n := len(a.PossibleSlots)
	// Pick one at random
	i0 := rand.IntN(n)
	// Test all possible slots, starting at this index, until a free one
	// is found.
	i := i0
	for {
		if ttinfo.TestPlacement(aix, a.PossibleSlots[i]) {
			ttinfo.PlaceActivity(aix, a.PossibleSlots[i])
			return true
		}
		i++
		if i == n {
			i = 0
		}
		if i == i0 {
			break
		}
	}
	return false
}

func placeFree(
	ttinfo *ttbase.TtInfo,
	weights []int,
	aix ttbase.ActivityIndex,
) bool {
	// Place the activity in one of the available slots. If no slot is
	// available return false.
	slots := freeSlots(ttinfo, aix)
	if len(slots) == 0 {
		return false
	}
	var slot int
	if len(slots) == 1 {
		slot = slots[0]
	} else {
		//slot = slots[rand.IntN(len(slots))]
		slot = slots[chooseWeightedSlot(weights, slots)]
	}
	ttinfo.PlaceActivity(aix, slot)
	return true
}

func freeSlots(ttinfo *ttbase.TtInfo, aix ttbase.ActivityIndex) []int {
	// Get free slots for the given activity
	a := ttinfo.Activities[aix]
	var slots []int
	for _, p := range a.PossibleSlots {
		if ttinfo.TestPlacement(aix, p) {
			slots = append(slots, p)
		}
	}
	return slots
}

func buildEarlyHourWeights(ndays int, nhours int, earlytimes int) []int {
	// Construct a list of weights favouring early hours, especially those
	// within the minimum hours range. Each time slot within the week has
	// an entry.
	// The weights are such that slots within up to break point (e.g. beginning
	// of lunch time, it could be earlier, though) are equally preferred. Later
	// slots can get a much lower weight, for example (for one day):
	//    10, 10, 10, 10, 5, 4, 3, 2, 1, 1

	w0 := 10
	w1 := 5
	dweights := make([]int, nhours)
	for h := 0; h < nhours; h++ {
		if h < earlytimes {
			dweights[h] = w0
		} else {
			dweights[h] = w1
			if w1 > 1 {
				w1--
			}
		}
	}
	wweights := []int{}
	for d := 0; d < ndays; d++ {
		wweights = append(wweights, dweights...)
	}
	return wweights
}

func buildLateHourWeights(ndays int, nhours int, latetimes int) []int {
	// Construct a list of weights favouring later hours. Each time slot
	// within the week has an entry.
	// The weights are such that slots after a break point (e.g. beginning
	// of lunch time) are equally preferred. Earlier slots can get a much
	// lower weight, for example (for one day):
	//    1, 2, 3, 4, 5, 10, 10, 10, 10, 10

	w0 := 10
	w1 := 5
	dweights := make([]int, nhours)
	for h := nhours - 1; h >= 0; h-- {
		if h >= latetimes {
			dweights[h] = w0
		} else {
			dweights[h] = w1
			if w1 > 1 {
				w1--
			}
		}
	}
	wweights := []int{}
	for d := 0; d < ndays; d++ {
		wweights = append(wweights, dweights...)
	}
	return wweights
}

func chooseWeightedSlot(weights []int, slots []int) int {
	wlist := make([]int, len(slots))
	w := -1
	for i, slot := range slots {
		w += weights[slot]
		wlist[i] = w
	}
	n, _ := slices.BinarySearch(wlist, rand.IntN(w+1))
	return n

	//TODO--

	/* test
	wlist = []int{9, 19, 29, 34, 39, 41, 43, 44, 45}
	collect := map[int]int{}
	for i := 0; i < 100000000; i++ {
		w := rand.IntN(45 + 1)
		n, _ := slices.BinarySearch(weights, w)
		collect[n]++
		//fmt.Printf("===== %d: %d\n", w, n)
	}

	for i := 0; i < len(weights); i++ {
		fmt.Printf(">>>>> %d: %d\n", i, collect[i])
	}
	*/

}

func makeResourceSlotActivityMap(ttinfo *ttbase.TtInfo,
) map[ttbase.ResourceIndex]map[int][]ttbase.ActivityIndex {
	rpamap := map[ttbase.ResourceIndex]map[int][]ttbase.ActivityIndex{}
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		a := ttinfo.Activities[aix]
		for _, r := range a.Resources {
			pamap, ok := rpamap[r]
			if !ok {
				pamap = map[int][]ttbase.ActivityIndex{}
			}
			for _, p := range a.PossibleSlots {
				pamap[p] = append(pamap[p], aix)
			}
			rpamap[r] = pamap
		}
	}
	return rpamap
}

func integrityCheck(ttinfo *ttbase.TtInfo) {
	rmap := make([]bool, len(ttinfo.TtSlots))
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		a := ttinfo.Activities[aix]
		p := a.Placement
		if p < 0 {
			continue
		}
		for _, r := range a.Resources {
			rix := r*ttinfo.SlotsPerWeek + p
			for i := 0; i < a.Duration; i++ {
				if ttinfo.TtSlots[rix+i] != aix {
					base.Bug.Fatalf("$?$ Resource: %d Slot: %d --> %d\n"+
						"  -- Activity: %+v\n",
						r, p, ttinfo.TtSlots[rix+i], a)
				}
				rmap[rix+i] = true
			}
		}
	}
	for rix, rp := range rmap {
		if !rp {
			aix := ttinfo.TtSlots[rix]
			if aix > 0 {
				base.Bug.Fatalf("$!$ Resource: %d Slot: %d -->"+
					" Activity: %+v\n",
					rix/ttinfo.SlotsPerWeek,
					rix%ttinfo.SlotsPerWeek,
					ttinfo.Activities[aix],
				)
			}

		}
	}
}

type activityPlacement struct {
	placement int
	//fixed bool
	xrooms []ttbase.ResourceIndex
}

type ttState struct {
	placements []activityPlacement
	added      []int64
	count      int64
}

func (pmon *placementMonitor) saveState() ttState {
	alist := pmon.ttinfo.Activities
	state := ttState{}
	state.placements = make([]activityPlacement, len(alist))
	for aix := 1; aix < len(alist); aix++ {
		a := alist[aix]
		ap := activityPlacement{
			placement: a.Placement,
			//fixed: a.Fixed,
			//xrooms: a.Xrooms,
		}
		state.placements[aix] = ap
	}
	state.added = make([]int64, len(pmon.added))
	copy(state.added, pmon.added)
	state.count = pmon.count
	return state
}

func (pmon *placementMonitor) restoreState(state ttState) {
	// This assumes the length of the activities list is fixed. If new
	// activities are added, or some removed, appropriate changes would
	// need to be made.
	alist := pmon.ttinfo.Activities
	// Integrity check
	if len(alist) != len(state.placements) {
		base.Bug.Fatalln("State restoration: number of activities changed")
	}
	for aix := 1; aix < len(alist); aix++ {
		a := alist[aix]
		ap := state.placements[aix]
		a.Placement = ap.placement
		//a.Fixed = ap.fixed
		//a.Xrooms = ap.xrooms
	}
	pmon.added = make([]int64, len(state.added))
	copy(pmon.added, state.added)
	pmon.count = state.count
	//TODO: Set the resource allocation
}
