package readxml

import (
	"W365toFET/w365tt"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type tmpCourse struct {
	Id             w365tt.Ref
	Subjects       []w365tt.Ref
	Groups         []w365tt.Ref
	Teachers       []w365tt.Ref
	PreferredRooms []w365tt.Ref
}

type xCourse struct {
	lessons []int
	super   *tmpCourse
	subs    []*tmpCourse
}

func readCourses(
	outdata *w365tt.DbTopLevel,
	id2node map[w365tt.Ref]interface{},
	items []Course,
) map[w365tt.Ref][]int {
	courseLessons := map[w365tt.Ref][]int{} // build return value here
	xcourses := map[string]xCourse{}
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		msg := fmt.Sprintf("Course %s in Subjects", nid)
		sbjs := GetRefList(id2node, n.Subjects, msg)
		msg = fmt.Sprintf("Course %s in Groups", nid)
		grps := GetRefList(id2node, n.Groups, msg)
		msg = fmt.Sprintf("Course %s in Teachers", nid)
		tchs := GetRefList(id2node, n.Teachers, msg)
		msg = fmt.Sprintf("Course %s in PreferredRooms", nid)
		rms := GetRefList(id2node, n.PreferredRooms, msg)

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

		if n.Categories != "" {
			msg := fmt.Sprintf("Category in Course %s", nid)
			reflist := GetRefList(id2node, n.Categories, msg)
			if len(reflist) != 0 {
				//fmt.Printf("Categories in Course %s:\n", nid)
				for _, cat := range reflist {
					catnode := id2node[cat].(*Category)
					//fmt.Printf("  :: %+v\n", catnode)
					if catnode.Role == 0 {
						//fmt.Printf("\n *** Course:\n%+v\n\n", n)

						// If catnode.Shortcut starts with "_", take this as
						// a block, a SuperCourse or a SubCourse:
						// If the Course has a SplitHoursPerWeek entry, this
						// defines the lesson lengths. If not, and HoursPerWeek
						// is given, take this as the length of a single lesson.
						// If neither is present, take this as a SubCourse.
						// Any teachers, groups or rooms of a SuperCourse will
						// be ignored?
						//
						if strings.HasPrefix(catnode.Shortcut, "_") {
							// Part of a block.
							tcourse := &tmpCourse{
								Id:             nid,
								Subjects:       sbjs,
								Groups:         grps,
								Teachers:       tchs,
								PreferredRooms: rms,
							}

							if len(llen) == 0 {
								// A SubCourse.
								xc := xcourses[catnode.Shortcut]
								xc.subs = append(xc.subs, tcourse)
								xcourses[catnode.Shortcut] = xc
							} else {
								// A SuperCourse
								xc := xcourses[catnode.Shortcut]
								if xc.super != nil {
									log.Fatalf("*ERROR* Block with two"+
										" SuperCourses: %s\n",
										catnode.Shortcut)
								}
								xc.super = tcourse
								xc.lessons = llen
								xcourses[catnode.Shortcut] = xc
							}
							goto next_course
						}
					}
				}
			}
		}
		if len(llen) != 0 {
			outdata.Courses = append(outdata.Courses, w365tt.Course{
				Id:             nid,
				Subjects:       sbjs,
				Groups:         grps,
				Teachers:       tchs,
				PreferredRooms: rms,
			})
			courseLessons[nid] = llen
		} // else Course with no lessons
	next_course:
	}
	for _, xc := range xcourses {
		//for key, xc := range xcourses {
		//fmt.Printf("\n *** XCOURSE: %s\n  %+v\n", key, xc)

		sc := xc.super
		scid := sc.Id
		//TODO: The SuperCourse may have only one subject
		outdata.SuperCourses = append(outdata.SuperCourses, w365tt.SuperCourse{
			Id:      scid,
			Subject: sc.Subjects[0],
			// All other fields are ignored.
		})
		courseLessons[scid] = xc.lessons

		// Now the SubCourses
		for _, sbc := range xc.subs {
			outdata.SubCourses = append(outdata.SubCourses, w365tt.SubCourse{
				// Note the use of Id0 instead of Id (see w365tt.structures.go)
				Id0:            sbc.Id,
				SuperCourse:    scid,
				Subjects:       sbc.Subjects,
				Groups:         sbc.Groups,
				Teachers:       sbc.Teachers,
				PreferredRooms: sbc.PreferredRooms,
			})
		}
	}
	return courseLessons
}
