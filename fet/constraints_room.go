package fet

import "strconv"

func (fetinfo *fetInfo) handle_room_constraints() {
	tt_data := fetinfo.tt_data
	db := tt_data.Db

	natimes := []roomNotAvailable{}
	for rix, matrix := range tt_data.RoomNotAvailable {
		nats := []notAvailableTime{}
		for d, hlist := range matrix {
			for h, blocked := range hlist {
				if blocked {
					nats = append(nats,
						notAvailableTime{
							Day:  strconv.Itoa(d),
							Hour: strconv.Itoa(h)})
				}
			}
		}
		if len(nats) > 0 {
			r := db.Rooms[rix]
			natimes = append(natimes,
				roomNotAvailable{
					Weight_Percentage:             100,
					Room:                          r.Tag,
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
				})
		}
	}
	fetinfo.fetdata.Space_Constraints_List.
		ConstraintRoomNotAvailableTimes = natimes
}
