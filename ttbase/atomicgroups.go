package ttbase

import (
	"W365toFET/base"
	"fmt"
	"strings"
)

// "Atomic Groups" are needed especially for the class handling.
// They should only be built for divisions which have lessons.
// So first the Lessons must be consulted for their Courses
// and thus their groups – which can then be marked. Finally the divisions
// can be filtered on the basis of these marked groups.

type AtomicGroup struct {
	Class  Ref
	Groups []Ref
	Tag    string
}

func filterDivisions(ttinfo *TtInfo) {
	// Prepare filtered versions of the class Divisions containing only
	// those Divisions which have Groups used in Lessons.

	// Collect groups used in Lessons. Get them from the
	// ttinfo.courseInfo.groups map, which only includes courses with lessons.
	usedgroups := map[Ref]bool{}
	for _, cinfo := range ttinfo.courseInfo {
		for _, g := range cinfo.Groups {
			usedgroups[g] = true
		}
	}
	// Filter the class divisions, discarding the division names.
	cdivs := map[Ref][][]Ref{}
	for _, c := range ttinfo.db.Classes {
		divs := [][]Ref{}
		for _, div := range c.Divisions {
			for _, gref := range div.Groups {
				if usedgroups[gref] {
					divs = append(divs, div.Groups)
					break
				}
			}
		}
		cdivs[c.Id] = divs
	}
	ttinfo.classDivisions = cdivs
}

func makeAtomicGroups(ttinfo *TtInfo) {
	// An atomic group is an ordered list of single groups from each division.
	ttinfo.atomicGroups = map[Ref][]AtomicGroup{}
	// Go through the classes inspecting their Divisions. Retain only those
	// which have lessons.
	filterDivisions(ttinfo) // -> ttinfo.classDivisions
	// Go through the classes inspecting their Divisions.
	// Build a list-basis for the atomic groups based on the Cartesian product.
	for _, cl := range ttinfo.db.Classes {
		divs, ok := ttinfo.classDivisions[cl.Id]
		if !ok {
			base.Bug.Fatalf("ttinfo.classDivisions[%s]\n", cl.Id)
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
		aglist := []AtomicGroup{}
		for _, ag := range agrefs {
			glist := []string{}
			for _, gref := range ag {
				glist = append(glist, ttinfo.ref2grouponly[gref])
			}
			ago := AtomicGroup{
				Class:  cl.Id,
				Groups: ag,
				Tag: cl.Tag + ATOMIC_GROUP_SEP1 +
					strings.Join(glist, ATOMIC_GROUP_SEP2),
			}
			aglist = append(aglist, ago)
		}

		// Map the individual groups to their atomic groups.
		g2ags := map[Ref][]AtomicGroup{}
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
		if len(divs) != 0 {
			ttinfo.atomicGroups[cl.Id] = aglist
			for g, agl := range g2ags {
				ttinfo.atomicGroups[g] = agl
			}
		} else {
			ttinfo.atomicGroups[cl.Id] = []AtomicGroup{}
		}
	}
}

// For testing
func printAtomicGroups(ttinfo *TtInfo) {
	for _, cl := range ttinfo.db.Classes {
		agls := []string{}
		for _, ag := range ttinfo.atomicGroups[cl.Id] {
			agls = append(agls, ag.Tag)
		}
		fmt.Printf("  ++ %s: %+v\n", ttinfo.ref2tt[cl.Id], agls)
		for _, div := range ttinfo.classDivisions[cl.Id] {
			for _, g := range div {
				agls := []string{}
				for _, ag := range ttinfo.atomicGroups[g] {
					agls = append(agls, ag.Tag)
				}
				fmt.Printf("    -- %s: %+v\n", ttinfo.ref2tt[g], agls)
			}
		}
	}
}
