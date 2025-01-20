package ttbase

import (
	"W365toFET/base"
	"fmt"
	"strings"
)

type AtomicGroup struct {
	// Index is the index of this item within the [TtInfo.Resources] array.
	Index ResourceIndex
	// Class is the reference (Id) of the [base.Class]
	Class Ref
	// Groups lists the references (Id) of the constituent [base.Group]
	// elements
	Groups []Ref
	// Tag is a constructed tag to represent the atomic group
	Tag string
}

// makeAtomicGroups constructs "atomic" groups for the school classes
// [base.Class] elements. An atomic group is an ordered list of single
// groups, one from each division.
func (ttinfo *TtInfo) makeAtomicGroups() {
	db := ttinfo.Db
	ttinfo.AtomicGroups = map[Ref][]*AtomicGroup{}
	// The atomic groups are indexed according to the position they will
	// take in the [TtInfo.Resources] array
	atomicGroupIndex := 0

	// Go through the classes inspecting their Divisions.
	// Build a list-basis for the atomic groups based on the Cartesian product.
	for _, cl := range ttinfo.Db.Classes {
		divs, ok := ttinfo.ClassDivisions[cl.Id]
		if !ok {
			base.Bug.Fatalf("ttinfo.classDivisions[%s]\n", cl.Id)
		}

		if len(divs) == 0 {
			// No divisions, make an atomic group for the class
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
		// The length of the Ref-lists is equal to the number of divisions.
		// The number of Ref-lists is the product of all the division lengths.
		agrefs := [][]Ref{{}}
		for _, dglist := range divs {
			// Add another division, creating a new list of Ref-lists by
			// appending each group of the new division to the old list in
			// turn.
			agrefsx := [][]Ref{} // new list of Ref-lists
			for _, ag := range agrefs {
				// Each of the old Ref-lists is extended
				for _, g := range dglist {
					// Copy the old Ref-list to a new one, adding the new group
					gx := make([]Ref, len(ag)+1)
					copy(gx, append(ag, g))
					// Add to the new list of Rref-lists
					agrefsx = append(agrefsx, gx)
				}
			}
			// Replace the old list of Ref-lists by the new one
			agrefs = agrefsx
		}
		//fmt.Printf("  §§§ Divisions in %s: %+v\n", cl.Tag, divs)
		//fmt.Printf("     --> %+v\n", agrefs)

		// Make [AtomicGroup] elements
		aglist := []*AtomicGroup{}
		for _, ag := range agrefs {
			// Combine the short string representations of the groups
			glist := []string{}
			for _, gref := range ag {
				// Don't use ttinfo.Ref2Tag here because that would provide
				// the full group name (with class prefix)
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

		// Map the individual groups to their atomic groups. This algorithm
		// is described in the program documentation (atomicgroups.md).
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

		// Set up the mapping ttinfo.AtomicGroups, first for the
		// whole-class group, then for the division groups
		ttinfo.AtomicGroups[cl.ClassGroup] = aglist
		for g, agl := range g2ags {
			ttinfo.AtomicGroups[g] = agl
		}
	}

	// For testing: print out the atomic groups for each class
	//ttinfo.PrintAtomicGroups()
}

// For testing
func (ttinfo *TtInfo) PrintAtomicGroups() {
	for _, cl := range ttinfo.Db.Classes {
		agls := []string{}
		for _, ag := range ttinfo.AtomicGroups[cl.ClassGroup] {
			agls = append(agls, ag.Tag)
		}
		fmt.Printf("  ++ %s: %+v\n", ttinfo.Ref2Tag[cl.ClassGroup], agls)
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
