package autotimetable

import (
	"W365toFET/timetable"
	"sync"
)

// Structures and methods used in connection with automation of the
// timetable generation.

type TtInstance struct {
	Global  *GlobalData
	Id      int
	Delay   int // ticks
	Timeout int // ticks

	// `Termination` is normally 0 (not terminated "internally", i.e. from
	// the tick-loop). Before a timeout is sent, this value is set to 1.
	// Before a deletion is sent, this value is set to -1.
	Termination int

	TtData *timetable.TtData // current (possibly modified) data

	WaitGroup *sync.WaitGroup // for waiting until all goroutines finish

	HardConstraintEnabled [][]bool // [type][index] -> enabled

	// To be added in this instance:
	ConstraintType timetable.ConstraintType
	Constraints    []int // individual constraint indexes

	// Run time
	Instance0 *TtInstance
	Instance1 *TtInstance
}

type GlobalData struct {
	Ticks       int
	TtData_0    *timetable.TtData // original data
	Instances   []*TtInstance
	NewInstance chan *TtInstance // send here to request run start
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
