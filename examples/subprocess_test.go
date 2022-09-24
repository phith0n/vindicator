package main

import (
	"context"
	"fmt"
	"github.com/phith0n/vindicator"
	"os/exec"
	"testing"
	"time"
)

type ProcessWorker struct {
	isRunning bool
}

// Work must be a blocking function
func (pw *ProcessWorker) Work(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "sleep", "1h")
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (pw *ProcessWorker) SetRunning(run bool) {
	pw.isRunning = run
}

func (pw *ProcessWorker) GetRunning() bool {
	return pw.isRunning
}

func TestSubprocess(t *testing.T) {
	ctx := context.Background()
	worker := ProcessWorker{}
	v := vindicator.NewVindicator(&worker, 2)
	v.On("monitor:start", func(v *vindicator.Vindicator, args ...interface{}) {
		fmt.Println("start monitor")
	})
	v.On("monitor:working", func(v *vindicator.Vindicator, args ...interface{}) {
		fmt.Println("process is working normally...")
	})
	v.On("monitor:interrupt", func(v *vindicator.Vindicator, args ...interface{}) {
		fmt.Println("process is stopped unexpected, try to restart it...")
	})
	v.On("monitor:stop", func(v *vindicator.Vindicator, args ...interface{}) {
		fmt.Println("stop monitor")
	})
	v.On("worker:start", func(v *vindicator.Vindicator, args ...interface{}) {
		fmt.Println("start worker")
	})
	v.On("worker:stop", func(v *vindicator.Vindicator, args ...interface{}) {
		fmt.Println("stop worker")
	})

	go v.Start(ctx)
	go v.Monitor(ctx)

	timer := time.NewTimer(time.Second * 10)
	<-timer.C

	// stop the worker and the monitor manual
	v.Stop()

	time.Sleep(time.Second * 3)
}
