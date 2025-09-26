package autotimetable

import (
	"W365toFET/timetable"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"sync"
	"syscall"
	"time"
)

// TODO: How to set this up?
var NULL_TIMEOUT int = 10 // timeout ticks for unconstrained trial
var TIMEOUT_1 int = 10    // ticks for single constraint type test functions
var TIMEOUT_2 int = 10    // ticks for added constraint type test functions
//TODO: TIMEOUT_2 is used for adding constraints to an already somewhat
// constraint data set. Should later additions get more time?

//TODO: Consider starting instances with each (used) constraint type, individually.
// The first one which completes successfully could be taken as a new basis, to
// which the next completed ones could be added sequentially ... until no time is left?

// TODO: It may well be desirable to be able to override this – see also GOMAXPROCS
var MAXPROCESSES int = runtime.NumCPU() //TODO: use this?

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

// Function to generate timetable from the given data
var TtGenerate func(*timetable.TtData, *sync.WaitGroup)

func StartGeneration(tt_data_0 *timetable.TtData, TIMEOUT int) {
	tt_shared_data := tt_data_0.SharedData

	// Catch termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait group to ensure all goroutines finish before exiting
	var wg sync.WaitGroup

	// `workingdir` provides the path to a working directory which can be used
	// freely during processing. It may or may not already exist, existing
	// contents need not be preserved during processing.

	//TODO? stop still needed?
	// Open communication channels
	//stop := make(chan bool)

	//TODO: Consider buffer size and blocking ...
	//add_instance := make(chan *TtInstance, 10)
	//instance_done := make(chan *TtInstance, 10)

	/*
		global_data := GlobalData{
			Ticks:    0,
			TtData_0: tt_data_0,
			//Instances: []*TtInstance{},
			//NewInstance: add_instance,
		}
	*/

	tt_data_0.Description = "COMPLETE"
	workingdir := tt_shared_data.WorkingDir

	// Provide an empty working directory.
	os.RemoveAll(workingdir)
	err := os.Mkdir(workingdir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	// First run: all constraints enabled
	//TODO: On successful completion, all other instances should be stopped.
	// If it fails, just this instance should be wound up. Otherwise it
	// should still be running when the whole process finishes, and would
	// need stopping.
	full_instance := &TtInstance{
		//Id:     -1,
		TtData_0: tt_data_0,
		Delay:    -1,
		//Ticks:       0,
		//Timeout: TIMEOUT,

		TtData: tt_data_0,

		//Stop:        stop,
		//WaitGroup: &wg,

		//HardConstraintEnabled: setup_hard_constraint_map(
		//	tt_data_0.HardConstraints),

		//State:    0,
		//Progress: 0,
		//LastTime: 0,

		//TODO: Maybe not using these any more ...
		//FailurePath: TtChainedFunc{
		//	Delay: 1, Func: test_sequence},

		//SuccessPath: TtChainedFunc{
		//	Delay: 0, Func: full_success},
	}

	//for _, c := range tt_data_0.Db.Classes {
	//	fmt.Printf("??? %+v\n", c)
	//}

	//TODO ...

	// Request start of full instance
	//start_constraint_trial(instance)

	// Start run
	TtGenerate(tt_data_0, &wg)

	// Unconstrained instance
	null_instance := &TtInstance{
		//Id:     -1,
		//Global: &global_data,
		TtData_0: tt_data_0,
		Delay:    -1,
		//Ticks:       0,
		//Timeout: TIMEOUT,

		TtData: new_ttdata(tt_data_0, "ONLY_BLOCKED_SLOTS"),
		HardConstraintEnabled: setup_hard_constraint_map(
			tt_data_0.HardConstraints),
	}
	disable_all_constraints(null_instance.TtData)
	// Start run
	TtGenerate(null_instance.TtData, &wg)

	// Request start
	//start_constraint_trial(inst)

	/*
		// Request start of instances with individually enabled constraint
		// types (in goroutine to avoid blocking main goroutine here)
		wg.Add(1)
		go func() {
			defer wg.Done()
			start_constraints(inst, instance_done)
		}()
	*/

	// *** Ticker loop ***
	ticks := 0
	null_timeout := NULL_TIMEOUT
	stage := 0
	next_step := 0
	var basic_constraints []*TtInstance
	var current_instance *TtInstance
	steps := []*TtInstance{}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

tickloop:
	for {
		select {

		case ossig := <-sigChan:
			//TODO
			fmt.Printf("*** SIGNAL *** %+v", ossig)
			return

		case <-ticker.C:
			ticks++

			//TODO?
			TIMEOUT--
			if TIMEOUT == 0 {
				// Cancel all runs and return the "best" instance so far.

				//TODO ...

			}

			if full_instance.TtData.State == 1 {
				// Cancel all other runs and return this as the result.

				//TODO ...

				current_instance = full_instance
				break tickloop
			}

			if stage == 0 {
				if null_timeout >= 0 {
					if null_instance.TtData.State == 1 {
						current_instance = null_instance

						//TODO: Start trials ...
						basic_constraints = start_basic_constraints(
							null_instance, &wg)
						stage = 1
						continue
					}
					null_timeout--
					if null_timeout < 0 {
						// The null instance took "too long".
						stage = -1

						//TODO: Seek problems in the unconstrained data.
					}
				}
				continue
			}

			if stage == 1 {
				// Awaiting completion of the single-constraint instances,
				// check their status
				newlist := []*TtInstance{}
				for _, bc := range basic_constraints {
					//TODO? This assumes the basic constraints are handled as
					// full steps, always returning a usable result.
					if bc.Result == nil {
						newlist = append(newlist, bc)
					} else {
						steps = append(steps, bc.Result)

						//TODO: At this stage, it might be possible to set
						// `current_instance` and add the other constraints
						// step by step ...

					}
				}
				basic_constraints = newlist
				if len(basic_constraints) == 0 {
					stage = 2 // all basic constraint tests completed
				}
				if len(steps) != 0 {
					if next_step == 0 {
						// Start adding constraint types
						current_instance = steps[0]
						next_step = 1
					}
				}
			}

			if current_instance.Result != nil {
				if next_step < len(steps) {
					// Add next constraint type
					st1 := steps[next_step]
					desc := current_instance.TtData.Description + "*" +
						st1.TtData.Description
					current_instance = new_instance(
						current_instance,
						desc,
						st1.ConstraintType,
						st1.Constraints,
						TIMEOUT_2)
					next_step++
				} else if stage == 2 {
					// No more constraint types => finished ...
					if full_instance.TtData.State == 0 {
						// Cancel full_instance
						full_instance.TtData.SharedData.Abort(full_instance.TtData)
					}
					break tickloop
				}
			}
		}
	}

	//TODO: Consider also the possibility that there may be no (or only one)
	// basic constraint types.

	/*
	 * Terminating an instance can lead to various completion state values.
	 * If the state is set in its own goroutine, the value can be -1 if
	 * the run failed for some reason within the data, or 5 if the run
	 * was terminated externally. However, if the termination is due to a
	 * timeout, the value will be set to -3 in the tick-loop. Any instance
	 * terminated by a `cancelAll` call will have its state set to -4.
	 */

	active_instances := []*TtInstance{}
	inactive_instances := []*TtInstance{}
	ended := []*TtInstance{}

	//TODO
	//var result *TtInstance

loop:
	for {
		select {

		case ossig := <-sigChan:
			//TODO
			fmt.Printf("*** SIGNAL *** %+v", ossig)
			return

		//case <-stop:
		//fmt.Println("checkProgress done!")
		//TODO: tidying up? This channel is currently unused!
		//return

		case new_instance := <-add_instance:
			//TODO--?
			// Queue start of generator back-end

			fmt.Printf("+++ Add: %s @ %d\n",
				new_instance.TtData.Description, global_data.Ticks)

			active_instances = append(active_instances, new_instance)

		case <-ticker.C:
			if len(active_instances) == 0 && len(inactive_instances) != 0 {
				break loop
			}
			global_data.Ticks++

			// Update the progress records of the currently active
			// subprocesses, handle tick-related events.
			for _, inst := range active_instances {
				// Once the instance goroutine has finished, `tt_data.State`
				// is > 0.
				// If the state is 1, the instance completed successfully.
				// If the state is 2, the instance failed (somehow the data
				// was discovered to be insoluble).
				// If the state is 3, the run was stopped.

				//TODO--
				// There are two kinds of "external" termination:
				//  - a timeout, which works like a pre-empted failure, and
				//    is indicated by state 3;
				//  - a cancellation, which can be used to terminate a
				//    an instance which (with hindsight) should not have been
				//    started in the first place. This is indicated by
				//    `inst.Cancelled` being `true`.

				if inst.Delay >= 0 {
					inst.Delay--
					if inst.Delay < 0 {
						// Start generator back-end

						fmt.Printf("*** Start: %s @ %d\n",
							inst.TtData.Description, global_data.Ticks)

						TtGenerate(inst.TtData, &wg)
					}
					continue
				}

				//TODO: timeout – consider interaction with delay, etc.
				if inst.Timeout >= 0 {
					inst.Timeout--
					if inst.Timeout == 0 {
						inst.Termination = 1
						tt_shared_data.Abort(inst.TtData)
					}
				}

				tt_data := inst.TtData
				if tt_data.State > 0 {
					// The goroutine has finished – or was cancelled,
					// mark the instance for removal from the active list.
					ended = append(ended, inst)

					//TODO ...

					// If appropriate, activate follow-on processes.
					if tt_data.State == 1 {
						// succeeded ...
						/*cancelPath(inst.FailureInstance)
						inst.FailureInstance = nil
						if inst.SuccessPath.Delay >= 0 {
							if inst.SuccessPath.Func != nil {
								inst.SuccessInstance = inst.SuccessPath.Func(inst)
							}
							inst.SuccessPath.Delay = -1
						}
						*/
					} else if tt_data.State != 5 {
						// failed ...
						/*cancelPath(inst.SuccessInstance)
						inst.SuccessInstance = nil
						if inst.FailurePath.Delay >= 0 {
							if inst.FailurePath.Func != nil {
								inst.FailureInstance = inst.FailurePath.Func(inst)
							}
							inst.FailurePath.Delay = -1
						}
						*/
					}
					continue
				}
				tt_data.Ticks++
				h := tt_shared_data.TickHandler
				if h != nil {
					// The handler should only be set when the the process is
					// fully running
					h(tt_data)
					if tt_data.State != 0 {
						continue
					}
				}

				/*TODO?
				// Handle timeout
				if timed_out(inst) {
					inst.State = -1
					//TODO: using an interface?
					// inst.Abort(inst.HandlerData)
					continue
				}
				*/

				/*
					// Handle starting of follow-on paths after their delays
					if inst.SuccessPath.Delay == inst.Ticks {
						if inst.SuccessPath.Func != nil {
							inst.SuccessInstance = inst.SuccessPath.Func(inst)
						}
						inst.SuccessPath.Delay = -1 // flag already started
					}
					if inst.FailurePath.Delay == inst.Ticks {
						if inst.FailurePath.Func != nil {
							inst.FailureInstance = inst.FailurePath.Func(inst)
						}
						inst.FailurePath.Delay = -1 // flag already started
					}
					for _, chfunc := range inst.OtherPaths {
						if chfunc.Delay == inst.Ticks {
							inst.OtherInstances = append(inst.OtherInstances,
								chfunc.Func(inst))
							//chfunc.Delay = -1
							//inst.OtherPaths[i] = chfunc
						}
					}
				*/
			}
			// Remove terminated instances from the active list
			if len(ended) != 0 {
				for _, inst := range ended {
					tt_data := inst.TtData
					if tt_data.State != 5 {
						inactive_instances = append(inactive_instances, inst)
					}

					fmt.Printf("--- End: %s cc=%d (%d) @ %d\n",
						inst.TtData.Description, tt_data.State,
						tt_data.Progress, global_data.Ticks)

					if inst.Id >= 0 {
						instance_done <- inst
					} else {
						if inst.Id <= ID_BLOCKED_SLOTS {
							if tt_data.State != 1 {

								//TODO: cancel running instances,
								// start testing individual classes

							}
						} else {
							// A "complete" instance has succeeded.

							//TODO: cancel running instances, wind everything up.

							//TODO: result = inst

						}
					}

				}
				fmt.Printf("§§§ Before: %d %d\n", len(active_instances), len(ended))
				active_instances = slices.DeleteFunc(active_instances,
					func(tti *TtInstance) bool {
						return slices.Contains(ended, tti)
					})
				fmt.Printf("§§§ After: %d\n", len(active_instances))
				ended = ended[:0]

				//

			}
		}
	}
	instance_done <- nil
	wg.Wait()
	// All instances have completed.
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
		Termination: 0,
		TtData:      ttdata,
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
