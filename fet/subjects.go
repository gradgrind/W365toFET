package fet

import (
	"encoding/xml"
)

type fetSubject struct {
	XMLName   xml.Name `xml:"Subject"`
	Name      string
	Long_Name string
	Comments  string
}

type fetSubjectsList struct {
	XMLName xml.Name `xml:"Subjects_List"`
	Subject []fetSubject
}

func getSubjects(fetinfo *fetInfo) {
	items := []fetSubject{}
	for _, n := range fetinfo.tt_data.Db.Subjects {
		items = append(items, fetSubject{
			Name:      n.Tag,
			Long_Name: n.Name,
		})
	}
	fetinfo.fetdata.Subjects_List = fetSubjectsList{
		Subject: items,
	}
}
