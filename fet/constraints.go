package fet

import (
	"encoding/xml"
)

type startingTime struct {
	XMLName            xml.Name `xml:"ConstraintActivityPreferredStartingTime"`
	Weight_Percentage  int
	Activity_Id        int
	Preferred_Day      string
	Preferred_Hour     string
	Permanently_Locked bool
	Active             bool
}

type minDaysBetweenActivities struct {
	XMLName                 xml.Name `xml:"ConstraintMinDaysBetweenActivities"`
	Weight_Percentage       int
	Consecutive_If_Same_Day bool
	Number_of_Activities    int
	Activity_Id             []int
	MinDays                 int
	Active                  bool
}

// *** Teacher constraints
type lunchBreakT struct {
	XMLName             xml.Name `xml:"ConstraintTeacherMaxHoursDailyInInterval"`
	Weight_Percentage   int
	Teacher             string
	Interval_Start_Hour string
	Interval_End_Hour   string
	Maximum_Hours_Daily int
	Active              bool
}

type maxGapsPerDayT struct {
	XMLName           xml.Name `xml:"ConstraintTeacherMaxGapsPerDay"`
	Weight_Percentage int
	Teacher           string
	Max_Gaps          int
	Active            bool
}

type maxGapsPerWeekT struct {
	XMLName           xml.Name `xml:"ConstraintTeacherMaxGapsPerWeek"`
	Weight_Percentage int
	Teacher           string
	Max_Gaps          int
	Active            bool
}

type minLessonsPerDayT struct {
	XMLName             xml.Name `xml:"ConstraintTeacherMinHoursDaily"`
	Weight_Percentage   int
	Teacher             string
	Minimum_Hours_Daily int
	Allow_Empty_Days    bool
	Active              bool
}

type maxLessonsPerDayT struct {
	XMLName             xml.Name `xml:"ConstraintTeacherMaxHoursDaily"`
	Weight_Percentage   int
	Teacher             string
	Maximum_Hours_Daily int
	Active              bool
}

type maxDaysT struct {
	XMLName           xml.Name `xml:"ConstraintTeacherMaxDaysPerWeek"`
	Weight_Percentage int
	Teacher           string
	Max_Days_Per_Week int
	Active            bool
}

// for MaxAfternoons
type maxDaysinIntervalPerWeekT struct {
	XMLName             xml.Name `xml:"ConstraintTeacherIntervalMaxDaysPerWeek"`
	Weight_Percentage   int
	Teacher             string
	Interval_Start_Hour string
	Interval_End_Hour   string
	// Interval_End_Hour void ("") means the end of the day (which has no name)
	Max_Days_Per_Week int
	Active            bool
}

// *** Class constraints

type lunchBreak struct {
	XMLName             xml.Name `xml:"ConstraintStudentsSetMaxHoursDailyInInterval"`
	Weight_Percentage   int
	Students            string
	Interval_Start_Hour string
	Interval_End_Hour   string
	Maximum_Hours_Daily int
	Active              bool
}

type maxGapsPerDay struct {
	XMLName           xml.Name `xml:"ConstraintStudentsSetMaxGapsPerDay"`
	Weight_Percentage int
	Max_Gaps          int
	Students          string
	Active            bool
}

type maxGapsPerWeek struct {
	XMLName           xml.Name `xml:"ConstraintStudentsSetMaxGapsPerWeek"`
	Weight_Percentage int
	Max_Gaps          int
	Students          string
	Active            bool
}

type minLessonsPerDay struct {
	XMLName             xml.Name `xml:"ConstraintStudentsSetMinHoursDaily"`
	Weight_Percentage   int
	Minimum_Hours_Daily int
	Students            string
	Allow_Empty_Days    bool
	Active              bool
}

type maxLessonsPerDay struct {
	XMLName             xml.Name `xml:"ConstraintStudentsSetMaxHoursDaily"`
	Weight_Percentage   int
	Maximum_Hours_Daily int
	Students            string
	Active              bool
}

// for MaxAfternoons
type maxDaysinIntervalPerWeek struct {
	XMLName             xml.Name `xml:"ConstraintStudentsSetIntervalMaxDaysPerWeek"`
	Weight_Percentage   int
	Students            string
	Interval_Start_Hour string
	Interval_End_Hour   string
	// Interval_End_Hour void ("") means the end of the day (which has no name)
	Max_Days_Per_Week int
	Active            bool
}

// for ForceFirstHour
type maxLateStarts struct {
	XMLName                       xml.Name `xml:"ConstraintStudentsSetEarlyMaxBeginningsAtSecondHour"`
	Weight_Percentage             int
	Max_Beginnings_At_Second_Hour int
	Students                      string
	Active                        bool
}

/*
The different-days constraint for lessons belonging to a single course can
be added automatically, but it should be posible to disable it by passing in
an appropriate constraint. Thus, the built-in constraint must be traceable.
TODO: There could be a separate constraint to link different courses – the
alternative being a subject/atomic-group search.
*/
func addDifferentDaysConstraints(fetinfo *fetInfo) {
	mdba := []minDaysBetweenActivities{}
	for cref, cinfo := range fetinfo.courseInfo {
		nact := len(cinfo.activities)
		if nact < 2 || nact > len(fetinfo.days) {
			continue
		}
		// Need the Acivity_Ids for the Lessons, and whether they are fixed.
		// No two fixed activities should be different-dayed.

		fixeds := []int{}
		unfixeds := []int{}
		for i, l := range cinfo.lessons {
			if l.Fixed {
				fixeds = append(fixeds, cinfo.activities[i])
			} else {
				unfixeds = append(unfixeds, cinfo.activities[i])
			}
		}

		if len(fixeds) <= 1 {
			fetinfo.differentDayConstraints[cref] = []int{len(mdba)}
			mdba = append(mdba, minDaysBetweenActivities{
				Weight_Percentage:       100,
				Consecutive_If_Same_Day: true,
				Number_of_Activities:    len(cinfo.activities),
				Activity_Id:             cinfo.activities,
				MinDays:                 1,
				Active:                  true,
			})
			continue
		}

		if len(unfixeds) == 0 {
			continue
		}

		ddc := []int{} // Collect indexes within mdba
		for _, aid := range fixeds {
			aids := []int{aid}
			aids = append(aids, unfixeds...)
			ddc = append(ddc, len(mdba))
			mdba = append(mdba, minDaysBetweenActivities{
				Weight_Percentage:       100,
				Consecutive_If_Same_Day: true,
				Number_of_Activities:    len(aids),
				Activity_Id:             aids,
				MinDays:                 1,
				Active:                  true,
			})
		}
		fetinfo.differentDayConstraints[cref] = ddc
	}
	// Append constraints to full list
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintMinDaysBetweenActivities = append(
		fetinfo.fetdata.Time_Constraints_List.
			ConstraintMinDaysBetweenActivities,
		mdba...)
}
