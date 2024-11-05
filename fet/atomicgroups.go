package fet

import (
	"fmt"
	"log"
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

func filterDivisions(fetinfo *fetInfo) {
	// Prepare filtered versions of the class Divisions containing only
	// those Divisions which have Groups used in Lessons.

	// Collect groups used in Lessons. Get them from the
	// fetinfo.courseInfo.groups map, which only includes courses with lessons.
	usedgroups := map[Ref]bool{}
	for _, cinfo := range fetinfo.courseInfo {
		for _, g := range cinfo.groups {
			usedgroups[g] = true
		}
	}
	// Filter the class divisions, discarding the division names.
	cdivs := map[Ref][][]Ref{}
	for _, c := range fetinfo.db.Classes {
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
	fetinfo.classDivisions = cdivs
}

func makeAtomicGroups(fetinfo *fetInfo) {
	// An atomic group is an ordered list of single groups from each division.
	fetinfo.atomicGroups = map[Ref][]AtomicGroup{}
	// Go through the classes inspecting their Divisions. Retain only those
	// which have lessons.
	filterDivisions(fetinfo) // -> fetinfo.classDivisions
	// Go through the classes inspecting their Divisions.
	// Build a list-basis for the atomic groups based on the Cartesian product.
	for _, cl := range fetinfo.db.Classes {
		divs, ok := fetinfo.classDivisions[cl.Id]
		if !ok {
			log.Fatalf("*BUG* fetinfo.classDivisions[%s]\n", cl.Id)
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
				glist = append(glist, fetinfo.ref2grouponly[gref])
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
			fetinfo.atomicGroups[cl.Id] = aglist
			for g, agl := range g2ags {
				fetinfo.atomicGroups[g] = agl
			}
		} else {
			fetinfo.atomicGroups[cl.Id] = []AtomicGroup{}
		}
	}
}

// For testing
func printAtomicGroups(fetinfo *fetInfo) {
	for _, cl := range fetinfo.db.Classes {
		agls := []string{}
		for _, ag := range fetinfo.atomicGroups[cl.Id] {
			agls = append(agls, ag.Tag)
		}
		fmt.Printf("  ++ %s: %+v\n", fetinfo.ref2fet[cl.Id], agls)
		for _, div := range fetinfo.classDivisions[cl.Id] {
			for _, g := range div {
				agls := []string{}
				for _, ag := range fetinfo.atomicGroups[g] {
					agls = append(agls, ag.Tag)
				}
				fmt.Printf("    -- %s: %+v\n", fetinfo.ref2fet[g], agls)
			}
		}
	}
}
