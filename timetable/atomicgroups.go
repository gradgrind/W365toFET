package timetable

import (
	"strings"
)

const ATOMIC_GROUP_SEP1 = "#"
const ATOMIC_GROUP_SEP2 = "~"

// Prepare filtered versions of the class Divisions containing only
// those Divisions which have Groups used in Lessons.
func (tt_shared_data *TtSharedData) FilterDivisions() {
	db := tt_shared_data.Db

	// Collect groups used in courses
	usedgroups := map[NodeRef]bool{}

	// Gather groups from the SuperCourses.
	for _, spc := range db.SuperCourses {
		for _, sbc := range spc.SubCourses {
			for _, gref := range sbc.Groups {
				usedgroups[gref] = true
			}
		}
	}
	// Gather groups from the plain Courses.
	for _, c := range db.Courses {
		for _, gref := range c.Groups {
			usedgroups[gref] = true
		}
	}

	// Filter the class divisions, discarding the division names.
	for _, c := range tt_shared_data.Db.Classes {
		divs := [][]NodeRef{}
		for _, div := range c.Divisions {
			for _, gref := range div.Groups {
				if usedgroups[gref] {
					divs = append(divs, div.Groups)
					break
				}
			}
		}
		tt_shared_data.ClassDivisions = append(tt_shared_data.ClassDivisions,
			ClassDivision{c, divs})
	}
}

// TODO: Do I actually need the Index field?
type AtomicGroup struct {
	//Index  ResourceIndex
	Class  NodeRef
	Groups []NodeRef
	Tag    string // A constructed tag to represent the atomic group
}

func (a *AtomicGroup) GetResourceTag() string {
	return a.Tag
}

func (tt_shared_data *TtSharedData) MakeAtomicGroups() {
	// An atomic group is an ordered list of single groups, one from each
	// division.
	tt_shared_data.AtomicGroups = map[NodeRef][]ResourceIndex{}
	db := tt_shared_data.Db

	// Go through the classes inspecting their Divisions.
	// Build a list-basis for the atomic groups based on the Cartesian product.
	for _, cdivs := range tt_shared_data.ClassDivisions {
		cl := cdivs.Class
		if len(cdivs.Divisions) == 0 {
			// Make an atomic group for the class
			agix := len(tt_shared_data.Resources)
			ag := &AtomicGroup{
				//Index: agix,
				Class: cl.Id,
				Tag:   cl.Tag + ATOMIC_GROUP_SEP1,
			}
			tt_shared_data.Resources = append(tt_shared_data.Resources, ag)
			tt_shared_data.AtomicGroups[cl.ClassGroup] = []ResourceIndex{agix}
			continue
		}

		// The atomic groups will be built as a list of lists of Refs.
		agrefs := [][]NodeRef{{}}
		for _, dglist := range cdivs.Divisions {
			// Add another division – increases underlying list lengths.
			agrefsx := [][]NodeRef{}
			for _, ag := range agrefs {
				// Extend each of the old list items by appending each
				// group of the new division in turn – multiplies the
				// total number of atomic groups.
				newlen := len(ag) + 1
				for _, g := range dglist {
					gx := make([]NodeRef, newlen)
					copy(gx, append(ag, g))
					agrefsx = append(agrefsx, gx)
				}
			}
			agrefs = agrefsx
		}
		//fmt.Printf("  §§§ Divisions in %s: %+v\n", cl.Tag, divs)
		//fmt.Printf("     --> %+v\n", agrefs)

		// Make AtomicGroups
		aglist := []ResourceIndex{}
		for _, ag := range agrefs {
			glist := []string{}
			for _, gref := range ag {
				gtag := db.Ref2Tag(gref)
				glist = append(glist, gtag)
			}
			agix := len(tt_shared_data.Resources)
			ag := &AtomicGroup{
				//Index:  agix,
				Class:  cl.Id,
				Groups: ag,
				Tag: cl.Tag + ATOMIC_GROUP_SEP1 +
					strings.Join(glist, ATOMIC_GROUP_SEP2),
			}
			tt_shared_data.Resources = append(tt_shared_data.Resources, ag)
			aglist = append(aglist, agix)
		}
		tt_shared_data.AtomicGroups[cl.ClassGroup] = aglist
		// Map the individual groups to their atomic groups.
		count := 1
		divIndex := len(cdivs.Divisions)
		for divIndex > 0 {
			divIndex--
			divGroups := cdivs.Divisions[divIndex]
			agi := 0 // ag index
			for agi < len(aglist) {
				for _, g := range divGroups {
					for j := 0; j < count; j++ {
						tt_shared_data.AtomicGroups[g] = append(
							tt_shared_data.AtomicGroups[g], aglist[agi])
						agi++
					}
				}
			}
			count *= len(divGroups)
		}
	}
}

/*TODO
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
*/
