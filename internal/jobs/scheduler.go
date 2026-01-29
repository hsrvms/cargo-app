package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Job interface {
	Start(ctx context.Context)
	GetName() string
}

type JobScheduler struct {
	jobs      map[string]Job
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.RWMutex
	started   bool
	startTime time.Time
}

type SchedulerConfig struct {
	GracefulShutdownTimeout time.Duration
}

func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		GracefulShutdownTimeout: 30 * time.Second,
	}
}

func NewJobScheduler() *JobScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &JobScheduler{
		jobs:   make(map[string]Job),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *JobScheduler) RegisterJob(name string, job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return ErrSchedulerAlreadyStarted
	}

	if _, exists := s.jobs[name]; exists {
		return ErrJobAlreadyExists
	}

	s.jobs[name] = job
	log.Printf("Registered job: %s", name)
	return nil
}

func (s *JobScheduler) UnregisterJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return ErrSchedulerAlreadyStarted
	}

	if _, exists := s.jobs[name]; !exists {
		return ErrJobNotFound
	}

	delete(s.jobs, name)
	log.Printf("Unregistered job: %s", name)
	return nil
}

func (s *JobScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return ErrSchedulerAlreadyStarted
	}

	s.started = true
	s.startTime = time.Now()

	log.Printf("Starting job scheduler with %d registered jobs", len(s.jobs))

	for name, job := range s.jobs {
		s.wg.Add(1)
		go func(jobName string, j Job) {
			defer s.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Job %s panicked: %v", jobName, r)
				}
			}()

			log.Printf("Starting job: %s", jobName)
			j.Start(s.ctx)
			log.Printf("Job stopped: %s", jobName)
		}(name, job)
	}

	log.Println("Job scheduler started successfully")
	return nil
}

func (s *JobScheduler) Stop() error {
	return s.StopWithTimeout(DefaultSchedulerConfig().GracefulShutdownTimeout)
}

func (s *JobScheduler) StopWithTimeout(timeout time.Duration) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return ErrSchedulerNotStarted
	}
	s.mu.Unlock()

	log.Printf("Stopping job scheduler (timeout: %v)", timeout)

	s.cancel()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All jobs stopped gracefully")
	case <-time.After(timeout):
		log.Printf("Jobs did not stop within timeout (%v), forcing shutdown", timeout)
		return ErrShutdownTimeout
	}

	s.mu.Lock()
	s.started = false
	s.mu.Unlock()

	uptime := time.Since(s.startTime)
	log.Printf("Job scheduler stopped after running for %v", uptime)
	return nil
}

func (s *JobScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

func (s *JobScheduler) GetRegisteredJobs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.jobs))
	for name := range s.jobs {
		names = append(names, name)
	}
	return names
}

func (s *JobScheduler) GetUptime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.started {
		return 0
	}
	return time.Since(s.startTime)
}

func (s *JobScheduler) GetContext() context.Context {
	return s.ctx
}

type JobWrapper struct {
	name string
	fn   func(ctx context.Context)
}

func NewJobWrapper(name string, fn func(ctx context.Context)) *JobWrapper {
	return &JobWrapper{
		name: name,
		fn:   fn,
	}
}

func (jw *JobWrapper) Start(ctx context.Context) {
	jw.fn(ctx)
}

func (jw *JobWrapper) GetName() string {
	return jw.name
}

var (
	ErrSchedulerAlreadyStarted = fmt.Errorf("scheduler is already started")
	ErrSchedulerNotStarted     = fmt.Errorf("scheduler is not started")
	ErrJobAlreadyExists        = fmt.Errorf("job already exists")
	ErrJobNotFound             = fmt.Errorf("job not found")
	ErrShutdownTimeout         = fmt.Errorf("shutdown timeout exceeded")
)
