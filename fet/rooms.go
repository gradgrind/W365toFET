package fet

import (
	"W365toFET/timetable"
	"encoding/xml"
	"fmt"
	"strings"
)

type fetRoom struct {
	XMLName                      xml.Name `xml:"Room"`
	Name                         string   // e.g. k3 ...
	Long_Name                    string
	Capacity                     int           // 30000
	Virtual                      bool          // false
	Number_of_Sets_of_Real_Rooms int           `xml:",omitempty"`
	Set_of_Real_Rooms            []realRoomSet `xml:",omitempty"`
	Comments                     string
}

type realRoomSet struct {
	Number_of_Real_Rooms int // normally 1, I suppose
	Real_Room            []string
}

type fetRoomsList struct {
	XMLName xml.Name `xml:"Rooms_List"`
	Room    []fetRoom
}

type placedRoom struct {
	XMLName              xml.Name `xml:"ConstraintActivityPreferredRoom"`
	Weight_Percentage    int
	Activity_Id          timetable.ActivityIndex
	Room                 string
	Number_of_Real_Rooms int      `xml:",omitempty"`
	Real_Room            []string `xml:",omitempty"`
	Permanently_Locked   bool     // false
	Active               bool     // true
}

type roomChoice struct {
	XMLName                   xml.Name `xml:"ConstraintActivityPreferredRooms"`
	Weight_Percentage         int
	Activity_Id               timetable.ActivityIndex
	Number_of_Preferred_Rooms int
	Preferred_Room            []string
	Active                    bool // true
}

type roomNotAvailable struct {
	XMLName                       xml.Name `xml:"ConstraintRoomNotAvailableTimes"`
	Weight_Percentage             int
	Room                          string
	Number_of_Not_Available_Times int
	Not_Available_Time            []notAvailableTime
	Active                        bool
}

// Generate the fet entries for the basic ("real") rooms.
func getRooms(fetinfo *fetInfo) {
	rooms := []fetRoom{}
	for _, n := range fetinfo.tt_data.SharedData.Db.Rooms {
		rooms = append(rooms, fetRoom{
			Name:      n.Tag,
			Long_Name: n.Name,
			Capacity:  30000,
			Virtual:   false,
			Comments:  string(n.Id),
		})
	}
	fetinfo.fetdata.Rooms_List = fetRoomsList{
		Room: rooms,
	}
}

func (fetinfo *fetInfo) getFetRooms(cinfo *timetable.CourseInfo) []string {
	// The fet virtual rooms are cached at fetinfo.fetVirtualRooms.
	var result []string
	tt_data := fetinfo.tt_data

	// First get the Element Tags for FET.
	rtags := []string{}
	for _, rr := range cinfo.FixedRooms {
		rtags = append(rtags,
			tt_data.SharedData.RoomNodes[rr].GetResourceTag())
	}
	rctags := [][]string{}
	for _, rc := range cinfo.RoomChoices {
		rcl := []string{}
		for _, rr := range rc {
			rcl = append(rcl,
				tt_data.SharedData.RoomNodes[rr].GetResourceTag())
		}
		rctags = append(rctags, rcl)
	}

	if len(rctags) == 0 && len(rtags) < 2 {
		result = rtags
	} else if len(rctags) == 1 && len(rtags) == 0 {
		result = rctags[0]
	} else {
		// Otherwise a virtual room is necessary.
		srctags := []string{}
		for _, rcl := range rctags {
			srctags = append(srctags, strings.Join(rcl, ","))
		}
		key := strings.Join(rtags, ",") + "+" + strings.Join(srctags, "|")
		vr, ok := fetinfo.fetVirtualRooms[key]
		if !ok {
			// Make virtual room, using rooms list from above.
			rrslist := []realRoomSet{}
			for _, rt := range rtags {
				rrslist = append(rrslist, realRoomSet{
					Number_of_Real_Rooms: 1,
					Real_Room:            []string{rt},
				})
			}
			// Add choice lists from above.
			for _, rtl := range rctags {
				rrslist = append(rrslist, realRoomSet{
					Number_of_Real_Rooms: len(rtl),
					Real_Room:            rtl,
				})
			}
			vr = fmt.Sprintf(
				"%s%03d", VIRTUAL_ROOM_PREFIX, len(fetinfo.fetVirtualRooms)+1)
			vroom := fetRoom{
				Name:                         vr,
				Capacity:                     30000,
				Virtual:                      true,
				Number_of_Sets_of_Real_Rooms: len(rrslist),
				Set_of_Real_Rooms:            rrslist,
			}
			// Add the virtual room to the fet file
			fetinfo.fetdata.Rooms_List.Room = append(
				fetinfo.fetdata.Rooms_List.Room, vroom)
			// Remember key/value
			fetinfo.fetVirtualRooms[key] = vr
			fetinfo.fetVirtualRoomN[vr] = len(rrslist)
		}
		result = []string{vr}
	}
	//--fmt.Printf("   --> %+v\n", result)
	return result
}
