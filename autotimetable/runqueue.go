package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"fmt"
	"slices"
)

type RunQueue struct {
	Queue      []*TtInstance
	Active     map[*TtInstance]struct{}
	MaxRunning int
	Next       int
}

func (rq *RunQueue) add(instance *TtInstance) {
	instance.ProcessingState = -1 // not started yet
	rq.Queue = append(rq.Queue, instance)
	//base.Message.Printf("(TODO) [%d] Queue %s\n",
	//	Ticks, instance.TtData.Description)
}

func (rq *RunQueue) add_front(instance *TtInstance) {
	instance.ProcessingState = -1 // not started yet
	rq.Queue = slices.Insert(rq.Queue, rq.Next, instance)
	base.Message.Printf("(TODO) [%d] Queue Front %s\n",
		Ticks, instance.TtData.Description)
}

//func (rq *RunQueue) disable() {
//	rq.MaxRunning = 0 // no new starts possible
//}

//func (rq *RunQueue) instance_completed(instance *TtInstance, state int) {
//	instance.ProcessingState = state
//	if state != 2 {
//		rq.instance_deactivate(instance)
//	}
//}

func (rq *RunQueue) instance_deactivate(instance *TtInstance) {
	delete(rq.Active, instance)
}

func (rq *RunQueue) update_instances() {
	// First increment the ticks of active instances.
	for instance := range rq.Active {
		ttdata := instance.TtData
		//base.Message.Printf("(TODO) [%d] ? ACTIVE (%d / %d): %s\n",
		//	Ticks, ttdata.State, instance.ProcessingState, ttdata.Description)
		if ttdata.State != 0 && instance.ProcessingState < 2 {
			// This should only be possible after the call to
			// `timetable.BACKEND.Tick` below.
			panic(fmt.Sprintf("Bug, State = %d", ttdata.State))
		}
		if ttdata.State == 0 {
			ttdata.Ticks++
			// Among other things, update the state:
			timetable.BACKEND.Tick(ttdata)
		} else if instance.ProcessingState < 2 {
			// This should only be possible after the call to
			// `timetable.BACKEND.Tick`.
			panic(fmt.Sprintf("Bug, State = %d", ttdata.State))
		}
		//	}

		//	for instance := range rq.Active {
		//		ttdata := instance.TtData

		//???
		if instance.ProcessingState == 3 {
			// Await completion of the goroutine
			if ttdata.State != 0 {
				rq.instance_deactivate(instance)
			}
			continue
		}

		switch ttdata.State {
		case 0: // running, not finished
			// check for timeout
			if instance.Timeout == ttdata.Ticks {
				base.Message.Printf("(TODO) [%d] Timeout %s @ %d (%d)\n",
					Ticks, ttdata.Description, ttdata.Ticks, ttdata.Progress)

				//TODO: even if it adds only one constraint?
				// Stop instance
				abort_instance(instance)
			}

		case 1: // completed successfully
			base.Message.Printf("(TODO) [%d] <<+ %s @ %d\n",
				Ticks, ttdata.Description, ttdata.Ticks)
			instance.ProcessingState = 1
			rq.instance_deactivate(instance)

		default: // completed unsuccessfully
			base.Message.Printf("(TODO) [%d] <<- %s @ %d\n",
				Ticks, ttdata.Description, ttdata.Ticks)
			instance.ProcessingState = 2
			rq.instance_deactivate(instance)
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

		base.Message.Printf("(TODO) [%d] >> %s\n",
			Ticks, instance.TtData.Description)
		timetable.BACKEND.Run(instance.TtData, TESTING)
	}
	//TODO--
	//fmt.Printf("$ [%d] Running/Active instances: %d/%d\n",
	//	Ticks, running, len(rq.Active))
	//base.Message.Printf("$ [%d] Running/Active instances: %d/%d\n",
	//	Ticks, running, len(rq.Active))
	return len(rq.Active)
}
