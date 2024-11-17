package readxml

import (
	"W365toFET/base"
	"fmt"
	"slices"
)

func (cdata *conversionData) readRooms() {
	slices.SortFunc(cdata.xmlin.Rooms, func(a, b Room) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	rglist := map[Ref]*Room{} // RoomGroup elements
	for i := 0; i < len(cdata.xmlin.Rooms); i++ {
		n := &cdata.xmlin.Rooms[i]
		// Extract RoomGroup elements
		if n.RoomGroups != "" {
			rglist[n.Id] = n
			continue
		}
		// Normal Room
		e := cdata.db.NewRoom(n.Id)
		e.Name = n.Name
		e.Tag = n.Shortcut
		e.NotAvailable = cdata.getAbsences(n.Absences,
			fmt.Sprintf("In Room %s (Absences)", n.Id))
	}

	// Now handle the RoomGroups
	for nid, n := range rglist {
		e := base.NewDb().NewRoomGroup(nid)
		e.Name = n.Name
		e.Tag = n.Shortcut
		e.Rooms = cdata.checkRealRooms(n.RoomGroups,
			fmt.Sprintf("In Room %s (RoomGroups)", nid))
	}
}

func (cdata *conversionData) checkRealRooms(
	reflist RefList,
	msg string,
) []Ref {
	result := []Ref{}
	for _, ref := range SplitRefList(reflist) {
		e, ok := cdata.db.Elements[ref]
		if ok {
			_, ok = e.(*base.Room)
			if ok {
				result = append(result, ref)
				continue
			}
		}
		base.Error.Fatalf("%s:\n  -- Invalid Room: %s\n", msg, ref)
	}
	return result
}
