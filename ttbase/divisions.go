package ttbase

import (
	"W365toFET/base"
)

func filterDivisions(ttinfo *TtInfo) {
	// Prepare filtered versions of the class Divisions containing only
	// those Divisions which have Groups used in Lessons.

	// Collect groups used in Lessons. Get them from the
	// ttinfo.courseInfo.groups map, which only includes courses with lessons.
	usedgroups := map[Ref]bool{}
	for _, cinfo := range ttinfo.CourseInfo {
		for _, g := range cinfo.Groups {
			usedgroups[g] = true
		}
	}
	// Filter the class divisions, discarding the division names.
	cdivs := map[Ref][][]Ref{}
	for _, c := range ttinfo.Db.Classes {
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
	ttinfo.ClassDivisions = cdivs
}

type FractionChip struct {
	Groups      []string
	ExtraGroups []string
	Fraction    int
	Offset      int
	Total       int
}

func (ttinfo *TtInfo) SortClassGroups(
	class Ref,
	groups []Ref,
) []FractionChip {
	// Given a class and a list of groups, separate those groups from the
	// given class.
	// Also build corresponding "fractional chip" info.
	db := ttinfo.Db
	elements := db.Elements
	mygroups := map[Ref]bool{}
	othergroups := []Ref{}
	for _, gref := range groups {
		g := elements[gref].(*base.Group)
		if g.Class == class {
			if g.Tag == "" {
				// whole class
				mygroups[""] = true
			} else {
				mygroups[gref] = true
			}
		} else {
			othergroups = append(othergroups, gref)
		}
	}
	xgroups := ttinfo.SortList(othergroups)
	chips := []FractionChip{}

	if mygroups[""] {
		chips = append(chips, FractionChip{
			Groups:      []string{},
			ExtraGroups: xgroups,
			Fraction:    1,
			Offset:      0,
			Total:       1,
		})
	} else {
		for _, div := range ttinfo.ClassDivisions[class] {
			start := -1
			var glist []string = nil
			for i, gref := range div {
				if mygroups[gref] {
					if glist == nil {
						start = i
					}
					glist = append(glist, elements[gref].(*base.Group).Tag)
					delete(mygroups, gref)
				} else if glist != nil {
					chips = append(chips, FractionChip{
						Groups:      glist,
						ExtraGroups: xgroups,
						Fraction:    len(glist),
						Offset:      start,
						Total:       len(div),
					})
					glist = nil
				}
			}
			if glist != nil {
				chips = append(chips, FractionChip{
					Groups:      glist,
					ExtraGroups: xgroups,
					Fraction:    len(glist),
					Offset:      start,
					Total:       len(div),
				})
			}
			if start >= 0 {
				// Groups from this division were found
				if len(mygroups) == 0 {
					break
				} else {
					// There are groups from other divisions, too.
					// Return the whole class.
					chips = nil
					chips = append(chips, FractionChip{
						Groups:      []string{},
						ExtraGroups: xgroups,
						Fraction:    1,
						Offset:      0,
						Total:       1,
					})
				}
			}
		}
	}
	return chips
}
