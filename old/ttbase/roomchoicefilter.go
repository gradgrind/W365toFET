package ttbase

import (
	"maps"
)

func roomChoiceFilter(nroomlist []Ref, rclist [][]Ref) VirtualRoom {
	type nothing = struct{}

	necessary := map[Ref]nothing{}
	for _, r := range nroomlist {
		necessary[r] = nothing{}
	}

restart:

	newlist := [][]Ref{}

	cp := []map[Ref]nothing{map[Ref]nothing{}}
	for i, rc0 := range rclist {
		rc := []Ref{}
		for _, r := range rc0 {
			if _, ok := necessary[r]; !ok {
				rc = append(rc, r)
			}
		}
		if len(rc) < 2 {
			if len(rc) == 1 {
				necessary[rc[0]] = nothing{}
				rclist = append(newlist, rclist[i+1:]...)
				goto restart // need to start again ...
			}
			continue
		}

		newlist = append(newlist, rc)

		newcp := []map[Ref]nothing{}
		for _, cpm := range cp {
			for _, r := range rc {
				if _, ok := cpm[r]; !ok {
					newcpm := map[Ref]nothing{r: nothing{}}
					maps.Copy(newcpm, cpm)
					newcp = append(newcp, newcpm)
				}
			}
		}
		cp = newcp

		// Look for "necessary" rooms
		for _, r := range rc {
			for _, cpm := range cp {
				if _, ok := cpm[r]; !ok {
					goto next
				}
			}

			// The room is in all sets
			necessary[r] = nothing{}
			goto restart // need to start again ...

		next:
		}
	}
	rooms := []Ref{}
	for r := range necessary {
		rooms = append(rooms, r)
	}
	return VirtualRoom{
		Rooms:       rooms,
		RoomChoices: newlist,
	}
}
