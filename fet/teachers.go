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
