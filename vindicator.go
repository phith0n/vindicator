package vindicator

import (
	"context"
	"github.com/asaskevich/EventBus"
	"sync"
	"time"
)

type Worker interface {
	Work(ctx context.Context) error // the worker() must be a blocking function
	GetRunning() bool
	SetRunning(bool)
}

type Vindicator struct {
	interval    int // check worker every Interval seconds
	worker      Worker
	lock        sync.Mutex
	stopWorker  func() // stop the worker
	stopMonitor func() // stop the monitor
	bus         EventBus.Bus
}

type VindicatorFn func(v *Vindicator, args ...interface{})

func NewVindicator(worker Worker, interval int) *Vindicator {
	return &Vindicator{
		interval: interval,
		worker:   worker,
		bus:      EventBus.New(),
	}
}

func (v *Vindicator) Start(ctx context.Context) error {
	v.SetRunning()
	defer v.SetStopped()
	v.bus.Publish("worker:start", v)
	defer v.bus.Publish("worker:stop", v)

	newCtx, cancel := context.WithCancel(ctx)
	v.stopWorker = cancel

	if err := v.worker.Work(newCtx); err != nil {
		v.bus.Publish("worker:error", v, err)
		return err
	}

	return nil
}

func (v *Vindicator) Monitor(ctx context.Context) {
	v.bus.Publish("monitor:start", v)

	newCtx, cancel := context.WithCancel(ctx)
	v.stopMonitor = cancel
	for {
		time.Sleep(time.Duration(v.interval) * time.Second)

		select {
		case <-newCtx.Done():
			v.bus.Publish("monitor:stop", v)
			return
		default:
			if !v.worker.GetRunning() {
				v.bus.Publish("monitor:interrupt", v)
				go func() {
					_ = v.Start(ctx)
				}()
			} else {
				v.bus.Publish("monitor:working", v)
			}
		}
	}
}

func (v *Vindicator) Stop() {
	if v.stopMonitor != nil {
		v.stopMonitor()
	}

	if v.stopWorker != nil {
		v.stopWorker()
	}

	// block the Stop function until the worker is stopped
	v.Wait()
}

func (v *Vindicator) SetRunning() {
	v.lock.Lock()
	v.worker.SetRunning(true)
}

func (v *Vindicator) SetStopped() {
	v.worker.SetRunning(false)
	v.lock.Unlock()
}

func (v *Vindicator) Wait() {
	v.lock.Lock()
	v.lock.Unlock()
}

func (v *Vindicator) On(eventName string, callback VindicatorFn) {
	_ = v.bus.Subscribe(eventName, callback)
}
