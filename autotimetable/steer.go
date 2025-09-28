package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"
)

// TODO: How to set this up?
var UNCONSTRAINED_TIMEOUT int = 10 // timeout ticks for unconstrained trial
var DELAY_BINARY_CHOP int = 3

var TIMEOUT_1 int = 10 // ticks for single constraint type test functions
var TIMEOUT_2 int = 10 // ticks for added constraint type test functions
//TODO: TIMEOUT_2 is used for adding constraints to an already somewhat
// constraint data set. Should later additions get more time?

//TODO: Consider starting instances with each (used) constraint type, individually.
// The first one which completes successfully could be taken as a new basis, to
// which the next completed ones could be added sequentially ... until no time is left?

// TODO: It may well be desirable to be able to override this – see also GOMAXPROCS
// var MAXPROCESSES int = runtime.NumCPU() //TODO: use this?
var MAXPROCESSES int = 100

// TODO: At present this only supports a FET back-end. Perhaps a choice should
// be possible ...

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

const ID_BLOCKED_SLOTS int = -100

func StartGeneration(tt_data_0 *timetable.TtData, TIMEOUT int) {
	tt_shared_data := tt_data_0.SharedData

	// Catch termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//TODO-- Wait group to ensure all goroutines finish before exiting
	//var wg sync.WaitGroup

	runqueue := RunQueue{
		Queue:      nil,
		Running:    map[*TtInstance]struct{}{},
		MaxRunning: MAXPROCESSES, //TODO????
		Next:       0,
	}

	// `workingdir` provides the path to a working directory which can be used
	// freely during processing. It may or may not already exist, existing
	// contents need not be preserved during processing.

	//TODO? stop still needed?
	// Open communication channels
	//stop := make(chan bool)

	//TODO: Consider buffer size and blocking ...
	//add_instance := make(chan *TtInstance, 10)
	//instance_done := make(chan *TtInstance, 10)

	tt_data_0.Description = "COMPLETE"
	tt_data_0.State = -1 // not started yet
	workingdir := tt_shared_data.WorkingDir

	// Provide an empty working directory.
	os.RemoveAll(workingdir)
	err := os.Mkdir(workingdir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	global_data := &GlobalData{
		Ticks:    0,
		TtData_0: tt_data_0,
	}

	// First run: all constraints enabled.
	//TODO: On successful completion, all other instances should be stopped.
	// If it fails, just this instance should be wound up. Otherwise it
	// should run until it times out, at which point any other active
	// instances should be stopped and the "best" solution at this point
	// chosen.
	full_instance := &TtInstance{
		Global: global_data,
		Delay:  TIMEOUT,
		TtData: tt_data_0,
	}

	//for _, c := range tt_data_0.Db.Classes {
	//	fmt.Printf("??? %+v\n", c)
	//}

	//TODO ...

	// Add to run queue (and start running)
	runqueue.Add(full_instance)

	// Unconstrained instance
	null_instance := &TtInstance{
		Global: global_data,
		Delay:  UNCONSTRAINED_TIMEOUT,

		TtData: new_ttdata(tt_data_0, "ONLY_BLOCKED_SLOTS"),
		HardConstraintEnabled: setup_hard_constraint_map(
			tt_data_0.HardConstraints),
	}
	disable_all_constraints(null_instance.TtData)
	// Add to run queue (and start running)
	runqueue.Add(null_instance)

	// *** Ticker loop ***
	stage := 0
	next_step := 0
	var basic_constraints map[*TtInstance]struct{}
	var current_instance *TtInstance
	steps := []*TtInstance{}
	ticker := time.NewTicker(time.Second)
	runqueue.Update()
	defer ticker.Stop()

tickloop:
	for {
		select {

		case ossig := <-sigChan:
			//TODO
			fmt.Printf("*** SIGNAL *** %+v", ossig)
			return

		case <-ticker.C:
			global_data.Ticks++

			if full_instance.TtData.State == 1 {
				// Cancel all other runs and return this as the result.
				runqueue.Disable()
				stop_instance(null_instance)
				stop_instance(current_instance)
				current_instance = full_instance
				break tickloop
			} else if full_instance.TtData.State == 0 {
				if full_instance.Delay >= 0 {
					if full_instance.Delay == 0 {
						runqueue.Disable()
						stop_instance(full_instance)
						stop_instance(null_instance)
						if current_instance.Result == nil {
							stop_instance(current_instance)

							//TODO
							panic("TODO: Seek best result so far")

						}
						break tickloop
					}
					full_instance.Delay--
				}
				timetable.BACKEND.Tick(full_instance.TtData)
			}

			//TODO: Special treatment if there are no constraints to add?

			//TODO: The time taken to complete the unconstrained instance
			// should be taken into account when setting up delay/timeout
			// for the constrained versions.

			if stage == 0 {
				// During stage 0 only `full_instance` and `null_instance`
				// are running.
				if null_instance.TtData.State != 0 {
					if null_instance.TtData.State == 1 {
						// The null instance completed successfully.
						current_instance = null_instance
						// Start trials
						basic_constraints = start_basic_constraints(
							null_instance, &runqueue)
						stage = 1
						runqueue.Update()
					} else {
						// The null instance failed.
						stage = -1
						base.Message.Println("(TODO) Unconstrained instance failed")

						//TODO: Seek problems in the unconstrained data.

					}
				} else {
					if null_instance.Delay == 0 {
						stop_instance(null_instance)
					} else {
						timetable.BACKEND.Tick(null_instance.TtData)
						if null_instance.Delay > 0 {
							null_instance.Delay--
						}
					}
					goto tickloop_end
				}
			}

			if stage == 1 {
				// During stage 1 we are awaiting completion of the single-
				// constraint instances: check their status.
				for bc := range basic_constraints {
					//TODO? This assumes the basic constraints are handled as
					// full steps, always returning a usable result.
					tick_instance(&runqueue, bc)
					if bc.Result != nil {
						steps = append(steps, bc.Result)
						delete(basic_constraints, bc)
					}
				}
				if len(basic_constraints) == 0 {
					stage = 2 // all basic-constraint trials completed
				}
				if next_step == 0 && len(steps) != 0 {
					// Start adding constraint types
					current_instance = steps[0]
					base.Message.Printf("(TODO) First constraint: %s\n",
						current_instance.TtData.Description)
					next_step = 1
				}

				fmt.Printf("STAGE1: %d %d %d\n", stage, next_step, len(steps))
			}

			// Handle the active instances (recursively)
			if current_instance != nil {
				if current_instance.Result == nil {
					tick_instance(&runqueue, current_instance)
				} else {
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
							TIMEOUT_2)
						next_step++
						runqueue.Add(current_instance)
					} else if stage == 2 {
						// No more constraint types => finished ...
						// Cancel full_instance
						//TODO: but only if timeout reached?
						stop_instance(full_instance)
						break tickloop
					}
				}
			}
		}
	tickloop_end:
		runqueue.Update()
	} // tickloop: end

	//TODO: Consider also the possibility that there may be no (or only one)
	// basic constraint types.

	// The result is in `current_instance`.
	//TODO
	fmt.Printf("RESULT: %s\n", current_instance.TtData.Description)
}

type RunQueue struct {
	Queue      []*TtInstance
	Running    map[*TtInstance]struct{}
	MaxRunning int
	Next       int
}

func (rq *RunQueue) Add(instance *TtInstance) {
	rq.Queue = append(rq.Queue, instance)
	base.Message.Printf("(TODO) [%d] Queue %s\n",
		instance.Global.Ticks, instance.TtData.Description)
}

func (rq *RunQueue) Update() {
	for instance := range rq.Running {
		if instance.TtData.State != 0 {
			base.Message.Printf("(TODO) [%d] Done %s\n",
				instance.Global.Ticks, instance.TtData.Description)
			delete(rq.Running, instance)
		}
	}
	for rq.Next < len(rq.Queue) && len(rq.Running) < rq.MaxRunning {
		instance := rq.Queue[rq.Next]
		rq.Next++
		instance.TtData.State = 0 // indicate started
		rq.Running[instance] = struct{}{}
		base.Message.Printf("(TODO) [%d] Start %s\n",
			instance.Global.Ticks, instance.TtData.Description)
		timetable.BACKEND.Run(instance.TtData)
	}
	//TODO--
	fmt.Printf("$Running instances: %d\n", len(rq.Running))
}

func (rq *RunQueue) Disable() {
	rq.MaxRunning = 0 // no new starts possible
}

func tick_instance(runqueue *RunQueue, instance *TtInstance) {
	ttdata := instance.TtData
	if ttdata.State == 0 {
		// Still running
		timetable.BACKEND.Tick(ttdata)
		/*TODO: Test whether "stuck".
		if ttdata.State == 0 {
			step := ttdata.Progress - instance.LastProgress
			if step != 0 {
				instance.LastStep = step
				steptime := instance.Global.Ticks - instance.LastChange
				instance.LastChange = instance.Global.Ticks
			} else {
				step = instance.LastStep
			}

			remaining := 100 - ttdata.Progress
		}
		*/
	}
	if ttdata.State == 1 {
		// Completed successfully
		instance.Result = instance
		// Stop subsidiary instances
		stop_instance(instance.Instance0)
		stop_instance(instance.Instance1)
		return
	}
	if instance.Delay >= 0 {
		if instance.Delay == 0 {
			// Start subsidiary activities, if any
			if len(instance.Constraints) == 1 {
				// No subsidiary activities: quit, returning base instance
				stop_instance(instance)
				instance.Result = instance.BaseInstance
			} else {
				// Divide constraints into two lists
				half := len(instance.Constraints) / 2
				i0 := new_instance(
					instance.BaseInstance,
					instance.TtData.Description+"~0",
					instance.ConstraintType,
					instance.Constraints[:half],
					DELAY_BINARY_CHOP,
				)
				i1 := new_instance(
					instance.BaseInstance,
					instance.TtData.Description+"~1",
					instance.ConstraintType,
					instance.Constraints[half:],
					DELAY_BINARY_CHOP,
				)
				instance.Instance0 = i0
				instance.Instance1 = i1
				runqueue.Add(i0)
				runqueue.Add(i1)
			}
		} else {
		}
		instance.Delay--

	} else {
		// If there are subsidiary instances, they have already started.

		//TODO?
		if instance.Instance0 == nil {
			// No subsidiary instances
			return
		}
		tick_instance(runqueue, instance.Instance0)
		if instance.Instance1 == nil {
			// Running the combined instance
			if instance.Instance0.Result != nil {
				// The combined instance is finished, take this as result
				// and stop the primary instance.

				//TODO
				if ttdata.State == 0 {
					stop_instance(instance)
					instance.Result = instance.Instance0.Result
				}
			}

		} else {
			tick_instance(runqueue, instance.Instance1)
			i0 := instance.Instance0.Result
			i1 := instance.Instance1.Result
			if i0 != nil {
				if i1 != nil {
					//Combine the results
					if i0 == instance.BaseInstance {
						instance.Result = i1
					} else if i1 == instance.BaseInstance {
						instance.Result = i0
					} else {
						// Use `i0` as base and try to add the constraints from `i1`.

						//TODO
						// Get the constraints from `i1`.
						i1clist := []int{}
						for c, ok := range i1.HardConstraintEnabled[i1.ConstraintType] {
							if ok {
								i1clist = append(i1clist, c)
							}
						}
						i2 := new_instance(
							i0,
							instance.TtData.Description+"~2",
							instance.ConstraintType,
							i1clist,
							DELAY_BINARY_CHOP,
						)
						instance.Instance0 = i2
						instance.Instance1 = nil
						runqueue.Add(i2)
					}
				}
			}
		}
	}
}

func stop_instance(instance *TtInstance) {
	if instance == nil || instance.Result != nil {
		return
	}
	if instance.TtData.State == 0 {
		timetable.BACKEND.Abort(instance.TtData)
	}
	// Stop subsidiary instances
	stop_instance(instance.Instance0)
	stop_instance(instance.Instance1)
}

func new_instance(
	instance_0 *TtInstance,
	descriptor string,
	constraint_type timetable.ConstraintType,
	constraint_indexes []int,
	division_delay int,
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
		//Id: ???,
		Global: instance_0.Global,
		Delay:  division_delay,
		//Timeout: ???,
		//Termination: 0,
		TtData: ttdata,
		//? WaitGroup:             instance_0.WaitGroup,
		HardConstraintEnabled: hcmat,

		// Base data for this instance:
		BaseInstance:   instance_0,
		ConstraintType: constraint_type,
		Constraints:    constraint_indexes,

		// Run time
		Instance0: nil,
		Instance1: nil,
	}

	// Enable the constraints in `ttdata` and `hcmat`
	enable_hard_constraints(instance, constraint_type, constraint_indexes)

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
		State:                   -1, // not started
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

/*TODO? Stop an instance and all of its children. Not only the "success" branch is
// to be stopped, but also all the others.
func cancelPath(instance *TtInstance) {
	if instance == nil {
		return
	}
	if instance.State == 0 {
		//instance.Abort(instance.HandlerData)
		instance.State = 5
	}
	cancelPath(instance.FailureInstance)
	instance.FailureInstance = nil
	cancelPath(instance.SuccessInstance)
	instance.SuccessInstance = nil
	for i, inst := range instance.OtherInstances {
		cancelPath(inst)
		instance.OtherInstances[i] = nil
	}
}
*/

/*TODO-- If the run with all constraints enabled succeeds, there is probably
// no need for further diagnosis, so cancel all other instances.
func full_success(instance_0 *TtInstance) *TtInstance {
	cancelPath(instance_0.FailureInstance)
	instance_0.FailureInstance = nil
	return nil
}
*/

/*
func newInstance(
	instance_0 *TtInstance, descriptor string, tag int,
) *TtInstance {
	tt_data_0 := instance_0.TtData
	// Copy original TtData (shallow copy only!)
	tt_data := timetable.TtData{
		Description:             descriptor,
		SharedData:              tt_data_0.SharedData,
		TeacherNotAvailable:     tt_data_0.TeacherNotAvailable,
		ClassNotAvailable:       tt_data_0.ClassNotAvailable,
		RoomNotAvailable:        tt_data_0.RoomNotAvailable,
		WITHOUT_ROOM_PLACEMENTS: tt_data_0.WITHOUT_ROOM_PLACEMENTS,
	}

	// Make a deeper copy of the constraints so that these can be
	// switched on or off without affecting those in the original `TtData`
	// and `DbTopLevel`.

	// Make a copy of the constraints lists
	hcmap := make(map[timetable.ConstraintType][]any,
		len(tt_data_0.HardConstraints))
	for k, v := range tt_data_0.HardConstraints {
		hcmap[k] = slices.Clone(v)
	}
	tt_data.HardConstraints = hcmap

	scmap := make(map[timetable.ConstraintType][]any,
		len(tt_data_0.SoftConstraints))
	for k, v := range tt_data_0.SoftConstraints {
		scmap[k] = slices.Clone(v)
	}
	tt_data.SoftConstraints = scmap

	tt_data.Description = descriptor

	// Make a new `TtInstance`
	instance := *instance_0
	instance.TtData = &tt_data
	instance.Id = tag
	instance.Timeout = TIMEOUT_1 // default timeout ticks
	// Make a deep copy of the hard constraint matrix
	hcmat := instance_0.HardConstraintEnabled
	instance.HardConstraintEnabled = make([][]bool, len(hcmat))
	for i, c := range hcmat {
		instance.HardConstraintEnabled[i] = slices.Clone(c)
	}
	return &instance
}
*/

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
