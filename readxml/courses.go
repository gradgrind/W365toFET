package readxml

import (
	"W365toFET/base"
	"log"
	"strconv"
	"strings"
)

type tmpCourse struct {
	Id       Ref
	Subject  Ref
	Groups   []Ref
	Teachers []Ref
	Room     Ref
}

type xCourse struct {
	lessons []int
	super   *tmpCourse
	subs    []*tmpCourse
}

func (cdata *conversionData) readCourses() map[Ref][]int {
	courseLessons := map[Ref][]int{} // build return value here
	xcourses := map[string]xCourse{}
	db := cdata.db

	for i := 0; i < len(cdata.xmlin.Courses); i++ {
		n := &cdata.xmlin.Courses[i]

		// Get lesson lengths.
		// The Course should have a SplitHoursPerWeek entry. This
		// defines the lesson lengths. If not, and HoursPerWeek
		// is given, use this as the number of single lessons.
		llen := []int{}
		if n.SplitHoursPerWeek != "" {
			hpw := strings.Split(n.SplitHoursPerWeek, "+")
			for _, l := range hpw {
				if l != "" {
					ll, err := strconv.Atoi(l)
					if err != nil {
						base.Error.Fatalf(" In Course %s:\n"+
							"  -- SplitHoursPerWeek = %s\n",
							n.Id, n.SplitHoursPerWeek)
					}
					llen = append(llen, ll)
				}
			}
		}
		if n.HoursPerWeek != 0.0 {
			if n.SplitHoursPerWeek != "" {
				base.Error.Fatalf("In Course %s:\n"+
					"  -- Entries for SplitHoursPerWeek AND HoursPerWeek",
					n.Id)
			}
			base.Warning.Printf("In Course %s:\n"+
				"  -- HoursPerWeek specified\n", n.Id)
			for i := 0; i < int(n.HoursPerWeek); i++ {
				llen = append(llen, 1)
			}
		} // else no lessons

		subject := cdata.getCourseSubject(n)
		groups := cdata.getCourseGroups(n)
		teachers := cdata.getCourseTeachers(n)
		room := cdata.getCourseRoom(n)

		// Handle Block tags defined as Categories
		blockTag := cdata.getBlockTag(n.Categories, n.Id)
		if blockTag != "" {
			//
			// The existence of  block tag makes this into a SuperCourse
			// or a SubCourse:
			// If no lesson lengths are specified, take this as a SubCourse.
			// Any teachers, groups or rooms of a SuperCourse will
			// be ignored.
			//
			tcourse := &tmpCourse{
				Id:       n.Id,
				Subject:  subject,
				Groups:   groups,
				Teachers: teachers,
				Room:     room,
			}
			if len(llen) == 0 {
				// A SubCourse.
				xc := xcourses[blockTag]
				xc.subs = append(xc.subs, tcourse)
				xcourses[blockTag] = xc
			} else {
				// A SuperCourse
				xc := xcourses[blockTag]
				if xc.super != nil {
					log.Fatalf("*ERROR* Block with two"+
						" SuperCourses: %s\n",
						blockTag)
				}
				xc.super = tcourse
				xc.lessons = llen
				xcourses[blockTag] = xc
			}
		} else if len(llen) != 0 {
			e := db.NewCourse(n.Id)
			e.Subject = subject
			e.Groups = groups
			e.Teachers = teachers
			e.Room = room
			courseLessons[n.Id] = llen
		} // else Course with no lessons

	}

	//TODO
	for _, xc := range xcourses {
		//for key, xc := range xcourses {
		//fmt.Printf("\n *** XCOURSE: %s\n  %+v\n", key, xc)

		//TODO: Multiple SuperCourses in a SubCourse

		spct := xc.super
		scid := spct.Id
		//TODO: The SuperCourse may have only one subject
		spc := db.NewSuperCourse(scid)
		spc.Subject = spct.Subject
		// All other fields are ignored.
		courseLessons[scid] = xc.lessons

		// Now the SubCourses
		for _, sbct := range xc.subs {
			sbc := db.NewSubCourse(sbct.Id)
			sbc.SuperCourses = []Ref{scid}
			sbc.Subject = sbc.Subject
			sbc.Groups = sbc.Groups
			sbc.Teachers = sbc.Teachers
			sbc.Room = sbc.Room
		}
	}
	return courseLessons
}
