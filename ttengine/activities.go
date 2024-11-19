package ttengine

import (
	"W365toFET/ttbase"
)

type SlotIndex int
type ResourceIndex = ttbase.ResourceIndex

type Activity struct {
	Index         int
	Duration      int
	Resources     []ResourceIndex
	PossibleSlots []SlotIndex
	Fixed         bool
	Placement     int // day * nhours + hour, or -1 if unplaced
}

func (tt *TtCore) addActivities(
	t2tt map[Ref]ResourceIndex,
	r2tt map[Ref]ResourceIndex,
	g2tt map[Ref][]ResourceIndex,
	lessons []ttbase.TtLesson,
) {
	// Construct the Activities from the ttinfo.TtLessons.
	for i, ttl := range lessons {
		l := ttl.Lesson
		p := -1
		if l.Day >= 0 {
			p = l.Day*tt.NDays + l.Hour
		}
		cinfo := ttl.CourseInfo

		//TODO

		tt.Activities[i+1] = &Activity{
			Index:    i + 1,
			Duration: l.Duration,
			//Resources: TODO,
			//PossibleSlots: TODO,
			Fixed:     l.Fixed,
			Placement: p,
		}
	}
}

/*TODO--
type TtLesson struct {
	Index      LessonIndex
	CourseInfo *CourseInfo
	Lesson     *base.Lesson
}

type CourseInfo struct {
	Subject  Ref
	Groups   []Ref
	Teachers []Ref
	Room     VirtualRoom
	Lessons  []LessonIndex
}
*/
