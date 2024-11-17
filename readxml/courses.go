package readxml

import (
	"W365toFET/base"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type tmpCourse struct {
	Id             Ref
	Subjects       []Ref
	Groups         []Ref
	Teachers       []Ref
	PreferredRooms []Ref
}

type xCourse struct {
	lessons []int
	super   *tmpCourse
	subs    []*tmpCourse
}

func (cdata *conversionData) readCourses() map[Ref][]int {
	courseLessons := map[Ref][]int{} // build return value here

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
				Id:             n.Id,
				Subjects:       cdata.getCourseSubjects(n),
				Groups:         grps,
				Teachers:       tchs,
				PreferredRooms: rms,
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
			outdata.Courses = append(outdata.Courses, &base.Course{
				Id:             nid,
				Subjects:       sbjs,
				Groups:         grps,
				Teachers:       tchs,
				PreferredRooms: rms,
			})
			courseLessons[nid] = llen
		} // else Course with no lessons

	}

	//TODO??

	return courseLessons
}

func (cdata *conversionData) getCourseSubject(c *Course) Ref {
	if c.Subjects == "" {
		base.Error.Fatalf("In Course %s:\n  -- No Subject\n", c.Id)
	}
	slist := []Ref{}
	for _, ref := range splitRefList(c.Subjects) {
		s, ok := cdata.db.Elements[ref]
		if ok {
			_, ok := s.(*base.Subject)
			if ok {
				slist = append(slist, ref)
				continue
			}
		}
		base.Error.Fatalf("In Course %s:\n  -- Invalid Subject: %s\n",
			c.Id, ref)
	}
	if len(slist) == 1 {
		return slist[0]
	}

}

// TODO: See courseSubject in subjects.go!
func (cdata *conversionData) makeCompoundSubject(
	srefs []Ref,
	courseId Ref,
) Ref {
	db := cdata.db
	// Make a subject name
	sklist := []string{}
	for _, sref := range srefs {
		// Need Tag/Shortcut field
		s, ok := db.Elements[sref]
		if ok {
			ss, ok := s.(*base.Subject)
			if ok {
				sklist = append(sklist, ss.Tag)
				continue
			}

		}
		base.Bug.Fatalf("In Course %s:\n  -- Invalid Subject: %s\n",
			courseId, sref)
	}
	sktag := strings.Join(sklist, ",")
	sref, ok := cdata.subjectTags[sktag]
	if !ok {
		// Need a new Subject.

		sref = db.makeNewSubject(newdb, sktag, "Compound Subject")
		cdata.subjectTags[sktag] = sref
	}
	return sref
}

//TODO--

func readCourses0(
	outdata *base.DbTopLevel,
	id2node map[Ref]interface{},
	items []Course,
) map[Ref][]int {
	courseLessons := map[Ref][]int{} // build return value here
	xcourses := map[string]xCourse{}
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		msg := fmt.Sprintf("Course %s in Subjects", nid)
		sbjs := cdata.getRefList(n.Subjects, msg)
		msg = fmt.Sprintf("Course %s in Groups", nid)
		grps := cdata.getRefList(n.Groups, msg)
		msg = fmt.Sprintf("Course %s in Teachers", nid)
		tchs := cdata.getRefList(n.Teachers, msg)
		msg = fmt.Sprintf("Course %s in PreferredRooms", nid)
		rms := cdata.getRefList(n.PreferredRooms, msg)

		// Get lesson lengths
		llen := []int{}
		if n.SplitHoursPerWeek != "" {
			hpw := strings.Split(n.SplitHoursPerWeek, "+")
			for _, l := range hpw {
				if l != "" {
					ll, err := strconv.Atoi(l)
					if err != nil {
						log.Fatalf("*ERROR* Course %s:\n"+
							"  ++ SplitHoursPerWeek = %s\n",
							nid, n.SplitHoursPerWeek)
					}
					llen = append(llen, ll)
				}
			}
		} else if n.HoursPerWeek != 0.0 {
			llen = append(llen, int(n.HoursPerWeek))
		} // else no lessons

		blockTag := getBlockTag(id2node, n.Categories, nid)
		if blockTag != "" {
			//
			// The existence of  block tag makes this into a SuperCourse
			// or a SubCourse:
			// If the Course has a SplitHoursPerWeek entry, this
			// defines the lesson lengths. If not, and HoursPerWeek
			// is given, take this as the length of a single lesson.
			// If neither is present, take this as a SubCourse.
			// Any teachers, groups or rooms of a SuperCourse will
			// be ignored?
			//
			tcourse := &tmpCourse{
				Id:             nid,
				Subjects:       sbjs,
				Groups:         grps,
				Teachers:       tchs,
				PreferredRooms: rms,
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
			outdata.Courses = append(outdata.Courses, &base.Course{
				Id:             nid,
				Subjects:       sbjs,
				Groups:         grps,
				Teachers:       tchs,
				PreferredRooms: rms,
			})
			courseLessons[nid] = llen
		} // else Course with no lessons
	}
	for _, xc := range xcourses {
		//for key, xc := range xcourses {
		//fmt.Printf("\n *** XCOURSE: %s\n  %+v\n", key, xc)

		//TODO: Multiple SuperCourses in a SubCourse

		sc := xc.super
		scid := sc.Id
		//TODO: The SuperCourse may have only one subject
		outdata.SuperCourses = append(outdata.SuperCourses, &base.SuperCourse{
			Id:      scid,
			Subject: sc.Subjects[0],
			// All other fields are ignored.
		})
		courseLessons[scid] = xc.lessons

		// Now the SubCourses
		for _, sbc := range xc.subs {
			outdata.SubCourses = append(outdata.SubCourses, &base.SubCourse{
				// Note the use of Id0 instead of Id (see base.structures.go)
				Id0:            sbc.Id,
				SuperCourses:   []Ref{scid},
				Subjects:       sbc.Subjects,
				Groups:         sbc.Groups,
				Teachers:       sbc.Teachers,
				PreferredRooms: sbc.PreferredRooms,
			})
		}
	}
	return courseLessons
}
