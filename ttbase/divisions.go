package ttbase

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
