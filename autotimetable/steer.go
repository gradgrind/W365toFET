package autotimetable

import (
	"W365toFET/base"
	"W365toFET/fet"
	"W365toFET/timetable"
	"os"
	"runtime"
	"time"
)

// TODO: It may well be desirable to be able to override this
var MAXPROCESSES int = runtime.NumCPU() //TODO: use this?

/*
A `TtRunData` structure is constructed to manage the data being used by the
timetable engine. Each trial run has a `TtInstance` structure to manage the
data specific to this instance.

A goroutine is used to run `checkProgress`, which monitors the state of all
active runs by querying them every second, updating their `TtInstance`
structures when progress is made. It is also possible to register timed
handlers, e.g. a timeout.

The `TtInstance` structures have an `Abort` function, allowing a run to be
terminated before it stops naturally.

Each trial run should have a success path and a failure path. One of these
paths can be started preemptively, perhaps after a certain delay, if
processor units are available. When a trial is resolved, it should be able
to cancel any of the trials on the now invalidated path. Cancellation of a
trial would need to specify which of the paths is to be taken (or none!).

Would it be sensible to limit preemptive spawning of new instances, either by
limiting the total number of active Instances allowed or by setting a minimum
path start delay? The latter alone might not be enough, though.

Perhaps the main goroutine, the one which starts the ball rolling, could
act as a sort of general controller, receiving signals on a channel to abort
runs, determine when the algorithm has finished, etc.
*/

var Descriptions map[string]string = map[string]string{
	"COMPLETE":           "All constraints active",
	"ONLY_BLOCKED_SLOTS": "All constraints – except blocked slots – disabled",
}

// TODO: At present this only supports a FET back-end. Perhaps a choice should
// be possible ...

func SteerGeneration(tt_data_0 *timetable.TtData, workingdir string) {

	// `workingdir` provides the path to a working directory which can be used
	// freely during processing. It may or may not already exist, existing
	// contents need not be preserved during processing.

	// Open communication channels
	stop := make(chan bool)
	make_instance := make(chan *timetable.TtInstance)

	{
		// Provide an empty working directory.
		os.RemoveAll(workingdir)
		err := os.Mkdir(workingdir, 0755)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}

		// First run: all constraints enabled, no timeout
		//TODO: On successful completion, all other instances should be stopped.
		// If it fails, just this instance should be wound up. Otherwise it
		// should still be running when the whole process finishes, and would
		// need stopping.
		instance := &timetable.TtInstance{
			//Id:          0,
			Description: "COMPLETE",
			Ticks:       0,
			WorkingDir:  workingdir,
			TtData_0:    tt_data_0,
			TtData:      tt_data_0,
			NewInstance: make_instance,
			Stop:        stop,

			FailurePath: timetable.TtChainedFunc{
				Delay: 1, Func: test_sequence},
		}

		instance.FailurePath = timetable.TtChainedFunc{
			Delay: 1, Func: test_sequence}

		// Request start of instance
		make_instance <- instance
	}

	// *** Channel reader loop ***

	active_instances := map[*timetable.TtInstance]struct{}{}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			//fmt.Println("checkProgress done!")
			//TODO: tidying up?
			return
		case new_instance := <-make_instance:
			//TODO: The timetable "engine" should be replaceable.
			fet.NewFet(new_instance)
			active_instances[new_instance] = struct{}{}
		case <-ticker.C:
			// Update the progress records of the currently active
			// subprocesses, handle tick-related events.
			for inst := range active_instances {
				inst.Ticks++
				h := inst.UpdateHandler
				if h != nil {
					// The handler should only be set when the the process is
					// fully running
					h(inst)
				}

				// Handle starting of follow-on paths
				if inst.SuccessPath.Delay >= inst.Ticks {
					inst.SuccessInstance = inst.SuccessPath.Func(inst)
					inst.SuccessPath.Delay = -1
				}
				if inst.FailurePath.Delay >= inst.Ticks {
					inst.FailureInstance = inst.FailurePath.Func(inst)
					inst.FailurePath.Delay = -1
				}
				for i, chfunc := range inst.OtherPaths {
					if chfunc.Delay >= inst.Ticks {
						inst.OtherInstances = append(inst.OtherInstances,
							chfunc.Func(inst))
						chfunc.Delay = -1
						inst.OtherPaths[i] = chfunc
					}
				}
			}
		}
	}
}

func test_sequence(instance_0 *timetable.TtInstance) *timetable.TtInstance {
	// Copy original DbTopLevel (shallow copy only!)
	db0 := instance_0.TtData_0.Db
	db := *db0

	// Copy original TtData (shallow copy only!)
	tt_data := *instance_0.TtData_0
	tt_data.Db = &db

	// Make a new `TtInstance`
	instance := &timetable.TtInstance{
		Description: "ONLY_BLOCKED_SLOTS",
		Ticks:       0,
		WorkingDir:  instance_0.WorkingDir,
		TtData_0:    instance_0.TtData_0,
		TtData:      &tt_data,
		NewInstance: instance_0.NewInstance,
		Stop:        instance_0.Stop,
		//TODO: follow-on paths
	}

	// Keep only the hard-blocked time slots and the fixed activities.

	// Remove activity constraints
	tt_data.Constraints = map[string][]any{}
	tt_data.MinDaysBetweenLessons = nil
	tt_data.ParallelLessons = nil
	tt_data.WITHOUT_ROOM_PLACEMENTS = true

	// Regenerate the teachers list without constraints
	new_teachers := make([]*base.Teacher, len(db.Teachers))
	for i, t0p := range db.Teachers {
		t := *t0p
		t.MinLessonsPerDay = -1 // unconstrained
		t.MaxLessonsPerDay = -1 // unconstrained
		t.MaxDays = -1          // unconstrained
		t.MaxGapsPerDay = -1    // unconstrained
		t.MaxGapsPerWeek = -1   // unconstrained
		t.MaxAfternoons = -1    // unconstrained
		t.LunchBreak = false
		new_teachers[i] = &t
	}
	db.Teachers = new_teachers

	// Regenerate the classes list without constraints
	new_classes := make([]*base.Class, len(db.Classes))
	for i, c0p := range db.Classes {
		c := *c0p
		c.MinLessonsPerDay = -1 // unconstrained
		c.MaxLessonsPerDay = -1 // unconstrained
		c.MaxGapsPerDay = -1    // unconstrained
		c.MaxGapsPerWeek = -1   // unconstrained
		c.MaxAfternoons = -1    // unconstrained
		c.LunchBreak = false
		c.ForceFirstHour = false
		new_classes[i] = &c
	}
	db.Classes = new_classes

	// Request start of instance
	instance.NewInstance <- instance
	return instance
}

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
