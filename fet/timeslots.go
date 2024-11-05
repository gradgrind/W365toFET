package fet

import (
	"encoding/xml"
	"fmt"
)

type fetDay struct {
	XMLName   xml.Name `xml:"Day"`
	Name      string
	Long_Name string
}

type fetDaysList struct {
	XMLName        xml.Name `xml:"Days_List"`
	Number_of_Days int
	Day            []fetDay
}

type fetHour struct {
	XMLName   xml.Name `xml:"Hour"`
	Name      string
	Long_Name string
}

type fetHoursList struct {
	XMLName         xml.Name `xml:"Hours_List"`
	Number_of_Hours int
	Hour            []fetHour
}

func getDays(fetinfo *fetInfo) {
	days := []fetDay{}
	dlist := []string{}
	for _, n := range fetinfo.db.Days {
		days = append(days, fetDay{Name: n.Tag, Long_Name: n.Name})
		dlist = append(dlist, n.Tag)
	}
	fetinfo.days = dlist
	fetinfo.fetdata.Days_List = fetDaysList{
		Number_of_Days: len(days),
		Day:            days,
	}
}

func getHours(fetinfo *fetInfo) {
	hours := []fetHour{}
	hlist := []string{}
	for _, n := range fetinfo.db.Hours {
		hours = append(hours, fetHour{
			Name:      n.Tag,
			Long_Name: fmt.Sprintf("%s@%s-%s", n.Name, n.Start, n.End),
		})
		hlist = append(hlist, n.Tag)
	}
	fetinfo.hours = hlist
	fetinfo.fetdata.Hours_List = fetHoursList{
		Number_of_Hours: len(hours),
		Hour:            hours,
	}
}
