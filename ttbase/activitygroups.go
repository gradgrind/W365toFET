package ttbase

// An ActivityGroup manages placement of the lessons of a course and
// any hard-parallel courses.
type ActivityGroup struct {
	Resources          []ResourceIndex
	LessonUnits        []*TtLesson //TODO: or []LessonUnitIndex?
	PossiblePlacements [][]SlotIndex
}

type TtLesson struct {
	Resources *[]ResourceIndex // points to ActivityGroup Resources?
	// ... if not dynamic, it could just be a "copy"
	Placement SlotIndex
}

// PrepareActivityGroups creates the [ActivityGroup] items from the
// [Activity] items, taking their duration, courses, hard-parallel and
// hard-different-day constraints into account.
func (ttinfo *TtInfo) PrepareActivityGroups() {

}
