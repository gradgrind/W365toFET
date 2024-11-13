package readxml

import (
	"W365toFET/w365tt"
	"fmt"
	"strings"
)

const NO_LUNCH_BREAK = "-Mp"

// Teachers and classes can have a "flag" (implemented here as a Category)
// to signal "no lunch break". By default the element should have a
// lunch break, but if this flag is present the result here will be false.
func withLunchBreak(
	id2node map[w365tt.Ref]any,
	refs RefList,
	nodeId w365tt.Ref,
) bool {
	msg := fmt.Sprintf("Category in Course %s", nodeId)
	reflist := GetRefList(id2node, refs, msg)
	//fmt.Printf("Categories for Teacher or Class %s:\n", nodeId)
	for _, cat := range reflist {
		catnode := id2node[cat].(*Category)
		//fmt.Printf("  :: %+v\n", catnode)
		if catnode.Role == 0 && catnode.Shortcut == NO_LUNCH_BREAK {
			return false
		}
	}
	return true
}

// A class can have a "flag" (implemented here as a Category) to indicate
// that it is not a real class, but a collection of stand-in "lessons".
// Return true if the flag is set.
func isStandIns(
	id2node map[w365tt.Ref]any,
	refs RefList,
	nodeId w365tt.Ref,
) bool {
	msg := fmt.Sprintf("Category in Class %s", nodeId)
	reflist := GetRefList(id2node, refs, msg)
	//fmt.Printf("Categories for Class %s:\n", nodeId)
	for _, cat := range reflist {
		catnode := id2node[cat].(*Category)
		//fmt.Printf("  :: %+v\n", catnode)
		if catnode.Role == 7 {
			return true
		}
	}
	return false
}

// Look for a block tag in the Categories. Only the one will be recognized,
// the first one.
func getBlockTag(
	id2node map[w365tt.Ref]any,
	refs RefList,
	nodeId w365tt.Ref,
) string {
	if refs != "" {
		msg := fmt.Sprintf("Category in Course %s", nodeId)
		reflist := GetRefList(id2node, refs, msg)
		//fmt.Printf("Categories in Course %s:\n", nodeId)
		for _, cat := range reflist {
			catnode := id2node[cat].(*Category)
			//fmt.Printf("  :: %+v\n", catnode)
			if catnode.Role == 0 {
				// If catnode.Shortcut starts with "_", take this as
				// a block, a SuperCourse or a SubCourse:
				if strings.HasPrefix(catnode.Shortcut, "_") {
					return catnode.Shortcut
				}
			}
		}
	}
	return ""
}
