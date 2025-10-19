package autotimetable

import (
	"W365toFET/timetable"
)

// Structures and methods used in connection with automation of the
// timetable generation.

var Ticks int // global time ticker
// The instance tick counter is in `TtData` because it may be needed
// by the back-end.
var TtData_0 *timetable.TtData // the original data

type TtInstance struct {
	Timeout int               // ticks
	TtData  *timetable.TtData // current (possibly modified) data

	// Base data for this instance:
	BaseInstance *TtInstance
	// Matrix detailing inclusion state of each constraint in the base
	// instance:
	HardConstraintEnabledMatrix [][]bool // [type][index] -> enabled
	SoftConstraintEnabledMatrix [][]bool // [type][index] -> enabled
	// Constraints to be added in this instance:
	ConstraintType timetable.ConstraintType
	Constraints    []int // individual constraint indexes

	// Run time
	Stopped         bool // `abort_instance()` has been called on this instance
	ProcessingState int  // -1: queued, 0: running, 1: success, 2: failure,
	// there is also 3: cancelled
}

type ManageRun struct {
	Instance *TtInstance
	Status   int
}

type GlobalData struct {
	Ticks    int
	TtData_0 *timetable.TtData // original data
	//Instances   []*TtInstance
	//NewInstance chan *TtInstance // send here to request run start
}

/* TODO?
type TtHandler interface {
	Update(*TtInstance)
}
*/

/* TODO?
type TtChainedFunc struct {
	// `Delay` specifies the number of ticks after which the function should
	// be called. 0 specifies an infinite delay (i.e. the function must be
	// started in some other way), -1 specifies that the function will never
	// be called – possibly because it has already been started.
	Delay int
	Func  func(*TtInstance) *TtInstance
}

type NewState struct {
	State   int
	Message string
}

// TODO: Which bits are still needed?
type SearchInfo struct {
	// This assumes a contiguous range of constraint indexes (from `index0`).
	Constraint int
	Index0     int
	Enabled    []bool // with an entry for each constraint to be tested
	Part       int    // 1 or 2
	Done       int
	//TODO--?
	NextPath TtChainedFunc
}
*/
