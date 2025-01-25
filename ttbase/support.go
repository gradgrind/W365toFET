package ttbase

import "fmt"

// DEBUGGING only
func (ttinfo *TtInfo) CheckResourceIntegrity() {
	for rix := 0; rix < len(ttinfo.Resources); rix++ {
		slot0 := rix * ttinfo.SlotsPerWeek
		for p := 0; p < ttinfo.SlotsPerWeek; p++ {
			aix := ttinfo.TtSlots[slot0+p]
			if aix <= 0 {
				continue
			}
			a := ttinfo.Activities[aix]
			ap := a.Placement
			if ap < 0 {
				panic(fmt.Sprintf("Resource (%d) of unplaced Activity (%d)"+
					" at position %d:", rix, aix, ap))
			}
			for i := 0; i < a.Duration; i++ {
				if ap+i == p {
					goto pok
				}
			}
			panic(fmt.Sprintf("Resource (%d) of Activity (%d)"+
				" at wrong position %d (should be %d):",
				rix, aix, p, ap))
		pok:
		}
	}
}
