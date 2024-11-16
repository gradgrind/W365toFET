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

func getExtraConstraints(fetinfo *fetInfo) {
	tclist := &fetinfo.fetdata.Time_Constraints_List
	db := fetinfo.db
	for _, c := range db.Constraints {
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

}
