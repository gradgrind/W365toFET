package fet

import (
	"W365toFET/base"
	"encoding/xml"
)

type fetCategory struct {
	//XMLName             xml.Name `xml:"Category"`
	Number_of_Divisions int
	Division            []string
}

type fetSubgroup struct {
	Name string // 13.m.MaE
	//Number_of_Students int // 0
	//Comments string // ""
}

type fetGroup struct {
	Name string // 13.K
	//Number_of_Students int // 0
	//Comments string // ""
	Subgroup []fetSubgroup
}

type fetClass struct {
	//XMLName  xml.Name `xml:"Year"`
	Name      string
	Long_Name string
	Comments  string
	//Number_of_Students int (=0)
	// The information regarding categories, divisions of each category,
	// and separator is only used in the dialog to divide the year
	// automatically by categories.
	Number_of_Categories int
	Separator            string // CLASS_GROUP_SEP
	Category             []fetCategory
	Group                []fetGroup
}

type fetStudentsList struct {
	XMLName xml.Name `xml:"Students_List"`
	Year    []fetClass
}

func getClasses(fetinfo *fetInfo) {
	tt_data := fetinfo.tt_data
	shared_data := tt_data.SharedData
	db := shared_data.Db
	items := []fetClass{}
	for _, cdiv := range shared_data.ClassDivisions {
		cl := cdiv.Class
		cname := cl.Tag
		// Skip "special" classes.
		if cname == "" {
			continue
		}
		divs := cdiv.Divisions
		// Construct the Groups and Subgroups
		groups := []fetGroup{}
		for _, div := range divs {
			for _, gref := range div {

				// Need to construct group name with class, group
				// and CLASS_GROUP_SEP
				g := fetGroupTag(db.Elements[gref].(*base.Group))

				subgroups := []fetSubgroup{}
				ags := shared_data.AtomicGroups[gref]
				for _, ag := range ags {
					subgroups = append(subgroups,
						fetSubgroup{
							Name: shared_data.Resources[ag].GetResourceTag()},
					)
				}
				groups = append(groups, fetGroup{
					Name:     g,
					Subgroup: subgroups,
				})
			}
		}

		// Construct the "Categories" (divisions)
		categories := []fetCategory{}
		for _, divl := range divs {
			strcum := []string{}
			for _, i := range divl {
				strcum = append(strcum, fetinfo.ref2grouponly[i])
			}
			categories = append(categories, fetCategory{
				Number_of_Divisions: len(divl),
				Division:            strcum,
			})
		}
		items = append(items, fetClass{
			Name:                 cname,
			Long_Name:            cl.Name,
			Separator:            CLASS_GROUP_SEP,
			Number_of_Categories: len(categories),
			Category:             categories,
			Group:                groups,
		})
	}
	fetinfo.fetdata.Students_List = fetStudentsList{Year: items}
}

// In FET the group identifier is constructed from the class tag,
// CLASS_GROUP_SEP and the group tag. However, if the group is the
// whole class, just the class tag is used.
func fetGroupTag(g *base.Group) string {
	gt := g.Class.Tag
	if g.Tag != "" {
		gt += CLASS_GROUP_SEP + g.Tag
	}
	return gt
}
