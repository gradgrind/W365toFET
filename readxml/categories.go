package readxml

import (
	"W365toFET/base"
	"W365toFET/w365tt"
	"strings"
)

const NO_LUNCH_BREAK = "-Mp"

// Teachers and classes can have a "flag" (implemented here as a Category)
// to signal "no lunch break". By default the element should have a
// lunch break, but if this flag is present the result here will be false.
func (cdata *conversionData) withLunchBreak(
	refs RefList, // Category list
	nodeId Ref, // Teacher or Class with this Category list
) bool {
	//fmt.Printf("Categories for Teacher or Class %s:\n", nodeId)
	for _, catref := range splitRefList(refs) {
		cat, ok := cdata.categories[catref]
		if !ok {
			base.Error.Fatalf("Teacher or Class (%s):\n"+
				"  -- Invalid Category: %s", nodeId, catref)
		}
		//fmt.Printf("  :: %+v\n", cat)
		if cat.Shortcut == NO_LUNCH_BREAK {
			return false
		}
	}
	return true
}

// A class can have a "flag" (implemented here as a Category) to indicate
// that it is not a real class, but a collection of stand-in "lessons".
// Return true if the flag is set.
func (cdata *conversionData) isStandIns(
	refs RefList, // Category list
	nodeId w365tt.Ref, // Class
) bool {
	//fmt.Printf("Categories for Class %s:\n", nodeId)
	for _, catref := range splitRefList(refs) {
		cat, ok := cdata.categories[catref]
		if !ok {
			base.Error.Fatalf("Class (%s):\n"+
				"  -- Invalid Category: %s", nodeId, catref)
		}
		//fmt.Printf("  :: %+v\n", cat)
		if (cat.Role & 1) != 0 {
			return true
		}
	}
	return false
}

// Look for a block tag in the Categories. Only one will be recognized,
// the first one.
func (cdata *conversionData) getBlockTag(
	refs RefList,
	nodeId Ref,
) string {
	if refs != "" {
		//fmt.Printf("Categories in Course %s:\n", nodeId)
		for _, catref := range splitRefList(refs) {
			cat, ok := cdata.categories[catref]
			if !ok {
				base.Error.Fatalf("Class (%s):\n"+
					"  -- Invalid Category: %s", nodeId, catref)
			}
			//fmt.Printf("  :: %+v\n", cat)
			// If catnode.Shortcut starts with "_", take this as
			// a block, a SuperCourse or a SubCourse:
			if strings.HasPrefix(cat.Shortcut, "_") {
				return cat.Shortcut
			}
		}
	}
	return ""
}
