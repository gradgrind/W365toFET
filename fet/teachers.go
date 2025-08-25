package fet

import (
	"encoding/xml"
	"fmt"
	"slices"
	"strconv"
)

type fetTeacher struct {
	XMLName   xml.Name `xml:"Teacher"`
	Name      string
	Long_Name string
	Comments  string
}

type fetTeachersList struct {
	XMLName xml.Name `xml:"Teachers_List"`
	Teacher []fetTeacher
}

func getTeachers(fetinfo *fetInfo) {
	items := []fetTeacher{}
	for _, n := range fetinfo.tt_data.Db.Teachers {
		items = append(items, fetTeacher{
			Name: n.Tag,
			Long_Name: fmt.Sprintf("%s %s",
				n.Firstname,
				n.Name,
			),
			//<Target_Number_of_Hours>0</Target_Number_of_Hours>
			//<Qualified_Subjects></Qualified_Subjects>
		})
	}
	fetinfo.fetdata.Teachers_List = fetTeachersList{
		Teacher: items,
	}
}

/* ***** Constraints ***** */

/* Lunch-breaks

Lunch-breaks can be done using max-hours-in-interval constraint, but that
makes specification of max-gaps more difficult (becuase the lunch breaks
count as gaps).

The alternative is to add dummy lessons, clamped to the midday-break hours,
on the days where none of the midday-break hours are blocked. However, this
can also cause problems with gaps – the dummy lesson can itself create gaps,
for example when a teacher's lessons are earlier in the day.

All in all, I think the max-hours-in-interval constraint is probably better
for the teachers. If there is a maximum-gaps constraint, the user may need
to adjust it to take the lunch-breaks into acccount.
*/

func addTeacherConstraints(fetinfo *fetInfo) {
	natimes := []teacherNotAvailable{}
	tmaxdpw := []maxDaysT{}
	tminlpd := []minLessonsPerDayT{}
	tmaxlpd := []maxLessonsPerDayT{}
	tmaxgpd := []maxGapsPerDayT{}
	tmaxgpw := []maxGapsPerWeekT{}
	tmaxaft := []maxDaysinIntervalPerWeekT{}
	tlblist := []lunchBreakT{}
	tt_data := fetinfo.tt_data
	ndays := tt_data.NDays
	nhours := tt_data.NHours
	db := tt_data.Db

	for _, t := range db.Teachers {
		// "Not available" times
		nats := []notAvailableTime{}
		for _, dh := range t.NotAvailable {
			nats = append(nats,
				notAvailableTime{
					Day:  strconv.Itoa(dh.Day),
					Hour: strconv.Itoa(dh.Hour)})
		}
		if len(nats) > 0 {
			natimes = append(natimes,
				teacherNotAvailable{
					Weight_Percentage:             100,
					Teacher:                       t.Tag,
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
				})
		}

		n := t.MaxDays
		if n >= 0 && n < ndays {
			tmaxdpw = append(tmaxdpw, maxDaysT{
				Weight_Percentage: 100,
				Teacher:           t.Tag,
				Max_Days_Per_Week: n,
				Active:            true,
			})
		}

		n = t.MinLessonsPerDay
		if n >= 2 && n <= nhours {
			tminlpd = append(tminlpd, minLessonsPerDayT{
				Weight_Percentage:   100,
				Teacher:             t.Tag,
				Minimum_Hours_Daily: n,
				Allow_Empty_Days:    true,
				Active:              true,
			})
		}

		n = t.MaxLessonsPerDay
		if n >= 0 && n < nhours {
			tmaxlpd = append(tmaxlpd, maxLessonsPerDayT{
				Weight_Percentage:   100,
				Teacher:             t.Tag,
				Maximum_Hours_Daily: n,
				Active:              true,
			})
		}

		i := db.Info.FirstAfternoonHour
		maxpm := t.MaxAfternoons
		if maxpm >= 0 && i > 0 {
			tmaxaft = append(tmaxaft, maxDaysinIntervalPerWeekT{
				Weight_Percentage:   100,
				Teacher:             t.Tag,
				Interval_Start_Hour: strconv.Itoa(i),
				Interval_End_Hour:   "", // end of day
				Max_Days_Per_Week:   maxpm,
				Active:              true,
			})
		}

		// The lunch-break constraint may require adjustment of these:
		mgpday := t.MaxGapsPerDay
		mgpweek := t.MaxGapsPerWeek

		if mbhours := db.Info.MiddayBreak; len(mbhours) != 0 && t.LunchBreak {
			// Generate the constraint unless all days have a blocked lesson
			// at lunchtime.
			lbdays := ndays
			d := 0
			for _, ts := range t.NotAvailable {
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
				tlblist = append(tlblist, lunchBreakT{
					Weight_Percentage:   100,
					Teacher:             t.Tag,
					Interval_Start_Hour: strconv.Itoa(mbhours[0]),
					Interval_End_Hour:   strconv.Itoa(mbhours[0] + len(mbhours)),
					Maximum_Hours_Daily: len(mbhours) - 1,
					Active:              true,
				})
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
		}

		if mgpday >= 0 {
			tmaxgpd = append(tmaxgpd, maxGapsPerDayT{
				Weight_Percentage: 100,
				Teacher:           t.Tag,
				Max_Gaps:          mgpday,
				Active:            true,
			})
		}

		if mgpweek >= 0 {
			tmaxgpw = append(tmaxgpw, maxGapsPerWeekT{
				Weight_Percentage: 100,
				Teacher:           t.Tag,
				Max_Gaps:          mgpweek,
				Active:            true,
			})
		}

	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherNotAvailableTimes = natimes
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxDaysPerWeek = tmaxdpw
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMinHoursDaily = tminlpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxHoursDaily = tmaxlpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxGapsPerDay = tmaxgpd
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxGapsPerWeek = tmaxgpw
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherIntervalMaxDaysPerWeek = tmaxaft
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxHoursDailyInInterval = tlblist
}
