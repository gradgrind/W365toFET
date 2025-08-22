package timetable

import (
	"fmt"
	"slices"
)

func roomChoiceFilter(cinfo *CourseInfo) {
	delta := 0

	necessary := slices.Clone(cinfo.FixedRooms)
	rclist := cinfo.RoomChoices

	// The validity of the rooms in a RoomChoiceGroup has already been checked.
	// They have been ordered while converting to ResourceIndexes.

stage1:
	newlist := [][]ResourceIndex{}
	for i, rc0 := range rclist {
		// Filter out fixed rooms from the choice list
		rc := []ResourceIndex{}
		for _, r := range rc0 {
			if !slices.Contains(necessary, r) {
				rc = append(rc, r)
			}
		}
		if len(rc) >= 2 {
			// Sort the elements.
			slices.Sort(rc)
			fmt.Printf("$%d %v -> %v\n", i, rc0, rc)
			newlist = append(newlist, rc)
			fmt.Printf("(STATE1): [%d, %d, %d] %v\n",
				len(necessary), len(rclist), delta,
				necessary)
		} else if len(rc) == 1 {
			necessary = append(necessary, rc[0])
			fmt.Printf("$%d %v -> %d\n", i, rc0, rc[0])
			fmt.Printf("!!! FIXED %d, REPEATING\n", rc[0])
			rclist = append(newlist, rclist[i+1:]...)
			fmt.Printf("(STATE2): [%d, %d, %d] %v\n",
				len(necessary), len(rclist), delta,
				necessary)
			goto stage1
		} else {
			fmt.Printf("$%d %v -> {}\n", i, rc0)
			fmt.Printf("(STATE3): [%d, %d, %d] %v\n",
				len(necessary), len(rclist), delta,
				necessary)
			delta--
			if delta < 0 {
				fmt.Printf("ERROR: choice list %v has no new rooms\n", rc0)
				return
			}
		}
	}

	fmt.Printf("*******>>> %d %d %d\n", len(necessary), len(newlist), delta)

	// Now build the Cartesian product of the choice lists, omitting
	// values with duplicate rooms and duplicate values generally.
	cp := [][]ResourceIndex{{}} // build Cartesian product values here
	for i, rc := range newlist {
		// Add next choice list, extending the entries in `cp`
		newcp := [][]ResourceIndex{} // build new `cp` here
		for _, cp0 := range cp {     // for each C-p value
			for _, r := range rc { // add each room in current choice list
				if !slices.Contains(cp0, r) { // ... if not a duplicate
					cp1 := append(slices.Clone(cp0), r)
					//fmt.Printf("???4: %v\n", cp1)

					// ... and if the new C-p value is not a duplicate
					slices.Sort(cp1)
					//fmt.Printf("???5: %v\n", cp1)
					for _, cp2 := range newcp {
						if slices.Equal(cp2, cp1) {
							goto next
						}
					}
					newcp = append(newcp, cp1)
					//fmt.Printf("???6: %v\n", newcp)
				next:
				}
			}
		}

		if len(newcp) == 0 {
			fmt.Printf("!!!!! newcp empty: %v\n", rc)
			return
		} else {
			do_restart := false
			for _, r := range rc {
				for _, cp0 := range newcp {
					if !slices.Contains(cp0, r) {
						goto next2
					}
				}
				// r is in all combinations
				necessary = append(necessary, r)
				delta++
				fmt.Printf("!!! FIXED %d\n", r)
				do_restart = true
			next2:
			}
			if do_restart {
				fmt.Println(" ... restarting")
				rclist = newlist
				goto stage1
			}
		}

		cp = newcp
		fmt.Printf("???7: %d – %v\n", i, len(newcp))
	}

	slices.Sort(necessary)
	cinfo.FixedRooms = necessary
	cinfo.RoomChoices = newlist

	fmt.Printf("\n $$ NECESSARY: %v\n\n", necessary)
	for i, rc := range newlist {
		fmt.Printf("*** %d: %v\n", i, rc)
	}
	fmt.Printf("\n delta: %d\n", delta)
}
