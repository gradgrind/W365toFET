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
