package fet

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"encoding/xml"
	"strconv"
	"strings"
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

type studentsNotAvailable struct {
	XMLName                       xml.Name `xml:"ConstraintStudentsSetNotAvailableTimes"`
	Weight_Percentage             int
	Students                      string
	Number_of_Not_Available_Times int
	Not_Available_Time            []notAvailableTime
	Active                        bool
}

func getClasses(fetinfo *fetInfo) {
	ttinfo := fetinfo.ttinfo
	items := []fetClass{}
	natimes := []studentsNotAvailable{}
	for _, cl := range ttinfo.Db.Classes {
		cname := cl.Tag
		// Skip "special" classes.
		if cname == "" {
			continue
		}
		divs, ok := ttinfo.ClassDivisions[cl.Id]
		if !ok {
			base.Bug.Fatalf(
				"Class %s has no entry in ttinfo.ClassDivisions\n",
				cname)
		}

		// Construct the Groups and Subgroups
		groups := []fetGroup{}
		for _, div := range divs {
			for _, gref := range div {
				g := ttinfo.Ref2Tag[gref]
				subgroups := []fetSubgroup{}
				for _, agix := range ttinfo.AtomicGroupIndexes[gref] {
					tag := ttinfo.Resources[agix].(*ttbase.AtomicGroup).Tag
					subgroups = append(subgroups,
						fetSubgroup{Name: tag},
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

		// The following constraints don't concern dummy classes ending
		// in "X".
		if strings.HasSuffix(cname, "X") {
			continue
		}

		// "Not available" times.
		nats := []notAvailableTime{}
		day := 0
		for _, na := range cl.NotAvailable {
			if na.Day != day {
				if na.Day < day {
					base.Error.Fatalf(
						"Class %s has unordered NotAvailable times.\n",
						cname)
				}
				day = na.Day
			}
			nats = append(nats,
				notAvailableTime{
					Day: strconv.Itoa(day), Hour: strconv.Itoa(na.Hour)})
		}
		if len(nats) > 0 {
			natimes = append(natimes,
				studentsNotAvailable{
					Weight_Percentage:             100,
					Students:                      cname,
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
				})
		}

	}
	fetinfo.fetdata.Students_List = fetStudentsList{Year: items}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetNotAvailableTimes = natimes
}
