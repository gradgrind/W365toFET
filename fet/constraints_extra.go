package fet

import (
	"W365toFET/base"
	"encoding/xml"
)

type preferredSlots struct {
	XMLName                        xml.Name `xml:"ConstraintActivitiesPreferredTimeSlots"`
	Weight_Percentage              string
	Teacher                        string
	Students                       string
	Subject                        string
	Activity_Tag                   string
	Duration                       string
	Number_of_Preferred_Time_Slots int
	Preferred_Time_Slot            []preferredTime
	Active                         bool
}

type preferredTime struct {
	//XMLName                       xml.Name `xml:"Preferred_Time_Slot"`
	Preferred_Day  string
	Preferred_Hour string
}

type lessonEndsDay struct {
	XMLName           xml.Name `xml:"ConstraintActivityEndsStudentsDay"`
	Weight_Percentage string
	Activity_Id       int
	Active            bool
}

type activityPreferredTimes struct {
	XMLName                        xml.Name `xml:"ConstraintActivityPreferredTimeSlots"`
	Weight_Percentage              string
	Activity_Id                    int
	Number_of_Preferred_Time_Slots int
	Preferred_Time_Slot            []preferredTime
	Active                         bool
}

func getExtraConstraints(fetinfo *fetInfo) {
	tclist := &fetinfo.fetdata.Time_Constraints_List
	db := fetinfo.db
	for _, c := range db.Constraints {
		{
			cn, ok := c.(*base.AutomaticDifferentDays)
			if ok {
				if fetinfo.autoDifferentDays == nil {
					fetinfo.autoDifferentDays = cn
				} else {
					base.Bug.Fatalln(
						"More than one AutomaticDifferentDays constraint")
				}
				continue
			}
		}
		{
			cn, ok := c.(*base.DaysBetween)
			if ok {
				for _, cref := range cn.Courses {
					fetinfo.daysBetween[cref] = append(
						fetinfo.daysBetween[cref], cn)
				}
				continue
			}
		}
		{
			cn, ok := c.(*base.DaysBetweenJoin)
			if ok {
				c1 := fetinfo.courseInfo[cn.Course1]
				c2 := fetinfo.courseInfo[cn.Course2]
				mdba := []minDaysBetweenActivities{}
				for _, l1 := range c1.activities {
					for _, l2 := range c2.activities {
						mdba = append(mdba, minDaysBetweenActivities{
							Weight_Percentage:       weight2fet(cn.Weight),
							Consecutive_If_Same_Day: cn.ConsecutiveIfSameDay,
							Number_of_Activities:    2,
							Activity_Id:             []int{l1, l2},
							MinDays:                 cn.DaysBetween,
							Active:                  true,
						})
					}
				}
				// Append constraints to full list
				tclist.ConstraintMinDaysBetweenActivities = append(
					tclist.ConstraintMinDaysBetweenActivities,
					mdba...)
				continue
			}
		}
		{
			cn, ok := c.(*base.LessonsEndDay)
			if ok {
				cinfo, ok := fetinfo.courseInfo[cn.Course]
				if !ok {
					base.Bug.Fatalf("Invalid course: %s\n", cn.Course)
				}
				for _, aid := range cinfo.activities {
					tclist.ConstraintActivityEndsStudentsDay = append(
						tclist.ConstraintActivityEndsStudentsDay,
						lessonEndsDay{
							Weight_Percentage: weight2fet(cn.Weight),
							Activity_Id:       aid,
							Active:            true,
						})
				}
				continue
			}
		}
		{
			cn, ok := c.(*base.BeforeAfterHour)
			if ok {
				timeslots := []preferredTime{}
				if cn.After {
					for _, dd := range fetinfo.days {
						for h := cn.Hour + 1; h < len(fetinfo.hours); h++ {
							timeslots = append(timeslots, preferredTime{
								Preferred_Day:  dd,
								Preferred_Hour: fetinfo.hours[h],
							})
						}
					}
				} else {
					for _, dd := range fetinfo.days {
						for h := 0; h < cn.Hour; h++ {
							timeslots = append(timeslots, preferredTime{
								Preferred_Day:  dd,
								Preferred_Hour: fetinfo.hours[h],
							})
						}
					}
				}
				for _, k := range cn.Courses {
					cinfo, ok := fetinfo.courseInfo[k]
					if !ok {
						base.Bug.Fatalf("Invalid course: %s\n", k)
					}
					for _, aid := range cinfo.activities {
						tclist.ConstraintActivityPreferredTimeSlots = append(
							tclist.ConstraintActivityPreferredTimeSlots,
							activityPreferredTimes{
								Weight_Percentage:              weight2fet(cn.Weight),
								Activity_Id:                    aid,
								Number_of_Preferred_Time_Slots: len(timeslots),
								Preferred_Time_Slot:            timeslots,
								Active:                         true,
							})
					}
				}
				continue
			}
		}
	}

}
