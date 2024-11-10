package w365tt

import (
	"W365toFET/logging"
	"fmt"
	"strconv"
	"strings"
)

func (dbp *DbTopLevel) readRooms() {
	for i := 0; i < len(dbp.Rooms); i++ {
		n := &dbp.Rooms[i]
		_, nok := dbp.RoomTags[n.Tag]
		if nok {
			logging.Error.Fatalf(
				"Room Tag (Shortcut) defined twice: %s\n",
				n.Tag)
		}
		dbp.RoomTags[n.Tag] = n.Id

		if len(n.NotAvailable) == 0 {
			// Avoid a null value
			n.NotAvailable = []TimeSlot{}
		}
	}
}

// In the case of RoomGroups, cater for empty Tags/Shortcuts
func (dbp *DbTopLevel) readRoomGroups() {
	for i := 0; i < len(dbp.RoomGroups); i++ {
		n := &dbp.RoomGroups[i]
		if n.Tag != "" {
			_, nok := dbp.RoomTags[n.Tag]
			if nok {
				logging.Error.Fatalf(
					"Room Tag (Shortcut) defined twice: %s\n",
					n.Tag)
			}
			dbp.RoomTags[n.Tag] = n.Id
		}
	}
}

// Call this after all room types have been "read".
func (dbp *DbTopLevel) checkRoomGroups() {
	for i := 0; i < len(dbp.RoomGroups); i++ {
		n := &dbp.RoomGroups[i]
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

// Here the Tags are not checked (there should always be one ...).
func (dbp *DbTopLevel) readRoomChoiceGroups() {
	for i := 0; i < len(dbp.RoomChoiceGroups); i++ {
		n := &dbp.RoomChoiceGroups[i]
		_, nok := dbp.RoomTags[n.Tag]
		if nok {
			logging.Error.Fatalf(
				"Room Tag (Shortcut) defined twice: %s\n",
				n.Tag)
		}
		// Check component rooms.
		reflist := []Ref{}
		for _, rref := range n.Rooms {
			r, ok := dbp.Elements[rref]
			if ok {
				_, ok = r.(*Room)
				if ok {
					reflist = append(reflist, rref)
					continue
				}
			}
			logging.Error.Printf(
				"Invalid Room in RoomChoiceGroup %s:\n  %s\n",
				n.Tag, rref)
		}
		n.Rooms = reflist
		dbp.RoomTags[n.Tag] = n.Id
	}
}

func (dbp *DbTopLevel) makeRoomChoiceGroup(rooms []Ref) (Ref, string) {
	erlist := []string{} // Error messages
	// Collect the Ids and Tags of the component rooms.
	taglist := []string{}
	reflist := []Ref{}
	for _, rref := range rooms {
		r, ok := dbp.Elements[rref]
		if ok {
			rm, ok := r.(*Room)
			if ok {
				reflist = append(reflist, rref)
				taglist = append(taglist, rm.Tag)
				continue
			}
		}
		erlist = append(erlist,
			fmt.Sprintf(
				"  ++ Invalid Room in new RoomChoiceGroup:\n  %s\n", rref))
	}
	name := strings.Join(taglist, ",")
	// Reuse existing Element when the rooms match.
	id, ok := dbp.RoomChoiceNames[name]
	if !ok {
		// Make a new Tag
		var tag string
		i := 0
		for {
			i++
			tag = "[" + strconv.Itoa(i) + "]"
			_, nok := dbp.RoomTags[tag]
			if !nok {
				break
			}
		}
		// Add new Element
		id = dbp.NewId()
		rcglen := len(dbp.RoomChoiceGroups)
		dbp.RoomChoiceGroups = append(dbp.RoomChoiceGroups,
			RoomChoiceGroup{
				Id:    id,
				Tag:   tag,
				Name:  name,
				Rooms: reflist,
			})
		dbp.AddElement(id, &dbp.RoomChoiceGroups[rcglen])
		dbp.RoomTags[tag] = id
		dbp.RoomChoiceNames[name] = id
	}
	return id, strings.Join(erlist, "")
}
