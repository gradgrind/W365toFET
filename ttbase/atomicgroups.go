package ttbase

import (
	"W365toFET/base"
	"fmt"
	"strings"
)

type ResourceIndex int

type AtomicGroup struct {
	Index  ResourceIndex
	Class  Ref
	Groups []Ref
	Tag    string // A constructed tag to represent the atomic group
}

func makeAtomicGroups(ttinfo *TtInfo) {
	// An atomic group is an ordered list of single groups, one from each
	// division.
	db := ttinfo.Db
	ttinfo.AtomicGroups = map[Ref][]*AtomicGroup{}
	atomicGroupIndex := ResourceIndex(0)

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

// For testing
func printAtomicGroups(ttinfo *TtInfo) {
	for _, cl := range ttinfo.Db.Classes {
		agls := []string{}
		for _, ag := range ttinfo.AtomicGroups[cl.Id] {
			agls = append(agls, ag.Tag)
		}
		fmt.Printf("  ++ %s: %+v\n", ttinfo.Ref2Tag[cl.Id], agls)
		for _, div := range ttinfo.ClassDivisions[cl.Id] {
			for _, g := range div {
				agls := []string{}
				for _, ag := range ttinfo.AtomicGroups[g] {
					agls = append(agls, ag.Tag)
				}
				fmt.Printf("    -- %s: %+v\n", ttinfo.Ref2Tag[g], agls)
			}
		}
	}
}
