package fet

import (
	"W365toFET/base"
	"encoding/xml"
)

type preferredSlots struct {
	XMLName                        xml.Name `xml:"ConstraintActivitiesPreferredTimeSlots"`
	Weight_Percentage              int
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
