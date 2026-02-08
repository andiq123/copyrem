package server

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"time"
)

type JobStatus string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobDone      JobStatus = "done"
	JobFailed JobStatus = "failed"

	jobTTL          = 5 * time.Minute
	jobCleanupEvery = 30 * time.Second
)

type Job struct {
	ID           string
	Status       JobStatus
	Percent      int
	InPath       string
	OutPath      string
	OutPath2     string
	OriginalName string
	Error        string
	CreatedAt    time.Time
	Ctx          context.Context
	cancel       context.CancelFunc
}

type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

func NewJobStore() *JobStore {
	s := &JobStore{jobs: make(map[string]*Job)}
	go s.cleanup()
	return s
}

func (s *JobStore) Create(inPath, outPath, originalName string) *Job {
	return s.createJob(inPath, outPath, "", originalName)
}

func (s *JobStore) CreateWithTwoOutputs(inPath, outPath, outPath2, originalName string) *Job {
	return s.createJob(inPath, outPath, outPath2, originalName)
}

func (s *JobStore) createJob(inPath, outPath, outPath2, originalName string) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	j := &Job{
		ID:           randHex(8),
		Status:       JobPending,
		InPath:       inPath,
		OutPath:      outPath,
		OutPath2:     outPath2,
		OriginalName: originalName,
		CreatedAt:    time.Now(),
		Ctx:          ctx,
		cancel:       cancel,
	}
	s.mu.Lock()
	s.jobs[j.ID] = j
	s.mu.Unlock()
	return j
}

func (s *JobStore) Get(id string) *Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.jobs[id]
}

func (s *JobStore) SetRunning(id string) {
	s.mu.Lock()
	if j := s.jobs[id]; j != nil {
		j.Status = JobRunning
	}
	s.mu.Unlock()
}

func (s *JobStore) SetPercent(id string, pct int) {
	s.mu.Lock()
	if j := s.jobs[id]; j != nil {
		j.Percent = pct
	}
	s.mu.Unlock()
}

func (s *JobStore) SetDone(id string) {
	s.mu.Lock()
	if j := s.jobs[id]; j != nil {
		j.Status = JobDone
		j.Percent = 100
	}
	s.mu.Unlock()
}

func (s *JobStore) SetFailed(id string, errMsg string) {
	s.mu.Lock()
	if j := s.jobs[id]; j != nil {
		j.Status = JobFailed
		j.Error = errMsg
	}
	s.mu.Unlock()
}

func (s *JobStore) Cancel(id string) {
	s.mu.Lock()
	j := s.jobs[id]
	if j == nil {
		s.mu.Unlock()
		return
	}
	j.cancel()
	in, out, out2 := j.InPath, j.OutPath, j.OutPath2
	delete(s.jobs, id)
	s.mu.Unlock()
	_ = os.Remove(in)
	_ = os.Remove(out)
	if out2 != "" {
		_ = os.Remove(out2)
	}
}

func (s *JobStore) cleanup() {
	for {
		time.Sleep(jobCleanupEvery)
		s.mu.Lock()
		now := time.Now()
		for id, j := range s.jobs {
			if now.Sub(j.CreatedAt) > jobTTL {
				j.cancel()
				_ = os.Remove(j.InPath)
				_ = os.Remove(j.OutPath)
				if j.OutPath2 != "" {
					_ = os.Remove(j.OutPath2)
				}
				delete(s.jobs, id)
			}
		}
		s.mu.Unlock()
	}
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
