package fet

import (
	"encoding/xml"
	"fmt"
	"strconv"
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
	Activity_Id          int
	Room                 string
	Number_of_Real_Rooms int      `xml:",omitempty"`
	Real_Room            []string `xml:",omitempty"`
	Permanently_Locked   bool     // false
	Active               bool     // true
}

type roomChoice struct {
	XMLName                   xml.Name `xml:"ConstraintActivityPreferredRooms"`
	Weight_Percentage         int
	Activity_Id               int
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
	natimes := []roomNotAvailable{}
	for _, n := range fetinfo.db.Rooms {
		rooms = append(rooms, fetRoom{
			Name:      n.Tag,
			Long_Name: n.Name,
			Capacity:  30000,
			Virtual:   false,
			Comments:  getString(n.Id),
		})

		// "Not available" times
		nats := []notAvailableTime{}
		for _, dh := range n.NotAvailable {
			nats = append(nats,
				notAvailableTime{
					Day:  strconv.Itoa(dh.Day),
					Hour: strconv.Itoa(dh.Hour)})
		}

		if len(nats) > 0 {
			natimes = append(natimes,
				roomNotAvailable{
					Weight_Percentage:             100,
					Room:                          n.Tag,
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
				})
		}
	}
	fetinfo.fetdata.Rooms_List = fetRoomsList{
		Room: rooms,
	}
	fetinfo.fetdata.Space_Constraints_List.
		ConstraintRoomNotAvailableTimes = natimes
}

func getFetRooms(fetinfo *fetInfo, room virtualRoom) []string {
	// The fet virtual rooms are cached at fetinfo.fetVirtualRooms.
	// First convert the Ref values to Element Tags for FET.
	rtags := []string{}
	for _, rref := range room.rooms {
		rtags = append(rtags, fetinfo.ref2fet[rref])
	}
	rctags := [][]string{}
	for _, rc := range room.roomChoices {
		rcl := []string{}
		for _, rref := range rc {
			rcl = append(rcl, fetinfo.ref2fet[rref])
		}
		rctags = append(rctags, rcl)
	}

	if len(rctags) == 0 && len(rtags) < 2 {
		return rtags
	}
	if len(rctags) == 1 && len(rtags) == 0 {
		return rctags[0]
	}

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
	return []string{vr}
}
