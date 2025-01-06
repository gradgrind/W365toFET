package ttengine

import (
	"W365toFET/ttbase"
	"math/rand/v2"
)

type breakoutData struct {
	state *ttState
	i0    int // index of first slot to test
	i     int // index of last slot tested (-1 if none)
}

func (pmon *placementMonitor) initBreakoutData() {
	pmon.breakoutData = &breakoutData{
		i0: -1,
	}
}

func (pmon *placementMonitor) radicalStep() bool {
	ttinfo := pmon.ttinfo
	var i0 int
	var i int
	var aix ttbase.ActivityIndex
	var a *ttbase.Activity
	var nposs int
	var dpen Penalty
	var slot ttbase.SlotIndex
	var clashes []ttbase.ActivityIndex
	bdata := pmon.breakoutData
	if bdata.i0 < 0 {
		// The mechanism needs (re)initializing.
		bdata.state = pmon.saveState()
		bdata.i = -1
	} else {
		// Restore breakout state?
		pmon.restoreState(bdata.state)
	}
	for {
		// Seek next possible placement.
		aix = pmon.unplaced[len(pmon.unplaced)-1]
		a = ttinfo.Activities[aix]
		nposs = len(a.PossibleSlots)
		i0 = bdata.i0
		if i0 < 0 {
			i0 = rand.IntN(nposs)
			bdata.i0 = i0
		}
		i = bdata.i

		// Place the activity, whatever the penalty.
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
				bdata.i0 = -1
				return false
			}
		}
		slot = a.PossibleSlots[i]
		// Check "validity".
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
		// Continue with this slot
		break
	}
	// Remember last used index
	bdata.i = i
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
