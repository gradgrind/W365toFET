package ttbase

/*TODO: Integrate into the new activity group processing

func (ttinfo *TtInfo) addConstraints() {
	//TODO: The rest handles lessons, so it will need modifying for activity
	// groups ...

	for _, cinfo := range ttinfo.LessonCourses {
		cref := cinfo.Id
		// Determine groups of lessons to couple by means of the fixed flags.
		fixeds := []int{}
		unfixeds := []int{}
		for _, lref := range cinfo.Lessons {
			if ttinfo.Activities[lref].Fixed {
				fixeds = append(fixeds, lref)
			} else {
				unfixeds = append(unfixeds, lref)
			}
		}

		ddcs, ddcsok := ttinfo.DayGapConstraints.CourseConstraints[cref]
		if len(unfixeds) == 0 || (len(fixeds) == 0 && len(unfixeds) == 1) {
			// No constraints necessary
			if ddcsok {
				base.Warning.Printf("Superfluous DaysBetween constraint on"+
					" course:\n  -- %s", ttinfo.View(cinfo))
			}
			continue
		}
		aidlists := [][]int{}
		if len(fixeds) <= 1 {
			aidlists = append(aidlists, cinfo.Lessons)
		} else {
			for _, aid := range fixeds {
				aids := []int{aid}
				aids = append(aids, unfixeds...)
				aidlists = append(aidlists, aids)
			}
		}

		// Add constraints

		//TODO: ddays is, in effect, integrated into
		// ttinfo.DayGapConstraints.CourseConstraints
		if !ddays[cref] && diffDays.weight != 0 {
			// Add default constraint
			for _, alist := range aidlists {
				if len(alist) > ttinfo.NDays {
					base.Warning.Printf("Course has too many lessons for"+
						"DifferentDays constraint:\n  -- %s\n",
						ttinfo.View(cinfo))
					continue
				}
				mdba = append(mdba, MinDaysBetweenLessons{
					Weight:               diffDays.weight,
					ConsecutiveIfSameDay: diffDays.consecutiveIfSameDay,
					Lessons:              alist,
					MinDays:              1,
				})
			}
		}

		if ddcsok {
			// Generate the additional constraints
			for _, ddc := range ddcs {
				if ddc.Weight != 0 {
					for _, alist := range aidlists {
						if (len(alist)-1)*ddc.DayGap >= ttinfo.NDays {
							base.Warning.Printf("Course has too many lessons"+
								" for DaysBetween constraint:\n  -- %s\n",
								ttinfo.View(cinfo))
							continue
						}
						mdba = append(mdba, MinDaysBetweenLessons{
							Weight:               ddc.Weight,
							ConsecutiveIfSameDay: ddc.ConsecutiveIfSameDay,
							Lessons:              alist,
							MinDays:              ddc.DayGap,
						})
					}
				}
			}
		}
	}
	ttinfo.MinDaysBetweenLessons = mdba
}
*/
