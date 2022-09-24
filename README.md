# Vindicator

Vindicator is a lightweight Golang library that is designed to hold and check any blocking function, e.g. subprocess,
network connection...

## Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
  - [Vindicator](#vindicator-struct)
  - [Worker Interface](#worker-interface)
  - [Event](#event)
- [FAQ](#faq)
- [Contributing](#contributing)
- [License](#license)

## Installation

```go
go get -u github.com/phith0n/vindicator
```

## Quick Start

You have to write a struct I call it "Worker", that implements `vindicator.Worker`:

```go
type Worker interface {
    Work(ctx context.Context) error // the worker() must be a blocking function
    GetRunning() bool
    SetRunning(bool)
}
```

There are 3 functions in the `Worker` interface:

- `Work (ctx context.Context) error` this function must be a blocking function. the monitor will start a new worker if
  this function exit unexpected.
- `SetRunning(run bool)` this function should set the running status of the worker
- `GetRunning() bool` this function should return current running status

The `Work` function accepts a context object, it must control the lifecycle of your worker.

For example, if you want to start a subprocess and check it regularly if it is still running, here is the struct
implement:

```go
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
```

Then, use `Vindicator` to start and monitor the `ProcessWorker`:

```go
ctx := context.Background()
worker := ProcessWorker{}
v := vindicator.NewVindicator(&worker, 2)

// you can use event listener to execute some callback function
v.On("monitor:start", func(v *vindicator.Vindicator, args ...interface{}) {
    fmt.Println("start monitor")
})
v.On("monitor:working", func (v *vindicator.Vindicator, args ...interface{}) {
    fmt.Println("process is working normally...")
})
v.On("monitor:interrupt", func (v *vindicator.Vindicator, args ...interface{}) {
    fmt.Println("process is stopped unexpected, try to restart it...")
})
v.On("monitor:stop", func (v *vindicator.Vindicator, args ...interface{}) {
    fmt.Println("stop monitor")
})
v.On("worker:start", func (v *vindicator.Vindicator, args ...interface{}) {
    fmt.Println("start worker")
})
v.On("worker:stop", func (v *vindicator.Vindicator, args ...interface{}) {
    fmt.Println("stop worker")
})

// run worker and monitor in background
go v.Start(ctx)
go v.Monitor(ctx)

// to wait sometime...
timer := time.NewTimer(time.Second * 10)
<-timer.C

// demonstrate how to stop the worker and the monitor manual
v.Stop()
```

This example checks the process running status every 2 seconds, and stop it after 10 seconds.

The output:

```
start worker
start monitor
process is working normally...
process is working normally...
process is working normally...
process is working normally...
process is working normally...
stop monitor
stop worker
```

The full example code you can find [here](examples/subprocess_test.go).

## API Reference

### Vindicator Struct

Create a new `Vindicator`:

```go
v := vindicator.NewVindicator(&worker, 2)
```

The first argument is your custom `Worker` implements, the second argument is the monitor cycle time by seconds.

### Worker Interface

```go
type Worker interface {
    Work(ctx context.Context) error // the worker() must be a blocking function
    GetRunning() bool
    SetRunning(bool)
}
```

### Event

There are several events that you can listen and execute custom callback functions:

- `monitor:start` trigger when monitor is started
- `monitor:stop` trigger when monitor is stopped manual
- `monitor:interrupt` trigger when monitor is stopped unexpected
- `monitor:working` trigger when monitor works at cycle run
- `worker:start` trigger when worker is started 
- `worker:stop` trigger when worker is stopped
- `worker:error` trigger when an error is raised by worker

Use `Vindicator.On` to register a listener:

```go
type VindicatorFn func(v *Vindicator, args ...interface{})

func (v *Vindicator) On(eventName string, callback VindicatorFn) {
	// ...
}
```

## FAQ

**When should I use this library?**

You can use **github.com/phith0n/vindicator** when you are going to run a blocking function and maintain its status. For example, the subprocess, the TCP long connection, the Websocket connection, and any other program like these.

**Is there a document for this library?**

No yet. But there are only 100+ lines code for this project, you can kindly read the code and understand it by yourself.

## Contributing

If you'd like to help out with the project. You can put up a Pull Request.

## License

The Vindicator is open-sourced software licensed under the [MIT License](LICENSE).
