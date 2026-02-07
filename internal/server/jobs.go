package server

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"time"
)

type JobStatus string

const (
	JobPending JobStatus = "pending"
	JobRunning JobStatus = "running"
	JobDone    JobStatus = "done"
	JobFailed  JobStatus = "failed"

	jobTTL          = 5 * time.Minute
	jobCleanupEvery = 30 * time.Second
)

type Job struct {
	ID           string
	Status       JobStatus
	Percent      int
	InPath       string
	OutPath      string
	OriginalName string
	Error        string
	CreatedAt    time.Time
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
	j := &Job{
		ID:           randHex(8),
		Status:       JobPending,
		InPath:       inPath,
		OutPath:      outPath,
		OriginalName: originalName,
		CreatedAt:    time.Now(),
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

func (s *JobStore) cleanup() {
	for {
		time.Sleep(jobCleanupEvery)
		s.mu.Lock()
		now := time.Now()
		for id, j := range s.jobs {
			if now.Sub(j.CreatedAt) > jobTTL {
				_ = os.Remove(j.InPath)
				_ = os.Remove(j.OutPath)
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
