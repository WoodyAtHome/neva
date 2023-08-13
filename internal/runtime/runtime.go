// Package runtime implements environment for dataflow programs execution.
package runtime

import (
	"context"
	"errors"
	"fmt"
)

type Runtime struct {
	funcRunner FuncRunner
	connector  Connector
}

var ErrNilDeps = errors.New("runtime deps nil")

func New(c Connector, f FuncRunner) (Runtime, error) {
	if c == nil || f == nil {
		return Runtime{}, ErrNilDeps
	}
	return Runtime{
		connector:  c,
		funcRunner: f,
	}, nil
}

type (
	Connector interface {
		Connect(context.Context, []Connection) error
	}
	FuncRunner interface {
		Run(context.Context, []FuncRoutine) error
	}
	Func func(context.Context, FuncIO) error
)

var (
	ErrStartPortNotFound = errors.New("start port not found")
	ErrExitPortNotFound  = errors.New("exit port not found")
	ErrConnector         = errors.New("connector")
	ErrRoutineRunner     = errors.New("routine runner")
)

func (r Runtime) Run(ctx context.Context, prog Program) (code int, err error) {
	// FirstByName is not how this supposed to be working! There could be more "start" and "exit" ports!
	startPort := prog.Ports[PortAddr{Path: "main/in", Name: "start"}]
	if startPort == nil {
		return 0, ErrStartPortNotFound
	}

	exitPort := prog.Ports[PortAddr{Path: "main/in", Name: "start"}]
	if exitPort == nil {
		return 0, ErrExitPortNotFound
	}

	ctx, cancel := context.WithCancel(ctx)
	g, gctx := WithContext(ctx)

	g.Go(func() error {
		if err := r.connector.Connect(gctx, prog.Connections); err != nil {
			return fmt.Errorf("%w: %v", ErrConnector, err)
		}
		return nil
	})

	g.Go(func() error {
		if err := r.funcRunner.Run(gctx, prog.Funcs); err != nil {
			return fmt.Errorf("%w: %v", ErrRoutineRunner, err)
		}
		return nil
	})

	go func() { startPort <- nil }()

	var exitCode int
	go func() {
		exitCode = (<-exitPort).Int()
		cancel()
	}()

	return exitCode, g.Wait()
}
