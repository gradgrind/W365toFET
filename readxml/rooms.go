package readxml

import (
	"W365toFET/base"
	"fmt"
	"slices"
	"strconv"
	"strings"
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
		cdata.roomTags[n.Shortcut] = true
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
	for _, ref := range splitRefList(reflist) {
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

func (cdata *conversionData) getCourseRoom(c *Course) Ref {
	//
	// Deal with the PreferredRooms field of a Course – in W365 the entries
	// can be either a single RoomGroup or a list of Rooms. For the base db
	// there may only be one room, which may be a Room, RoomGroup or
	// RoomChoiceGroup. The latter must be built from a list of "real"
	// Rooms.
	//
	rlist := []Ref{}
	refs := splitRefList(c.PreferredRooms)
	for _, ref := range refs {
		s, ok := cdata.db.Elements[ref]
		if ok {
			_, ok := s.(*base.Room)
			if ok {
				rlist = append(rlist, ref)
				continue
			}
			_, ok = s.(*base.RoomGroup)
			if ok {
				if len(refs) != 1 {
					base.Error.Fatalf("In Course %s:\n"+
						"  -- a RoomGroup must be the only item in the"+
						" PreferredRooms list: %s\n",
						c.Id, ref)
				}
				return ref
			}
		}
		base.Error.Fatalf("In Course %s:\n  -- Invalid Room: %s\n",
			c.Id, ref)
	}
	if len(rlist) == 0 {
		return ""
	}
	if len(rlist) == 1 {
		return rlist[0]
	}
	// Build a RoomChoiceGroup – or fetch a cached one.
	return cdata.makeRoomChoiceGroup(rlist)
}

func (cdata *conversionData) makeRoomChoiceGroup(rooms []Ref) Ref {
	// Collect the Tags (Shortcuts) of the component rooms.
	taglist := []string{}
	db := cdata.db
	for _, rref := range rooms {
		r, ok := db.Elements[rref]
		if ok {
			rr, ok := r.(*base.Room)
			if ok {
				taglist = append(taglist, rr.Tag)
				continue
			}
		}
		base.Bug.Fatalf("%s is not a (real) Room\n", rref)
	}
	name := strings.Join(taglist, ",")
	// Reuse existing Element when the rooms match.
	ref, ok := cdata.roomChoiceNames[name]
	if !ok {
		// Make a new Tag
		var tag string
		i := 0
		for {
			i++
			tag = "[" + strconv.Itoa(i) + "]"
			_, nok := cdata.roomTags[tag]
			if !nok {
				break
			}
		}
		// Add new Element
		r := db.NewRoomChoiceGroup("")
		r.Tag = tag
		r.Name = name
		r.Rooms = rooms
		ref = r.Id
		cdata.roomTags[tag] = true
		cdata.roomChoiceNames[name] = ref
	}
	return ref
}
