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
	rs []Process
}

type SupervisorOption func(*Supervisor)

func New(opts ...SupervisorOption) *Supervisor {
	s := &Supervisor{
		rs: make([]Process, 0),
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
	var eg errgroup.Group

	for _, rs := range s.rs {
		eg.Go(func() error {
			return rs.Start()
		})
	}

	go func(ctx context.Context) {
		<-ctx.Done()
		s.Stop()
	}(ctx)

	return eg.Wait()
}

func (s *Supervisor) Stop() error {
	for _, rs := range s.rs {
		if err := rs.Stop(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Supervisor) addProcess(rs Process) {
	s.rs = append(s.rs, rs)
}
