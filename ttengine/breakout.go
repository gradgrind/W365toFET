package ttengine

import (
	"W365toFET/ttbase"
	"math/rand/v2"
)

type breakoutData struct {
	count      int64
	placecount []int64
	instance   int
}

const RPDELTA = 1 // The value has little obvious effect, but 1 is better
// than 0. If 1 is OK, I could just remember the last activity ...

func (pmon *placementMonitor) initBreakoutData() {
	pmon.breakoutData = &breakoutData{
		count:      0,
		placecount: make([]int64, len(pmon.ttinfo.Activities)),
		instance:   0,
	}
}

func (pmon *placementMonitor) radicalStep() bool {
	ttinfo := pmon.ttinfo
	var aix ttbase.ActivityIndex
	var a *ttbase.Activity
	var nposs int
	var dpen Penalty
	var slot ttbase.SlotIndex
	var clashes []ttbase.ActivityIndex
	bdata := pmon.breakoutData

	if bdata.instance == 1 { // value >1 seems to show little or no improvement
		bdata.instance = 0
		return false
	}

	// Seek next possible placement.
	aix = pmon.unplaced[len(pmon.unplaced)-1]
	a = ttinfo.Activities[aix]
	nposs = len(a.PossibleSlots)
	i0 := rand.IntN(nposs)
	i := i0

	//TODO: May want to weight the slots?

	for {
		slot = a.PossibleSlots[i]
		// Check "validity".
		clashes = ttinfo.FindClashes(aix, slot)
		if len(clashes) == 0 {
			// Only accept slots where a replacement is necessary.
			// This seems to be important.
			goto nextslot
		}
		for _, aixx := range clashes {
			if bdata.count-bdata.placecount[aixx] < RPDELTA {
				goto nextslot
			}
		}
		break
	nextslot:
		// Find next possible slot.
		if i < 0 {
			i = i0
		} else {
			i += 1
			if i == nposs {
				i = 0
			}
			if i == i0 {
				// All slots have been tested.
				bdata.count = 0
				return false
			}
		}
	}
	// Place the activity, whatever the penalty.

	bdata.placecount[aix] = bdata.count
	bdata.count++
	bdata.instance++

	// Deplace conflicting activities
	for _, aixx := range clashes {
		ttinfo.UnplaceActivity(aixx)
	}
	// Place new activity
	dpen = pmon.place(aix, slot)
	// Update penalty info
	for _, aixx := range clashes {
		dpen += pmon.evaluate1(aixx)
	}
	for r, p := range pmon.pendingPenalties {
		pmon.resourcePenalties[r] = p
	}
	pmon.score += dpen
	// Remove from "unplaced" list
	pmon.unplaced = pmon.unplaced[:len(pmon.unplaced)-1]
	pmon.unplaced = append(pmon.unplaced, clashes...)
	return true
}
