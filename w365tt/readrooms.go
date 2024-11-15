package w365tt

import (
	"W365toFET/base"
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
			base.Error.Fatalf(
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
	db.RoomGroupMap = map[Ref]*base.RoomGroup{}
	for _, e := range db.RoomGroups {
		if e.Tag != "" {
			_, nok := db.RoomTags[e.Tag]
			if nok {
				base.Error.Fatalf(
					"Room Tag (Shortcut) defined twice: %s\n",
					e.Tag)
			}
			db.RoomTags[e.Tag] = e.Id
		}
		// Copy to base db.
		r := &base.RoomGroup{
			Id:    e.Id,
			Tag:   e.Tag,
			Name:  e.Name,
			Rooms: e.Rooms,
		}
		newdb.RoomGroups = append(newdb.RoomGroups, r)
		db.RoomGroupMap[e.Id] = r
	}
}

// Call this after all room types have been "read".
func (db *DbTopLevel) checkRoomGroups(newdb *base.DbTopLevel) {
	for _, e := range newdb.RoomGroups {
		// Collect the Ids and Tags of the component rooms.
		taglist := []string{}
		reflist := []Ref{}
		for _, rref := range e.Rooms {
			r, ok := db.RealRooms[rref]
			if ok {
				reflist = append(reflist, rref)
				taglist = append(taglist, r.Tag)
				continue

			}
			base.Error.Printf(
				"Invalid Room in RoomGroup %s:\n  %s\n",
				e.Tag, rref)
		}
		if e.Tag == "" {
			// Make a new Tag
			var tag string
			i := 0
			for {
				i++
				tag = "{" + strconv.Itoa(i) + "}"
				_, nok := db.RoomTags[tag]
				if !nok {
					break
				}
			}
			e.Tag = tag
			db.RoomTags[tag] = e.Id
			// Also extend the name
			if e.Name == "" {
				e.Name = strings.Join(taglist, ",")
			} else {
				e.Name = strings.Join(taglist, ",") + ":: " + e.Name
			}
		} else if e.Name == "" {
			e.Name = strings.Join(taglist, ",")
		}
		e.Rooms = reflist
	}
}

func (db *DbTopLevel) makeRoomChoiceGroup(
	newdb *base.DbTopLevel,
	rooms []Ref,
) (Ref, string) {
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
		db.RoomTags[tag] = id
		db.RoomChoiceNames[name] = id
		newdb.RoomChoiceGroups = append(newdb.RoomChoiceGroups,
			&base.RoomChoiceGroup{
				Id:    id,
				Tag:   tag,
				Name:  name,
				Rooms: reflist,
			})
	}
	return id, strings.Join(erlist, "")
}
