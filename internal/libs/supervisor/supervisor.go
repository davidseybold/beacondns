package supervisor

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Process interface {
	Start() error
	Stop() error
}

type Supervisor struct {
	processes []Process
}

type SupervisorOption func(*Supervisor)

func New(opts ...SupervisorOption) *Supervisor {
	s := &Supervisor{
		processes: make([]Process, 0),
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func WithProcess(p Process) SupervisorOption {
	return func(s *Supervisor) {
		s.addProcess(p)
	}
}

func (s *Supervisor) Start(ctx context.Context) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var eg errgroup.Group

	for _, p := range s.processes {
		p := p
		eg.Go(func() error {
			return p.Start()
		})
	}

	eg.Go(func() error {
		<-cancelCtx.Done()
		return s.stop()
	})

	return eg.Wait()
}

func (s *Supervisor) stop() error {
	var eg errgroup.Group
	for _, p := range s.processes {
		lp := p
		eg.Go(func() error {
			return lp.Stop()
		})
	}
	return eg.Wait()
}

func (s *Supervisor) addProcess(rs Process) {
	s.processes = append(s.processes, rs)
}
