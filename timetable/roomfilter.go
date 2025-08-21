package ttbase

import (
	"W365toFET/base"
	"maps"
	"strings"
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

func filterRooms(
	fixedrooms []ResourceIndex,
	roomchoices [][]ResourceIndex,
) {
	// Go through the classes inspecting their Divisions.
	// Build a list-basis for the atomic groups based on the Cartesian product.
	for _, cl := range ttinfo.Db.Classes {
		divs, ok := ttinfo.ClassDivisions[cl.Id]
		if !ok {
			base.Bug.Fatalf("ttinfo.classDivisions[%s]\n", cl.Id)
		}

		if len(divs) == 0 {
			// Make an atomic group for the class
			cag := &AtomicGroup{
				Index: atomicGroupIndex,
				Class: cl.Id,
				Tag:   cl.Tag + ATOMIC_GROUP_SEP1,
			}
			atomicGroupIndex++
			ttinfo.AtomicGroups[cl.ClassGroup] = []*AtomicGroup{cag}
			continue
		}

		// The atomic groups will be built as a list of lists of Refs.
		agrefs := [][]Ref{{}}
		for _, dglist := range divs {
			// Add another division – increases underlying list lengths.
			agrefsx := [][]Ref{}
			for _, ag := range agrefs {
				// Extend each of the old list items by appending each
				// group of the new division in turn – multiplies the
				// total number of atomic groups.
				for _, g := range dglist {
					gx := make([]Ref, len(ag)+1)
					copy(gx, append(ag, g))
					agrefsx = append(agrefsx, gx)
				}
			}
			agrefs = agrefsx
		}
		//fmt.Printf("  §§§ Divisions in %s: %+v\n", cl.Tag, divs)
		//fmt.Printf("     --> %+v\n", agrefs)

		// Make AtomicGroups
		aglist := []*AtomicGroup{}
		for _, ag := range agrefs {
			glist := []string{}
			for _, gref := range ag {
				gtag := db.Elements[gref].(*base.Group).Tag
				glist = append(glist, gtag)
			}
			ago := &AtomicGroup{
				Index:  atomicGroupIndex,
				Class:  cl.Id,
				Groups: ag,
				Tag: cl.Tag + ATOMIC_GROUP_SEP1 +
					strings.Join(glist, ATOMIC_GROUP_SEP2),
			}
			atomicGroupIndex++
			aglist = append(aglist, ago)
		}

		// Map the individual groups to their atomic groups.
		g2ags := map[Ref][]*AtomicGroup{}
		count := 1
		divIndex := len(divs)
		for divIndex > 0 {
			divIndex--
			divGroups := divs[divIndex]
			agi := 0 // ag index
			for agi < len(aglist) {
				for _, g := range divGroups {
					for j := 0; j < count; j++ {
						g2ags[g] = append(g2ags[g], aglist[agi])
						agi++
					}
				}
			}
			count *= len(divGroups)
		}

		ttinfo.AtomicGroups[cl.ClassGroup] = aglist
		for g, agl := range g2ags {
			ttinfo.AtomicGroups[g] = agl
		}
	}
}
