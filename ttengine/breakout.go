package ttengine

import (
	"W365toFET/ttbase"
	"math/rand/v2"
)

// TODO: Is there an optimal limit? Too small and it may get trapped too
// easily? Larger values may use a bit more memory and seem slower.
// Around 3 – 5? Actually it is not clear that a value of more than 1 helps
// at all ...
const MAX_BREAKOUT_LEVELS = 1

type breakoutLevel struct {
	state *ttState
	i0    int // index of first slot to test
	i     int // index of last slot tested (-1 if none)
}

func (pmon *placementMonitor) radicalStep(levels *[]*breakoutLevel) bool {
	ttinfo := pmon.ttinfo
	var i0 int
	var i int
	var aix ttbase.ActivityIndex
	var a *ttbase.Activity
	var nposs int
	var dpen Penalty
	var slot ttbase.SlotIndex
	var clashes []ttbase.ActivityIndex
	var level *breakoutLevel
	if len(*levels) < MAX_BREAKOUT_LEVELS {
		// Go to next level.
		level = &breakoutLevel{
			state: pmon.saveState(),
			i0:    -1,
			i:     -1,
		}
		*levels = append(*levels, level)
	} else {
		// Restore state for this level.
		level = (*levels)[len(*levels)-1]
	}
	for {
		// Seek next possible placement.
		pmon.restoreState(level.state)
		aix = pmon.unplaced[len(pmon.unplaced)-1]
		a = ttinfo.Activities[aix]
		nposs = len(a.PossibleSlots)
		i0 = level.i0
		if i0 < 0 {
			i0 = rand.IntN(nposs)
			level.i0 = i0
		}
		i = level.i

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

				// Go up a level.
				nlevel := len(*levels) - 1
				*levels = (*levels)[:nlevel]
				if nlevel == 0 {
					return false
				}
				continue
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
	level.i = i
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
