package fet

import (
	"W365toFET/base"
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
				ags := ttinfo.AtomicGroups[gref]
				for _, ag := range ags {
					subgroups = append(subgroups,
						fetSubgroup{Name: ag.Tag},
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

	//TODO: Further constraints

	/*

			//fmt.Printf("\nCLASS %s: %+v\n", cl.SORTING, cl.DIVISIONS)

			//fmt.Printf("==== %s: %+v\n", cname, nats)

			// Limit gaps on a weekly basis.
			mgpw := 0 //TODO: An additional tweak may be needed for some classes.
			// Handle lunch breaks: The current approach counts lunch breaks as
			// gaps, so the gaps-per-week must be adjusted accordingly.
			if len(lbdays) > 0 {
				// Need lunch break(s).
				// This uses a general "max-lessons-in-interval" constraint.
				// As an alternative, adding dummy lessons (with time constraint)
				// can offer some advantages, like easing gap handling.
				// Set max-gaps-per-week accordingly.
				if lunch_break(fetinfo, &lunchconstraints, cname, lunchperiods) {
					mgpw += len(lbdays)
				}
			}
			// Add the gaps constraint.
			maxgaps = append(maxgaps, maxGapsPerWeek{
				Weight_Percentage: 100,
				Max_Gaps:          mgpw,
				Students:          cname,
				Active:            true,
			})

			// Minimum lessons per day
			mlpd0 := cl.CONSTRAINTS["MinLessonsPerDay"]
			mlpd, err := strconv.Atoi(mlpd0)
			if err != nil {
				base.Error.Fatalf(
					"INVALID MinLessonsPerDay: %s // %v\n", mlpd0, err)
			}
			minlessons = append(minlessons, minLessonsPerDay{
				Weight_Percentage:   100,
				Minimum_Hours_Daily: mlpd,
				Students:            cname,
				Allow_Empty_Days:    false,
				Active:              true,
			})
		}
		fetinfo.fetdata.Time_Constraints_List.
			ConstraintStudentsSetMaxHoursDailyInInterval = lunchconstraints
		fetinfo.fetdata.Time_Constraints_List.
			ConstraintStudentsSetMaxGapsPerWeek = maxgaps
		fetinfo.fetdata.Time_Constraints_List.
			ConstraintStudentsSetMinHoursDaily = minlessons
	*/
}
