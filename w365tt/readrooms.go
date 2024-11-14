package w365tt

import (
	"W365toFET/base"
	"W365toFET/logging"
	"fmt"
	"strconv"
	"strings"
)

func (db *DbTopLevel) readRooms(newdb *base.DbTopLevel) {
	db.RealRooms = map[base.Ref]*base.Room{}
	db.RoomTags = map[string]base.Ref{}
	db.RoomChoiceNames = map[string]base.Ref{}
	for _, e := range db.Rooms {
		// Perform some checks and add to the RoomTags map.
		_, nok := db.RoomTags[e.Tag]
		if nok {
			logging.Error.Fatalf(
				"Room Tag (Shortcut) defined twice: %s\n",
				e.Tag)
		}
		db.RoomTags[e.Tag] = e.Id
		// Copy to base db.
		tsl := []base.TimeSlot{}
		for _, ts := range e.NotAvailable {
			tsl = append(tsl, base.TimeSlot(ts))
		}
		r := &base.Room{
			Id:           e.Id,
			Tag:          e.Tag,
			Name:         e.Name,
			NotAvailable: tsl,
		}
		db.RealRooms[e.Id] = r
		newdb.Rooms = append(newdb.Rooms, r)
	}
}

// In the case of RoomGroups, cater for empty Tags (Shortcuts).
func (db *DbTopLevel) readRoomGroups(newdb *base.DbTopLevel) {
	for _, e := range db.RoomGroups {
		if e.Tag != "" {
			_, nok := db.RoomTags[e.Tag]
			if nok {
				logging.Error.Fatalf(
					"Room Tag (Shortcut) defined twice: %s\n",
					e.Tag)
			}
			db.RoomTags[e.Tag] = e.Id
		}
		// Copy to base db.
		newdb.RoomGroups = append(newdb.RoomGroups,
			&base.RoomGroup{
				Id:    e.Id,
				Tag:   e.Tag,
				Name:  e.Name,
				Rooms: e.Rooms,
			})
	}
}

// Call this after all room types have been "read".
// TODO: But then, on the base db ...
func (dbp *DbTopLevel) checkRoomGroups() {
	for i := 0; i < len(dbp.RoomGroups); i++ {
		n := dbp.RoomGroups[i]
		// Collect the Ids and Tags of the component rooms.
		taglist := []string{}
		reflist := []Ref{}
		for _, rref := range n.Rooms {
			r, ok := dbp.Elements[rref]
			if ok {
				rm, ok := r.(*Room)
				if ok {
					reflist = append(reflist, rref)
					taglist = append(taglist, rm.Tag)
					continue
				}
			}
			logging.Error.Printf(
				"Invalid Room in RoomGroup %s:\n  %s\n",
				n.Tag, rref)
		}
		if n.Tag == "" {
			// Make a new Tag
			var tag string
			i := 0
			for {
				i++
				tag = "{" + strconv.Itoa(i) + "}"
				_, nok := dbp.RoomTags[tag]
				if !nok {
					break
				}
			}
			n.Tag = tag
			dbp.RoomTags[tag] = n.Id
			// Also extend the name
			if n.Name == "" {
				n.Name = strings.Join(taglist, ",")
			} else {
				n.Name = strings.Join(taglist, ",") + ":: " + n.Name
			}
		} else if n.Name == "" {
			n.Name = strings.Join(taglist, ",")
		}
		n.Rooms = reflist
	}
}

func (db *DbTopLevel) makeRoomChoiceGroup(rooms []Ref) (Ref, string) {
	erlist := []string{} // Error messages
	// Collect the Ids and Tags of the component rooms.
	taglist := []string{}
	reflist := []Ref{}
	for _, rref := range rooms {
		r, ok := db.RealRooms[rref]
		if ok {
			reflist = append(reflist, rref)
			taglist = append(taglist, r.Tag)
			continue
		}
		erlist = append(erlist,
			fmt.Sprintf(
				"  ++ Invalid Room in new RoomChoiceGroup:\n  %s\n", rref))
	}
	name := strings.Join(taglist, ",")
	// Reuse existing Element when the rooms match.
	id, ok := db.RoomChoiceNames[name]
	if !ok {
		// Make a new Tag
		var tag string
		i := 0
		for {
			i++
			tag = "[" + strconv.Itoa(i) + "]"
			_, nok := db.RoomTags[tag]
			if !nok {
				break
			}
		}
		// Add new Element
		id = db.NewId()
		r := &base.RoomChoiceGroup{
			Id:    id,
			Tag:   tag,
			Name:  name,
			Rooms: reflist,
		}
		db.RoomTags[tag] = id
		db.RoomChoiceNames[name] = id
		//TODO
		newdb.RoomChoiceGroups = append(newdb.RoomChoiceGroups, r)
	}
	return id, strings.Join(erlist, "")
}
