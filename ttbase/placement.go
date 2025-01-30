package ttbase

import (
	"W365toFET/base"
	"fmt"
	"slices"
)

//TODO: These functions need modifying to work with ActiviyGroup and TtLesson items

func (ttinfo *TtInfo) initialPlacement() {
	ttplaces := ttinfo.Placements

	//TODO: How much of the following is still needed, and in what form?

	// Collect non-fixed lessons which need placing
	toplace := []LessonUnitIndex{}

	// First place the fixed lessons, then build the PossibleSlots for
	// non-fixed lessons.
	for _, ag := range ttplaces.ActivityGroups {

		for _, lix := range ag.LessonUnits {
			l := ttplaces.TtLessons[lix]
			p := l.Placement
			if p >= 0 {
				if !l.Fixed {
					toplace = append(toplace, lix)
					continue
				}
				if ttinfo.TestPlacement(lix, p) {
					// Perform placement
					ttinfo.PlaceActivity(lix, p)
				} else {
					//TODO?
					//a.XRooms = a.XRooms[:0]
					base.Error.Fatalf(
						"Placement of Fixed Activity %d @ %d failed:\n"+
							"  Course -- %s (or parallel)\n",
						lix, p, ttinfo.View(ttinfo.CourseInfo[ag.Courses[0]]))
				}
			}
		}
	}

	// Build PossibleSlots
	//TODO
	ttinfo.makePossibleSlots()

	// Place non-fixed lessons
	for _, lix := range toplace {
		ttl := ttplaces.TtLessons[lix]
		p := ttl.Placement

		if slices.Contains(ttl.PossibleSlots, p) &&
			ttinfo.TestPlacement(lix, p) {

			// Perform placement
			ttinfo.PlaceActivity(lix, p)
		} else {
			// Need CourseInfo for reporting details

			//TODO: the TtLesson items need a reference to the activity group

			ttl := ttinfo.Activities[aix-1]
			cinfo := ttl.CourseInfo
			//
			base.Warning.Printf(
				"Placement of Activity %d @ %d failed:\n"+
					"  -- %s\n",
				aix, p, ttinfo.View(cinfo))
			a.Placement = -1
			a.XRooms = a.XRooms[:0]
		}
	}

	// Add room choices where possible.
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		a := ttinfo.Activities[aix]
		if len(a.XRooms) != 0 {
			var rnew []ResourceIndex
			p := a.Placement
			for _, rix := range a.XRooms {
				if rix < 0 {
					continue
				}
				slot := rix*ttinfo.SlotsPerWeek + p
				if ttinfo.TtSlots[slot] == 0 {
					ttinfo.TtSlots[slot] = aix
				} else {
					base.Warning.Printf(
						"Lesson in course %s cannot use room %s\n",
						ttinfo.View(a.CourseInfo),
						ttinfo.Resources[rix].(*base.Room).Tag)
					rnew = append(rnew, rix)
				}
			}
			if len(rnew) != 0 {
				a.XRooms = a.XRooms[:len(rnew)]
				copy(a.XRooms, rnew)
			}
		}
	}
}

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
