package fet

import (
	"W365toFET/base"
	"encoding/xml"
	"slices"
	"strconv"
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
	db := tt_data.Db
	items := []fetClass{}
	for _, cdiv := range tt_data.ClassDivisions {
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
				ags := tt_data.AtomicGroups[gref]
				for _, ag := range ags {
					subgroups = append(subgroups,
						fetSubgroup{
							Name: tt_data.Resources[ag].GetResourceTag()},
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

/* Lunch-breaks

Lunch-breaks can be done using max-hours-in-interval constraint, but that
makes specification of max-gaps more difficult (becuase the lunch breaks
count as gaps).

The alternative is to add dummy lessons, clamped to the midday-break hours,
on the days where none of the midday-break hours are blocked. However, this
can also cause problems with gaps – the dummy lesson can itself create gaps,
for example when a class only has lessons earlier in the day.

Tests with the dummy lessons approach suggest that it is difficult to get the
number of these lessons and their placement on the correct days right.

This is an attempt with the max-hours-in-interval constraint.

*/

//TODO: Do I need to deal with this? Just for not available times?
// Some constraints don't concern dummy classes ending in "X".

func addClassConstraints(fetinfo *fetInfo) {
	natimes := []studentsNotAvailable{}
	cminlpd := []minLessonsPerDay{}
	cmaxlpd := []maxLessonsPerDay{}
	cmaxgpd := []maxGapsPerDay{}
	cmaxgpw := []maxGapsPerWeek{}
	cmaxaft := []maxDaysinIntervalPerWeek{}
	cmaxls := []maxLateStarts{}
	clblist := []lunchBreak{}
	tt_data := fetinfo.tt_data
	ndays := tt_data.NDays
	nhours := tt_data.NHours
	db := tt_data.Db

	for _, cl := range db.Classes {
		if cl.Tag == "" {
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
						cl.Tag)
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
					Students:                      cl.Tag,
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
				})
		}

		n := cl.MinLessonsPerDay
		if n >= 2 && n <= nhours {
			cminlpd = append(cminlpd, minLessonsPerDay{
				Weight_Percentage:   100,
				Students:            cl.Tag,
				Minimum_Hours_Daily: n,
				Allow_Empty_Days:    true,
				Active:              true,
			})
		}

		n = cl.MaxLessonsPerDay
		if n >= 0 && n < nhours {
			cmaxlpd = append(cmaxlpd, maxLessonsPerDay{
				Weight_Percentage:   100,
				Students:            cl.Tag,
				Maximum_Hours_Daily: n,
				Active:              true,
			})
		}

		i := db.Info.FirstAfternoonHour
		maxpm := cl.MaxAfternoons
		if maxpm >= 0 && i > 0 {
			cmaxaft = append(cmaxaft, maxDaysinIntervalPerWeek{
				Weight_Percentage:   100,
				Students:            cl.Tag,
				Interval_Start_Hour: strconv.Itoa(i),
				Interval_End_Hour:   "", // end of day
				Max_Days_Per_Week:   maxpm,
				Active:              true,
			})
		}

		if cl.ForceFirstHour {
			cmaxls = append(cmaxls, maxLateStarts{
				Weight_Percentage:             100,
				Max_Beginnings_At_Second_Hour: 0,
				Students:                      cl.Tag,
				Active:                        true,
			})
		}

		// The lunch-break constraint may require adjustment of these:
		mgpday := cl.MaxGapsPerDay
		mgpweek := cl.MaxGapsPerWeek
		if mgpweek < 0 {
			mgpweek = 0
		}

		if mbhours := db.Info.MiddayBreak; len(mbhours) != 0 && cl.LunchBreak {
			// Generate the constraint unless all days have a blocked lesson
			// at lunchtime.
			lbdays := ndays
			d := 0
			for _, ts := range cl.NotAvailable {
				if ts.Day < d {
					continue
				}
				if slices.Contains(mbhours, ts.Hour) {
					lbdays--
					d = ts.Day + 1
				}
			}
			if lbdays != 0 {
				// Add a lunch-break constraint.
				clblist = append(clblist, lunchBreak{
					Weight_Percentage:   100,
					Students:            cl.Tag,
					Interval_Start_Hour: strconv.Itoa(mbhours[0]),
					Interval_End_Hour:   strconv.Itoa(mbhours[0] + len(mbhours)),
					Maximum_Hours_Daily: len(mbhours) - 1,
					Active:              true,
				})
				//fmt.Printf("%s:: lbdays: %d maxpm: %d\n",
				//  cl.Tag, lbdays, maxpm)
				// Adjust gaps
				if maxpm < lbdays {
					lbdays = maxpm
				}
				if mgpday == 0 {
					mgpday = 1
				}
				if mgpweek >= 0 {
					mgpweek += lbdays
				}
			}
			//fmt.Printf("  --> %s::GapsPerDay: %d GapsPerWeek: %d\n",
			//	cl.Tag, mgpday, mgpweek)
		}
		if mgpday >= 0 {
			cmaxgpd = append(cmaxgpd, maxGapsPerDay{
				Weight_Percentage: 100,
				Students:          cl.Tag,
				Max_Gaps:          mgpday,
				Active:            true,
			})
		}

		if mgpweek >= 0 {
			cmaxgpw = append(cmaxgpw, maxGapsPerWeek{
				Weight_Percentage: 100,
				Students:          cl.Tag,
				Max_Gaps:          mgpweek,
				Active:            true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetNotAvailableTimes = natimes
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMinHoursDaily = cminlpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxHoursDaily = cmaxlpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxGapsPerDay = cmaxgpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxGapsPerWeek = cmaxgpw
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetIntervalMaxDaysPerWeek = cmaxaft
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetEarlyMaxBeginningsAtSecondHour = cmaxls
	// lunch breaks
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxHoursDailyInInterval = clblist
}
