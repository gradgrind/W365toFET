package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"cmp"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"syscall"
	"time"
)

// TODO: How to set this up?
var (
	UNCONSTRAINED_TIMEOUT_FRACTION = 10
	NEXT_STAGE_TIMEOUT_FACTOR      = 2
	NEXT_STAGE_TIMEOUT_MIN         = 20
)

//TODO: comment -> further below ...
// Consider starting instances with each (used) constraint type, individually.
// The first one which completes successfully could be taken as a new basis, to
// which the next completed ones could be added sequentially ... until no time is left?

var MAXPROCESSES int

//TODO: Suggestion for searches (binary or otherwise) using parallel
// operations. A structure (with pointer to it in the TtInstance) could
// contain the information needed to collate the results of a parallel run.
// Some care may be needed to avoid race conditions – perhaps the updating can
// be done in the tick handler? The follow-up could be registered for all
// bracnches, but it would only be called when all the results were.

/*
A `TtInstance` structure is constructed to manage the data for each
timetable generation run, each run having its own goroutine.

//TODO: no longer appropriate ...
Each instance can be given a set of "child" functions determining how the
tests proceed. For each of these functions a start-delay can be specified.
It is also possible to specify a function to be called on "success" of the
parent function, and one for failure. These can also be started pre-emptively
by specifying a delay.

A timetable instance can be cancelled, including all child instances. When
an instance finishes, any pre-emptively started child instances on the
branch which is now known to be wrong (success or failure) will be cancelled
automatically, including all their children. Because a run can continue
for a long time with no result, there is also a timeout function which can
stop the instance and take the "failure" branch.

The main function (`StartGeneration`) starts a run with the fully constrained
data and then enters a "tick-loop" which is triggered every second. This
monitors the progress of each active instance and handles the actions
resulting from the specified delays.
*/

var Descriptions map[string]string = map[string]string{
	"COMPLETE":           "All constraints active",
	"ONLY_BLOCKED_SLOTS": "All constraints – except blocked slots – disabled",
}

//TODO--? const ID_BLOCKED_SLOTS int = -100

func StartGeneration(tt_data_0 *timetable.TtData, TIMEOUT int) {
	// This approach relies on parallel processing. If there are too few real
	// processors it will be inefficient.
	MAXPROCESSES = max(runtime.NumCPU(), 4)
	tt_shared_data := tt_data_0.SharedData

	// Catch termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//TODO-- Wait group to ensure all goroutines finish before exiting
	//var wg sync.WaitGroup

	runqueue := RunQueue{
		Queue:      nil,
		Active:     map[*TtInstance]struct{}{},
		MaxRunning: MAXPROCESSES, //TODO????
		Next:       0,
	}

	// `workingdir` provides the path to a working directory which can be used
	// freely during processing. It may or may not already exist, existing
	// contents need not be preserved during processing.

	//TODO--
	// Open communication channels
	//stop := make(chan bool)

	//TODO-- Consider buffer size and blocking ...
	//add_instance := make(chan *TtInstance, 10)
	//instance_done := make(chan *TtInstance, 10)

	tt_data_0.Description = "COMPLETE"
	workingdir := tt_shared_data.WorkingDir

	// Provide an empty working directory.
	os.RemoveAll(workingdir)
	err := os.Mkdir(workingdir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	// Global data
	Ticks = 0
	TtData_0 = tt_data_0

	// First run: all constraints enabled.
	// On successful completion, all other instances should be stopped.
	// If it fails, just this instance should be wound up. Otherwise it
	// should run until it times out, at which point any other active
	// instances should be stopped and the "best" solution at this point
	// chosen.
	full_instance := &TtInstance{
		Tagged:  true,
		Timeout: TIMEOUT,
		TtData:  tt_data_0,
	}

	// Add to run queue
	runqueue.add(full_instance)

	//TODO: Instance without soft constraints (if any)?

	// Unconstrained instance
	tt_data := new_ttdata(tt_data_0, "ONLY_BLOCKED_SLOTS")
	null_instance := &TtInstance{
		Tagged: true,

		Timeout: max(TIMEOUT/UNCONSTRAINED_TIMEOUT_FRACTION, 10), //TODO??

		TtData: tt_data,
		HardConstraintEnabled: setup_hard_constraint_map(
			tt_data_0.HardConstraints),
	}
	disable_all_constraints(null_instance.TtData)
	// Add to run queue
	runqueue.add(null_instance)

	// Start stage 0
	stage := 0

	// *** Ticker loop ***
	next_step := 0
	var basic_constraints map[*TtInstance]struct{}
	var current_instance *TtInstance
	steps := []*TtInstance{}
	long_steps := []*TtInstance{}
	ticker := time.NewTicker(time.Second)
	//TODO? defer tidy(runqueue, ticker)
	defer ticker.Stop()
	unconstrained_time := 0

	//tickloop:
	for runqueue.update_queue() != 0 {
		select {

		case ossig := <-sigChan:
			//TODO: seek the best solution so far?
			base.Message.Printf("*** SIGNAL *** %+v", ossig)
			runqueue.disable()
			cancel_instance(full_instance)
			//cancel_instance(hard_only_instance)
			cancel_instance(null_instance)
			cancel_instance(current_instance)
			continue

		case <-ticker.C:
			Ticks++
			runqueue.update_instances()

			if Ticks == TIMEOUT {
				runqueue.disable()
				if full_instance.ProcessingState == 0 {
					base.Message.Printf(
						"(TODO) [%d] TIMEOUT\n", Ticks)
					cancel_instance(full_instance)
				}
				//cancel_instance(hard_only_instance)
				cancel_instance(null_instance)

				//TODO: Can current_instance be nil here?

				if current_instance.Result == nil {
					//TODO--
					fmt.Println("!!! No result")

					cancel_instance(current_instance)

					//TODO
					//panic("TODO: Seek best result so far")

				}
				continue
			}

			if full_instance.ProcessingState == 1 {
				// Cancel all other runs and return this instance as result.
				runqueue.disable()
				//cancel_instance(hard_only_instance)
				cancel_instance(null_instance)
				cancel_instance(current_instance)
				full_instance.Result = full_instance
				current_instance = full_instance
				continue
			}

			//TODO: Special treatment if there are no constraints to add?

			if stage == 0 {
				// During stage 0 only `full_instance` and `null_instance`
				// (and perhaps hard_only_instance) are running.
				switch null_instance.ProcessingState {
				case 0:
					if null_instance.TtData.Ticks == null_instance.Timeout {
						abort_instance(null_instance)
					}
				case 1:
					// The null instance completed successfully.
					current_instance = null_instance
					unconstrained_time = null_instance.TtData.Ticks
					base.Message.Printf("(TODO) [%d] UNCONSTRAINED TIME: %d\n",
						Ticks, unconstrained_time)
					// Start trials of single constraint types.
					basic_constraints = start_basic_constraints(
						null_instance, &runqueue, unconstrained_time)
					stage = 1
				default:
					// The null instance failed.
					stage = -1
					base.Message.Printf(
						"(TODO) [%d] Unconstrained instance failed", Ticks)

					//TODO: Seek problems in the unconstrained data.
					panic("TODO")

				}
				continue
			}

			if stage == 1 {
				//TODO: These should now always deliver a Result.

				// During stage 1 we are awaiting completion of the single-
				// constraint instances: check their states.
				for bc := range basic_constraints {
					// Handle completed instance.
					if bc.ProcessingState > 0 {
						if bc.ProcessingState == 1 {
							steps = append(steps, bc)
							if next_step == 0 {
								current_instance = bc
								base.Message.Printf(
									"(TODO) First constraint: %s\n",
									current_instance.TtData.Description)
								next_step = 1
							}
						} else {
							long_steps = append(long_steps, bc)
						}
						delete(basic_constraints, bc)
					}
				}
				if len(basic_constraints) == 0 {
					// All single-constraint instances have completed.
					if next_step == 0 {
						//TODO
						panic("No successful basic constraint trials")
					}
					// Sort `long_steps` according to progress (highest
					// percentages first) and append them to `steps`.
					if len(long_steps) > 1 {
						slices.SortFunc(long_steps, func(a, b *TtInstance) int {
							return cmp.Compare(b.TtData.Progress, a.TtData.Progress)
						})
					}
					steps = append(steps, long_steps...)
					stage = 2 // all basic-constraint trials completed
				}

				//TODO--
				base.Message.Printf("$ [%d] STAGE: %d @ %d steps: %d\n",
					Ticks, stage, next_step, len(steps))
			}

			// This bit handles the phase where constraint types are being
			// added step by step.

			// The basic idea is to try with all the constraints in the list
			// enabled, up to a time limit. If this fails, try recursively
			// with the first half, then add the second half recursively to
			// that result (with a newly calculated timeout). The recursion
			// stops when there is only one constraint in the list, it either
			// being added or not (if timed out).
			// Especially if there is a difficult constraint early in the
			// list, this can take a long time, with multiple failed runs.
			// By queueing the halves when the instance starts, it might be
			// possible to get some speed-up when multiple processors are
			// available, at least with favourable data.

			//TODO: A "stuck" analysis on running instances might help to
			// reduce processing time?

			if current_instance.Result != nil {
				fmt.Printf("??? %d @ %d: %d %d\n",
					current_instance.ProcessingState, stage, next_step, len(steps))
				if next_step < len(steps) {
					// Add next constraint type
					st1 := steps[next_step]
					desc := fmt.Sprintf("C%02d~%s",
						next_step, st1.TtData.Description)
					current_instance = new_instance(
						current_instance,
						desc,
						st1.ConstraintType,
						st1.Constraints,
						max(current_instance.TtData.Ticks*NEXT_STAGE_TIMEOUT_FACTOR,
							NEXT_STAGE_TIMEOUT_MIN))
					next_step++
					current_instance.Tagged = true
					runqueue.add(current_instance)
				} else if stage == 2 {
					// No more constraint types => finished ...
					if full_instance.ProcessingState != 0 {
						//TODO? && hard_only_instance != 0
						// Nothing left to wait for
					}
				}
			}
		}
	} // tickloop: end

	//TODO: Consider also the possibility that there may be no (or only one)
	// basic constraint types.

	var result *TtInstance
	if full_instance.Result != nil {
		result = full_instance.Result
	} else {
		result = preempt_result(current_instance)
	}

	//TODO
	base.Message.Printf("RESULT: %s\n", result.TtData.Description)
}

type RunQueue struct {
	Queue      []*TtInstance
	Active     map[*TtInstance]struct{}
	MaxRunning int
	Next       int
}

// TODO?
func preempt_result(instance *TtInstance) *TtInstance {
	if instance == nil {
		return nil
	}
	if instance.Result != nil {
		return instance.Result
	}
	r := preempt_result(instance.Instance1)
	if r != nil {
		return r
	}
	r = preempt_result(instance.Instance2)
	if r != nil {
		return r
	}
	return instance.BaseInstance
}

// TODO?
func tidy(rq RunQueue, ticker *time.Ticker) {
	fmt.Printf("TIDY %d\n", len(rq.Active))
	base.Message.Printf("TIDY %d\n", len(rq.Active))

	panic("TRACE")

	//TODO: Could the ticker be used here instead of sleep?
	ticker.Stop()

	for len(rq.Active) != 0 {
		for instance := range rq.Active {
			timetable.BACKEND.Tick(instance.TtData)
			if instance.TtData.State != 0 {
				base.Message.Printf("(TODO) Finished %s %d\n",
					instance.TtData.Description, instance.TtData.State)
				rq.instance_deactivate(instance)
			} else {
				base.Message.Printf("(TODO) Waiting? %s %d\n",
					instance.TtData.Description, instance.TtData.State)
			}
		}
		fmt.Println("Sleeping")
		base.Message.Println("Sleeping")
		time.Sleep(1 * time.Second)
	}
}

func (rq *RunQueue) add(instance *TtInstance) {
	instance.ProcessingState = -1 // not started yet
	rq.Queue = append(rq.Queue, instance)
	base.Message.Printf("(TODO) [%d] Queue %s\n",
		Ticks, instance.TtData.Description)
}

func (rq *RunQueue) add_front(instance *TtInstance) {
	instance.ProcessingState = -1 // not started yet
	rq.Queue = slices.Insert(rq.Queue, rq.Next, instance)
	base.Message.Printf("(TODO) [%d] Queue Front %s\n",
		Ticks, instance.TtData.Description)
}

func (rq *RunQueue) update_instances() {
	for instance := range rq.Active {
		ttdata := instance.TtData
		base.Message.Printf("(TODO) [%d] ? ACTIVE (%d): %s\n",
			Ticks, ttdata.State, ttdata.Description)
		if ttdata.State != 0 && instance.ProcessingState < 2 {
			// This should only be possible after the call to
			// `timetable.BACKEND.Tick` below.
			panic(fmt.Sprintf("Bug, State = %d", ttdata.State))
		}
		ttdata.Ticks++
		// Among other things, update the state:
		timetable.BACKEND.Tick(ttdata)
		switch ttdata.State {
		case 0: // running, not finished
			// check for timeout
			if instance.Timeout == ttdata.Ticks {
				base.Message.Printf("(TODO) TIMEOUT [%d] %s @ %d\n",
					Ticks, ttdata.Description, ttdata.Ticks)
				// Stop instance
				abort_instance(instance)
			}

		case 1: // completed successfully
			if instance.Tagged {
				base.Message.Printf("(TODO) [%d] <<+ %s @ %d\n",
					Ticks, ttdata.Description, ttdata.Ticks)
			} else {
				base.Message.Printf("(TODO) [%d] Done %s @ %d\n",
					Ticks, ttdata.Description, ttdata.Ticks)
			}
			// Cancel subsidiary instances
			cancel_instance(instance.Instance1)
			cancel_instance(instance.Instance2)
			instance.Result = instance
			rq.instance_completed(instance, 1)

		default: // completed unsuccessfully
			if instance.ProcessingState != 2 {
				// This is done only the first time round for this instance.
				// Some instances remain active after the primary run has
				// completed, so for subsequent loops this block should be
				// skipped.
				if instance.Tagged {
					base.Message.Printf("(TODO) [%d] <<- %s @ %d\n",
						Ticks, ttdata.Description, ttdata.Ticks)
				} else {
					base.Message.Printf("(TODO) [%d] Failed %s @ %d (%d)\n",
						Ticks, ttdata.Description, ttdata.Ticks, ttdata.State)
				}
				// This doesn't deactivate the instance (for state = 2):
				rq.instance_completed(instance, 2)
			}

			//TODO: If it is an actual error, the halves should perhaps still
			// be allowed to run, even though they may not have started yet.

			if instance.Instance1 == nil {
				if instance.Instance2 == nil {
					// No halves
					instance.Result = instance.BaseInstance
					rq.instance_deactivate(instance)
				} else {
					// The 1st half has completed, the 2nd half is now
					// building, possibly based on the result of the 1st half.
					if instance.Instance2.Result != nil {
						instance.Result = instance.Instance2.Result
						rq.instance_deactivate(instance)
					}
					// Otherwise the instance remains active (though not
					// running).
				}
				continue
			}

			// else: instance.Instance1 != nil

			if instance.Instance1.Result != nil {
				// The 1st half has completed.
				// Check that it actually added constraints.
				if instance.Instance1.Result == instance.BaseInstance {
					// 1st half: no constraints added,
					// wait for completion of 2nd half.
					instance.Instance1 = nil
				} else {
					if instance.Instance2.ProcessingState < 0 {
						// The 2nd half hasn't started yet: replace its
						// base by the completed 1st half.
						instance.Instance2.BaseInstance = instance.Instance1.Result
						instance.Instance1 = nil

						// Otherwise wait for the 2nd half to finish before
						// adding it to the 1st half.
					} else if instance.Instance2.Result != nil {
						i2 := instance.Instance2.Result
						if i2 == instance.BaseInstance {
							// 2nd half: no constraints added
							instance.Result = instance.Instance1.Result
							rq.instance_deactivate(instance)
						} else {
							next_instance := new_instance(
								instance.Instance1,
								instance.Instance1.TtData.Description+"+",
								i2.ConstraintType,
								i2.Constraints,
								instance.Timeout)
							instance.Instance2 = next_instance
							instance.Instance1 = nil
							rq.add_front(next_instance)
						}
					}
				}
			}
		}
	}
}

func (rq *RunQueue) update_queue() int {
	// Try to start queued instances
	running := 0
	for i := range rq.Active {
		if i.ProcessingState == 0 || i.ProcessingState == 3 {
			running++
		}
	}
	for rq.Next < len(rq.Queue) && running < rq.MaxRunning {
		instance := rq.Queue[rq.Next]
		rq.Next++

		if instance.ProcessingState < 0 {
			instance.ProcessingState = 0 // indicate started/running
			rq.Active[instance] = struct{}{}
			running++
		} else {
			if instance.ProcessingState != 3 {
				panic("Bug")
			}
			// Cancelled before starting, skip it
			continue
		}

		ttdata := instance.TtData

		// If eligible for binary splitting, queue the two halves
		if len(instance.Constraints) > 1 {
			nhalf := len(instance.Constraints) / 2
			instance.Instance1 = new_instance(
				instance.BaseInstance,
				ttdata.Description+"~0",
				instance.ConstraintType,
				instance.Constraints[:nhalf],
				instance.Timeout)
			rq.add(instance.Instance1)

			instance.Instance2 = new_instance(
				instance.BaseInstance,
				ttdata.Description+"~1",
				instance.ConstraintType,
				instance.Constraints[nhalf:],
				instance.Timeout)
			rq.add(instance.Instance2)
		}

		if instance.Tagged {
			base.Message.Printf("(TODO) [%d] >> %s\n",
				Ticks, ttdata.Description)
		} else {
			base.Message.Printf("(TODO) [%d] Start %s\n",
				Ticks, ttdata.Description)
		}
		timetable.BACKEND.Run(ttdata)
	}
	//TODO--
	fmt.Printf("$ [%d] Running/Active instances: %d/%d\n",
		Ticks, running, len(rq.Active))
	base.Message.Printf("$ [%d] Running/Active instances: %d/%d\n",
		Ticks, running, len(rq.Active))
	return running
}

func (rq *RunQueue) disable() {
	rq.MaxRunning = 0 // no new starts possible
}

func (rq *RunQueue) instance_completed(instance *TtInstance, state int) {
	instance.ProcessingState = state
	if state != 2 {
		rq.instance_deactivate(instance)
	}
}

func (rq *RunQueue) instance_deactivate(instance *TtInstance) {
	delete(rq.Active, instance)
}

// Cancelling an instance will abort it if it is running.
// The `ProcessingState` is set to 3 to indicate that a queued instance
// is not to be started. A cancelled instance is not needed, so no
// result is expected and its data can be removed.
func cancel_instance(instance *TtInstance) {
	if instance != nil {
		if instance.ProcessingState == 0 {
			abort_instance(instance)
		}
		instance.ProcessingState = 3
		// Cancel subsidiary instances
		cancel_instance(instance.Instance1)
		cancel_instance(instance.Instance2)
	}
}

func abort_instance(instance *TtInstance) {
	if !instance.Stopped {
		timetable.BACKEND.Abort(instance.TtData)
		instance.Stopped = true
	}
}

func new_instance(
	instance_0 *TtInstance,
	descriptor string,
	constraint_type timetable.ConstraintType,
	constraint_indexes []int,
	timeout int,
) *TtInstance {
	// Copy original TtData (shallow copy only!)
	ttdata := new_ttdata(instance_0.TtData, descriptor)

	// Make a deep copy of the hard constraint matrix
	hcmat0 := instance_0.HardConstraintEnabled
	hcmat := make([][]bool, len(hcmat0))
	for i, c := range hcmat0 {
		hcmat[i] = slices.Clone(c)
	}

	// Make a new `TtInstance`
	instance := &TtInstance{
		Timeout:               timeout,
		TtData:                ttdata,
		HardConstraintEnabled: hcmat,

		// Base data for this instance:
		BaseInstance:   instance_0,
		ConstraintType: constraint_type,
		Constraints:    constraint_indexes,

		// Run time
		Stopped:   false,
		Instance1: nil,
		Instance2: nil,
		Result:    nil,
	}

	// Enable the constraints in `ttdata` and `hcmat`
	//fmt.Printf("§ENABLE %s: %v\n", constraint_type.String(), constraint_indexes)

	// Mark the constraints in the matrix
	cmap := instance.HardConstraintEnabled[constraint_type]
	for _, i := range constraint_indexes {
		cmap[i] = true
	}
	// Reconstruct the constraint list
	newlist := []any{}
	for i, c := range TtData_0.HardConstraints[constraint_type] {
		if cmap[i] {
			newlist = append(newlist, c)
		}
	}
	instance.TtData.HardConstraints[constraint_type] = newlist

	return instance
}

func new_ttdata(
	ttdata_0 *timetable.TtData,
	descriptor string,
) *timetable.TtData {
	ttdata := &timetable.TtData{
		Description:             descriptor,
		SharedData:              ttdata_0.SharedData,
		TeacherNotAvailable:     ttdata_0.TeacherNotAvailable,
		ClassNotAvailable:       ttdata_0.ClassNotAvailable,
		RoomNotAvailable:        ttdata_0.RoomNotAvailable,
		WITHOUT_ROOM_PLACEMENTS: ttdata_0.WITHOUT_ROOM_PLACEMENTS,
		State:                   0,
	}

	// Make a deeper copy of the constraints so that these can be
	// switched on or off without affecting those in the original `TtData`.

	// Make a copy of the constraints lists
	hcmap := make(map[timetable.ConstraintType][]any,
		len(ttdata_0.HardConstraints))
	for k, v := range ttdata_0.HardConstraints {
		hcmap[k] = slices.Clone(v)
	}
	ttdata.HardConstraints = hcmap

	scmap := make(map[timetable.ConstraintType][]any,
		len(ttdata_0.SoftConstraints))
	for k, v := range ttdata_0.SoftConstraints {
		scmap[k] = slices.Clone(v)
	}
	ttdata.SoftConstraints = scmap
	return ttdata
}

/*
// This will probably need some tuning. There is probably no such thing as
// an optimal algorithm, that depends very much on the data.
func timed_out(instance *TtInstance) bool {
	if instance.Timeout < 0 {
		// This allows a complete override of the timeout feature.
		// Unlike Timeout = 0 it will not try to catch runs which get
		// stuck at low completion levels.
		return false
	}
	if instance.Timeout != 0 && instance.Ticks > instance.Timeout {
		return true
	}
	delta := instance.Ticks - instance.LastTime
	if delta < 5 {
		return false
	}
	// The acceptable time should depend on Progress.
	if instance.Progress < 80 && delta*2 > instance.Progress {
		return true
	}
	if instance.Progress < 95 && delta > 400 {
		return true
	}
	return false
}
*/

//
/*
// TODO:  Is this old bit fetching file names from db.ModuleData still
// needed somehow?

	fetfile := stempath
	mapfile := stempath
	thisdir := filepath.Dir(stempath)
	moduleData := db.ModuleData
	fetData, ok := moduleData["FetData"].(map[string]string)
	if ok {
		var f string
		f, ok = fetData["FetFile"]
		if ok {
			fetfile = filepath.Join(thisdir, f)
		}
		f, ok = fetData["MapFile"]
		if ok {
			mapfile = filepath.Join(thisdir, f)
		}
	}
	fetfile += ".fet"
	mapfile += ".map"
*/
