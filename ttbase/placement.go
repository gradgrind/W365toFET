package ttbase

import (
	"W365toFET/base"
	"fmt"
	"slices"
)

// TODO: Can I safely assume that no attempt will be made to unplace fixed
// Activities?
func (ttinfo *TtInfo) UnplaceActivity(aix ActivityIndex) {
	a := ttinfo.Activities[aix]
	slot := a.Placement

	//TODO--- for testing
	if a.Fixed {
		base.Bug.Fatalf("Can't unplace %d – fixed\n", aix)
	}
	if slot < 0 {
		base.Bug.Printf("Can't unplace %d – not placed\n", aix)
		panic(1)
		return
	}

	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			ttinfo.TtSlots[i+ix] = 0
		}
	}
	for _, rix := range a.XRooms {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			ttinfo.TtSlots[i+ix] = 0
		}
	}
	a.Placement = -1

	for _, aixp := range a.Parallel {
		a := ttinfo.Activities[aixp]
		for _, rix := range a.Resources {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				ttinfo.TtSlots[i+ix] = 0
			}
		}
		for _, rix := range a.XRooms {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				ttinfo.TtSlots[i+ix] = 0
			}
		}
		a.Placement = -1
	}
	//--ttinfo.CheckResourceIntegrity()
}

// Note that – at present – testPlacement, findClashes and placeActivity
// don't try to place room choices. This is intentional, assuming that these
// will be placed by other functions ...

func (ttinfo *TtInfo) TestPlacement(aix ActivityIndex, slot int) bool {
	// Simple boolean placement test. It assumes the slot is possible –
	// so that it will not, for example, be the last slot of a day if
	// the activity duration is 2.
	a := ttinfo.Activities[aix]
	day := slot / ttinfo.DayLength
	for _, addix := range a.DifferentDays {
		add := ttinfo.Activities[addix]
		if add.Placement >= 0 && add.Placement/ttinfo.DayLength == day {
			return false
		}
	}
	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			if ttinfo.TtSlots[i+ix] != 0 {
				return false
			}
		}
	}
	for _, aixp := range a.Parallel {
		a := ttinfo.Activities[aixp]
		for _, addix := range a.DifferentDays {
			add := ttinfo.Activities[addix]
			if add.Placement >= 0 && add.Placement/ttinfo.DayLength == day {
				return false
			}
		}
		for _, rix := range a.Resources {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				if ttinfo.TtSlots[i+ix] != 0 {
					return false
				}
			}
		}
	}
	return true
}

func (ttinfo *TtInfo) PlaceActivity(aix ActivityIndex, slot int) {
	// Allocate the resources, assuming none of the slots are blocked!
	//--fmt.Printf("++++++++ PLACE ++++++++ %d: %d\n", aix, slot)
	a := ttinfo.Activities[aix]

	//TODO-- This is for debugging
	p := a.Placement
	if p >= 0 && p != slot {
		fmt.Printf("::::: %+v\n", a)
		panic(fmt.Sprintf("Activity %d already placed: %d\n", aix, p))
	}
	//

	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			ttinfo.TtSlots[i+ix] = aix
		}
	}
	a.Placement = slot

	for _, aixp := range a.Parallel {
		a := ttinfo.Activities[aixp]
		for _, rix := range a.Resources {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				ttinfo.TtSlots[i+ix] = aixp
			}
		}
		a.Placement = slot
	}
	//--ttinfo.CheckResourceIntegrity()
}

func (ttinfo *TtInfo) FindClashes(aix ActivityIndex, slot int) []ActivityIndex {
	// Return a list of activities (indexes) which are in conflict with
	// the proposed placement. It assumes the slot is in principle possible –
	// so that it will not, for example, be the last slot of a day if
	// the activity duration is 2.
	clashes := []ActivityIndex{}
	a := ttinfo.Activities[aix]
	day := slot / ttinfo.DayLength
	//--fmt.Printf("????0 aix: %d slot %d\n", aix, slot)
	for _, addix := range a.DifferentDays {
		add := ttinfo.Activities[addix]
		if add.Placement >= 0 && add.Placement/ttinfo.DayLength == day {
			clashes = append(clashes, addix)
			//--fmt.Printf("????1 %d\n", addix)
		}
	}
	for _, rix := range a.Resources {
		i := rix*ttinfo.SlotsPerWeek + slot
		for ix := 0; ix < a.Duration; ix++ {
			c := ttinfo.TtSlots[i+ix]
			if c != 0 {
				//--xxx := ttinfo.Activities[c].Placement
				clashes = append(clashes, c)
				//--fmt.Printf("????2 %d %d r: %d p: %d\n", c, ix, rix, xxx)
			}
		}
	}
	for _, aixp := range a.Parallel {
		a := ttinfo.Activities[aixp]
		for _, addix := range a.DifferentDays {
			add := ttinfo.Activities[addix]
			if add.Placement >= 0 && add.Placement/ttinfo.DayLength == day {
				clashes = append(clashes, addix)
				//--fmt.Printf("????3 %d\n", addix)
			}
		}
		for _, rix := range a.Resources {
			i := rix*ttinfo.SlotsPerWeek + slot
			for ix := 0; ix < a.Duration; ix++ {
				c := ttinfo.TtSlots[i+ix]
				if c != 0 {
					clashes = append(clashes, c)
					//--fmt.Printf("????4 %d %d\n", c, ix)
				}
			}
		}
	}
	slices.Sort(clashes)
	return slices.Compact(clashes)
}
