package ttbase

import (
	"W365toFET/base"
	"fmt"
	"math/big"
	"slices"
	"strings"
)

type CourseInfo struct {
	Id       Ref
	Subject  Ref
	Groups   []Ref
	Teachers []Ref
	Room     VirtualRoom
	Lessons  []*base.Lesson
	// Resources is a list of all resource-indexes for this course.
	// It is filled separately, later, for working with placements.
	Resources []ResourceIndex
}

// Make a shortish string view of a CourseInfo – can be useful in tests
func (ttinfo *TtInfo) View(cinfo *CourseInfo) string {
	tlist := []string{}
	for _, t := range cinfo.Teachers {
		tlist = append(tlist, ttinfo.Ref2Tag[t])
	}
	glist := []string{}
	for _, g := range cinfo.Groups {
		gx, ok := ttinfo.Ref2Tag[g]
		if !ok {
			base.Bug.Fatalf("No Ref2Tag for %s\n", g)
		}
		glist = append(glist, gx)
	}
	rlist0 := []string{}
	for _, rref := range cinfo.Room.Rooms {
		rlist0 = append(rlist0, ttinfo.Ref2Tag[rref])
	}
	r0 := strings.Join(rlist0, ",")
	rlist1 := []string{}
	for _, rlist := range cinfo.Room.RoomChoices {
		rlist1a := []string{}
		for _, rref := range rlist {
			rlist1a = append(rlist1a, ttinfo.Ref2Tag[rref])
		}
		rlist1 = append(rlist1, strings.Join(rlist1a, "|"))
	}
	r1 := strings.Join(rlist1, " + ")

	return fmt.Sprintf("<Course %s/%s:%s [%s & %s]/>",
		strings.Join(glist, ","),
		strings.Join(tlist, ","),
		ttinfo.Ref2Tag[cinfo.Subject],
		r0,
		r1,
	)
}

// collectCourses gathers the courses ([base.Course] and [base.SuperCourse])
// elements with lessons, retaining the order of the source structures.
// The [CourseInfo] items generated for the supercourses combine the
// resources of their subcourses.
func (ttinfo *TtInfo) collectCourses() {
	ttinfo.CourseInfo = map[Ref]*CourseInfo{}
	roomData := map[Ref][]Ref{} // course -> []room (any sort of "room")
	db := ttinfo.Db

	// Create the CourseInfo items.
	// Gather first the SuperCourses, then the Courses.

	cinfo_list := []*CourseInfo{}
	clessons := [][]Ref{}
	for _, spc := range db.SuperCourses {
		cref := spc.Id
		groups := []Ref{}
		teachers := []Ref{}
		rooms := []Ref{}
		for _, sbcref := range spc.SubCourses {
			sbc := db.Elements[sbcref].(*base.SubCourse)
			// Add groups
			if len(sbc.Groups) != 0 {
				groups = append(groups, sbc.Groups...)
			}
			// Add teachers
			if len(sbc.Teachers) != 0 {
				teachers = append(teachers, sbc.Teachers...)
			}
			// Add rooms
			if sbc.Room != "" {
				rooms = append(rooms, sbc.Room)
			}
		}
		// Eliminate duplicates by sorting and then compacting
		slices.Sort(groups)
		slices.Sort(teachers)
		slices.Sort(rooms)
		cinfo_list = append(cinfo_list, &CourseInfo{
			Id:       cref,
			Subject:  spc.Subject,
			Groups:   slices.Compact(groups),
			Teachers: slices.Compact(teachers),
			//Room: filled later
			Lessons: []*base.Lesson{},
		})
		clessons = append(clessons, spc.Lessons)
		roomData[cref] = slices.Compact(rooms)
	}
	for _, c := range db.Courses {
		cref := c.Id
		rooms := []Ref{}
		if c.Room != "" {
			rooms = append(rooms, c.Room)
		}
		cinfo_list = append(cinfo_list, &CourseInfo{
			Id:       cref,
			Subject:  c.Subject,
			Groups:   c.Groups,
			Teachers: c.Teachers,
			//Room: filled later
			Lessons: []*base.Lesson{},
		})
		clessons = append(clessons, c.Lessons)
		roomData[cref] = rooms
	}

	// Retain this ordered list of courses (with lessons)
	ttinfo.LessonCourses = cinfo_list

	for i, cinfo := range cinfo_list {
		// Add lessons to CourseInfo
		llist := clessons[i]
		for _, lref := range llist {
			l := db.Elements[lref].(*base.Lesson)
			// A stand-in lesson can be in any student group, so set
			// the Groups field to nil to indicate this.
			//TODO: Explain why this is necessary!
			if slices.Contains(l.Flags, "SubstitutionService") {
				cinfo.Groups = nil
			}
			cinfo.Lessons = append(cinfo.Lessons, l)
		}

		// Add to CourseInfo map
		ttinfo.CourseInfo[cinfo.Id] = cinfo
	}

	ttinfo.filterRoomData(roomData)
}

// filterRoomData prepares the internal room structures, filtering the room
// lists of the SuperCourses.
func (ttinfo *TtInfo) filterRoomData(roomData map[Ref][]Ref) {
	db := ttinfo.Db

	for cref, crlist := range roomData {
		// Join all Rooms and the Rooms from RoomGroups into a "compulsory"
		// list. Then go through the RoomChoiceGroups. If one contains a
		// compulsory room, ignore the choice.
		// The result is a list of Rooms and a list of Room-choice-lists,
		// which can be combined in a [VirtualRoom].
		rooms := []Ref{}
		roomChoices := [][]Ref{}
		for _, rref := range crlist {
			rx := db.Elements[rref]
			_, ok := rx.(*base.Room)
			if ok {
				rooms = append(rooms, rref)
			} else {
				rg, ok := rx.(*base.RoomGroup)
				if ok {
					rooms = append(rooms, rg.Rooms...)
				} else {
					rc, ok := rx.(*base.RoomChoiceGroup)
					if !ok {
						base.Bug.Fatalf(
							"Invalid room in course %s:\n  %s\n",
							cref, rref)
					}
					roomChoices = append(roomChoices, rc.Rooms)
				}
			}
		}

		// Filter out any "necessary" rooms

	restart:
		//fmt.Printf("roomChoices1: %d\n", len(roomChoices))

		/*
			{ // Print roomChoices
				rlist1 := []string{}
				for _, rlist := range roomChoices {
					rlist1a := []string{}
					for _, rref := range rlist {
						rlist1a = append(rlist1a, ttinfo.Ref2Tag[rref])
					}
					rlist1 = append(rlist1, strings.Join(rlist1a, "|"))
				}
				slices.Sort(rlist1)
				r1 := strings.Join(rlist1, " + ")
				fmt.Printf(" 00> %s\n", r1)
			}
		*/
		{
			rclList := [][]Ref{}
			for _, rcl0 := range roomChoices {
				rcl := []Ref{}
				for _, rc := range rcl0 {
					// Using list search will be slow for long lists, but
					// most lists will be very short, which should
					// compensate for that.
					if slices.Contains(rooms, rc) {
						continue
					}
					rcl = append(rcl, rc)
				}
				if len(rcl) == 0 {
					// Skip choice
					continue
				}
				if len(rcl) == 1 {
					rooms = append(rooms, rcl[0])
					goto restart
				}
				rclList = append(rclList, rcl)
			}
			//fmt.Printf("roomChoices2: %d\n", len(rclList))
			roomChoices = rclList
			//ttinfo.pRooms(rooms, "$ Rooms", ",")
		}

		// Convert rooms to indexes to ease handling

		rmap := map[Ref]int{} // room id -> index
		rvec := []Ref{}       // room index -> id
		rindex := 0
		rilList := [][]int{}
		//for i, rcl := range roomChoices {
		for _, rcl := range roomChoices {
			//ttinfo.pRooms(rcl, fmt.Sprintf("$ RC0 %d", i), ",")
			ril := []int{}
			for _, rc := range rcl {
				// Index the room
				ri, ok := rmap[rc]
				if !ok {
					ri = rindex
					rmap[rc] = ri
					rvec = append(rvec, rc)
					rindex++
				}
				ril = append(ril, ri)
			}
			rilList = append(rilList, ril)
			//ttinfo.pRoomsI(ril, rvec, fmt.Sprintf("$ RC1 %d", i), ",")
		}

		// Investigate all paths, seeking and counting clashes per room

		rclashes := map[int]int{}
		paths := []*big.Int{}
		for _, ril := range rilList {

			if len(paths) == 0 {
				// Build initial paths
				for _, ri := range ril {
					b := big.NewInt(0)
					b.SetBit(b, ri, 1)
					paths = append(paths, b)
				}
				continue
			}

			// Extend paths
			paths1 := []*big.Int{}
			for _, b := range paths {
				for _, ri := range ril {
					if b.Bit(ri) == 0 {
						b1 := big.NewInt(0)
						b1.SetBit(b, ri, 1)
						// Skip duplicates
						for _, b0 := range paths1 {
							if b1.Cmp(b0) == 0 {
								goto skip1
							}
						}
						paths1 = append(paths1, b1)
					} else {
						// Clash, register and skip path
						rclashes[ri]++
					}
				skip1:
				}
			}
			if len(paths1) == 0 {
				// No paths left, make room with most clashes "necessary"
				rcm := -1
				nmax := 0
				for rc, n := range rclashes {
					if n > nmax {
						rcm = rc
					}
				}
				rooms = append(rooms, rvec[rcm])
				goto restart
			}
			if len(paths1) == 1 {
				// Make all rooms in paths1 "necessary"
				path := paths1[0]
				for ri := 0; ri < rindex; ri++ {
					if path.Bit(ri) == 1 {
						rooms = append(rooms, rvec[ri])
					}
				}
				goto restart
			}
			paths = paths1
		}

		// Add virtual room to CourseInfo item

		cinfo := ttinfo.CourseInfo[cref]
		cinfo.Room = VirtualRoom{
			Rooms:       rooms,
			RoomChoices: roomChoices,
		}
	}
}

func (ttinfo *TtInfo) checkAllocatedRooms(cinfo *CourseInfo) {
	// Check the room allocations for the lessons of the given course.
	// Unallocated room requirements are replaced by "", so that
	// the lesson's Rooms field has the correct length.

	for _, cinfo := range ttinfo.LessonCourses {
		vr := cinfo.Room
		nRooms := len(vr.Rooms) + len(vr.RoomChoices)
		// Handle each lesson independently
		for _, l := range cinfo.Lessons {
			lrooms := l.Rooms // the initial room allocation list
			l.Rooms = make([]base.Ref, nRooms)
			// If no rooms are allocated, don't regard this as invalid.
			if len(lrooms) == 0 {
				return
			}
			complete := true // whether all requirements are fulfilled
			// Check "compulsory" rooms.
			// First collect all allocated rooms in a set.
			lrmap := map[Ref]bool{}
			for _, rref := range lrooms {
				lrmap[rref] = true
			}
			// Remove those that belong to the "compulsory" list
			for i, rref := range vr.Rooms {
				if lrmap[rref] {
					delete(lrmap, rref)
					l.Rooms[i] = rref
				} else {
					// This "compulsory" room is not allocated
					complete = false
				}
			}

			// Check validity of "chosen" rooms, those remaining in lrmap.

			nrc := len(vr.RoomChoices) // number of chosen rooms required
			nl := len(lrmap)           // number of rooms still to check
			if nl < nrc {
				complete = false
			}
			cmap := make([]Ref, 0, nrc)

			// Use a recursive function to try to match the rooms to the choice
			// lists. The parameter i is the index of the vr.RoomChoices list,
			// n is the number of unmatched rooms from lrmap.

			nmin := nl                      // best result so far
			rmin := l.Rooms[len(vr.Rooms):] // rooms for best result so far
			// When the remaining number of unallocated rooms reaches nend,
			// a stoppping point has been reached.
			// reached
			nend := 0
			if nl > nrc {
				nend = nl - nrc
			}

			var fx func(i int, n int) bool
			fx = func(i int, n int) bool {
				if i == nrc {
					if n == nend {
						// An "optimal" selection has been found
						nmin = n
						copy(rmin, cmap)
						return true
					}
					if n < nmin {
						// An improvement on the current best ...
						nmin = n
						copy(rmin, cmap)
					}
					// ... but still not guaranteed optimal
					return false
				}

				// Look for a room in the choice list which is still in the
				// list of chosen rooms (lrmap)
				for _, rref := range vr.RoomChoices[i] {
					if lrmap[rref] && !slices.Contains(cmap, rref) {
						cmap = append(cmap, rref)
						// Continue search with next choice list
						if fx(i+1, n-1) {
							// The match worked
							return true
						}
						// The rest of the rooms didn't match, or weren't
						// optimal. Try again with the next room in the
						// choice list.
						cmap = cmap[:i]
					}
				}
				// No match found in this choice list
				cmap = append(cmap, "")
				return fx(i+1, n)
			}

			if !fx(0, nl) {
				complete = false
			}

			rlist := []string{}
			for _, rref := range rmin {
				delete(lrmap, rref)
			}
			for rref := range lrmap {
				rlist = append(rlist, ttinfo.Ref2Tag[rref])
			}
			if len(rlist) != 0 {
				base.Warning.Printf("Lesson in Course %s has invalid"+
					" room allocations: %+v\n",
					ttinfo.View(cinfo), rlist)
			} else if !complete {
				base.Warning.Printf("Lesson in Course %s has incomplete"+
					" room allocations\n",
					ttinfo.View(cinfo))
			}
		}
	}
}

// collectCourseResources collects resource-indexes for each course, setting
// the Resources field of the [CourseInfo] items.
func (ttinfo *TtInfo) collectCourseResources() {
	g2tt := ttinfo.AtomicGroupIndexes
	t2tt := ttinfo.TeacherIndexes
	r2tt := ttinfo.RoomIndexes
	for _, cinfo := range ttinfo.CourseInfo {
		resources := []ResourceIndex{}

		for _, tref := range cinfo.Teachers {
			resources = append(resources, t2tt[tref])
		}

		//TODO: Is this really useful?
		// Get class(es) ... and atomic groups
		// This is for finding the "extended groups" – in the activity's
		// class(es) but not involved in the activity. This list may help
		// finding activities which can be placed parallel.
		cagmap := map[base.Ref][]ResourceIndex{}
		for _, gref := range cinfo.Groups {
			cagmap[ttinfo.Db.Elements[gref].(*base.Group).Class] = nil
		}
		aagmap := map[ResourceIndex]bool{}
		for cref := range cagmap {
			c := ttinfo.Db.Elements[cref].(*base.Class)
			aglist := g2tt[c.ClassGroup]
			//fmt.Printf("???????? %s: %+v\n", c.Tag, aglist)
			for _, agix := range aglist {
				aagmap[agix] = true
			}
		}
		//--?

		// Handle groups
		for _, gref := range cinfo.Groups {
			for _, agix := range g2tt[gref] {
				// Check for repetitions
				if slices.Contains(resources, agix) {
					base.Warning.Printf(
						"Lesson with repeated atomic group"+
							" in Course: %s\n", ttinfo.View(cinfo))
				} else {
					resources = append(resources, agix)
					aagmap[agix] = false
				}
			}
		}

		//TODO: What, if anything, to do with this?
		extendedGroups := []ResourceIndex{}
		for agix, ok := range aagmap {
			if ok {
				extendedGroups = append(extendedGroups, agix)
			}
		}

		crooms := cinfo.Room.Rooms
		for _, rref := range crooms {
			// Only take the compulsory rooms here
			resources = append(resources, r2tt[rref])
		}
		cinfo.Resources = resources
	}
}

// for testing
func (ttinfo *TtInfo) pRooms(
	rlist []Ref,
	prefix string,
	jn string,
) {
	rlist1 := []string{}
	for _, rref := range rlist {
		rlist1 = append(rlist1, ttinfo.Ref2Tag[rref])
	}
	slices.Sort(rlist1)
	r1 := strings.Join(rlist1, jn)
	fmt.Printf("%s %s\n", prefix, r1)
}

// for testing
func (ttinfo *TtInfo) pRoomsI(
	rlist []int,
	rvec []Ref,
	prefix string,
	jn string,
) {
	rlist1 := []string{}
	for _, rix := range rlist {
		rlist1 = append(rlist1, ttinfo.Ref2Tag[rvec[rix]])
	}
	slices.Sort(rlist1)
	r1 := strings.Join(rlist1, jn)
	fmt.Printf("%s %s\n", prefix, r1)
}
