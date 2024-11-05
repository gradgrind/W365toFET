package fet

import (
	"encoding/xml"
	"fmt"
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

type notAvailableTime struct {
	XMLName xml.Name `xml:"Not_Available_Time"`
	Day     string
	Hour    string
}

type teacherNotAvailable struct {
	XMLName                       xml.Name `xml:"ConstraintTeacherNotAvailableTimes"`
	Weight_Percentage             int
	Teacher                       string
	Number_of_Not_Available_Times int
	Not_Available_Time            []notAvailableTime
	Active                        bool
}

func getTeachers(fetinfo *fetInfo) {
	items := []fetTeacher{}
	natimes := []teacherNotAvailable{}
	for _, n := range fetinfo.db.Teachers {
		items = append(items, fetTeacher{
			Name: n.Tag,
			Long_Name: fmt.Sprintf("%s %s",
				n.Firstname,
				n.Name,
			),
			//<Target_Number_of_Hours>0</Target_Number_of_Hours>
			//<Qualified_Subjects></Qualified_Subjects>
		})

		// "Not available" times
		nats := []notAvailableTime{}
		for _, dh := range n.NotAvailable {
			nats = append(nats,
				notAvailableTime{
					Day:  fetinfo.days[dh.Day],
					Hour: fetinfo.hours[dh.Hour]})
		}

		if len(nats) > 0 {
			natimes = append(natimes,
				teacherNotAvailable{
					Weight_Percentage:             100,
					Teacher:                       n.Tag,
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
				})
		}
	}

	fetinfo.fetdata.Teachers_List = fetTeachersList{
		Teacher: items,
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherNotAvailableTimes = natimes
}
