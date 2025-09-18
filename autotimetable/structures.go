package autotimetable

import (
	"W365toFET/timetable"
	"sync"
)

// Structures and methods used in connection with automation of the
// timetable generation.

type TtInstance struct {
	Global *GlobalData
	Id     int
	//Ticks       int
	Delay   int // ticks
	Timeout int // ticks

	// `Termination` is normally 0 (not terminated "internally", i.e. from
	// the tick-loop). Before a timeout is sent, this value is set to 1.
	// Before a deletion is sent, this value is set to -1.
	Termination int

	TtData *timetable.TtData // current (possibly modified) data

	// Communication channels
	//Stop        chan bool
	WaitGroup *sync.WaitGroup

	//TODO: clarify, see tick loop!
	// `State` values:
	//		 0: running
	//     	 1: finished successfully
	//		 2: failed
	//		 3: process aborted
	//       4: other incomplete termination
	//		 5: cancelled by `cancelAll`
	//		-1: timeout (awaiting completion)
	//State    int
	//Progress int // percentage of activities which have been placed
	// `LastTime` is the `Ticks` value at which the `Progress` field was
	// last updated.
	//LastTime int
	// Record the enablement status of each constraint:
	//ConstraintEnableMatrix [][]bool
	//
	HardConstraintEnabled map[timetable.ConstraintType]map[int]bool

	// Collate intermediate test results:
	//SearchInfo *SearchInfo
	// `HandlerData` provides a field to be used by the timetable "back-end".
	//HandlerData     any ... TODO: move to TtData?

	/*TODO???
	UpdateHandler func(instance *TtInstance)

	Message string // completion information

	Abort func(any) // pass HandlerData

	SuccessPath     TtChainedFunc
	SuccessInstance *TtInstance
	FailurePath     TtChainedFunc
	FailureInstance *TtInstance
	OtherPaths      []TtChainedFunc
	OtherInstances  []*TtInstance
	*/
}

type GlobalData struct {
	Ticks       int
	TtData_0    *timetable.TtData // original data
	Instances   []*TtInstance
	NewInstance chan *TtInstance // send here to request run start
}

// ?
type TtHandler interface {
	Update(*TtInstance)
}

// ?
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
