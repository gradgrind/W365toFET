package fet

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"encoding/xml"
	"io"
	"os"
)

type fetPlacement struct {
	// Note that this is intended for Days and Hours which are 0-based indexes.
	// It will not work with strings ("normal" FET references)
	Id        int
	Day       int
	Hour      int
	Room      string
	Real_Room []string `xml:",omitempty"`
}

type fetActivities struct {
	XMLName    xml.Name       `xml:"Activities_Timetable"`
	Placements []fetPlacement `xml:"Activity"`
}

type ActivityPlacement struct {
	Id    int
	Day   int
	Hour  int
	Rooms []Ref
}

func ReadPlacements(
	ttinfo *ttbase.TtInfo,
	xmlpath string,
) []ActivityPlacement {
	// Open the  XML activities file
	xmlFile, err := os.Open(xmlpath)
	if err != nil {
		base.Error.Fatal(err)
	}
	// Remember to close the file at the end of the function
	defer xmlFile.Close()
	// read the opened XML file as a byte array.
	base.Message.Printf("Reading: %s\n", xmlpath)
	byteValue, _ := io.ReadAll(xmlFile)
	v := fetActivities{}
	err = xml.Unmarshal(byteValue, &v)
	if err != nil {
		base.Error.Fatalf("XML error in %s:\n %v\n", xmlpath, err)
	}

	// Need mapping for the Rooms
	rmap := map[string]Ref{}
	for _, r := range ttinfo.Db.Rooms {
		rmap[r.Tag] = r.Id
	}

	placements := []ActivityPlacement{}
	for _, p := range v.Placements {
		rlist := []Ref{}
		if len(p.Real_Room) == 0 {
			if p.Room != "" {
				rlist = append(rlist, rmap[p.Room])
			}
		} else {
			for _, r := range p.Real_Room {
				rlist = append(rlist, rmap[r])
			}
		}
		placements = append(placements, ActivityPlacement{
			Id:    p.Id,
			Day:   p.Day,
			Hour:  p.Hour,
			Rooms: rlist,
		})
	}
	return placements
}
