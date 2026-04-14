package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// JobStatus represents the lifecycle state of a queued job.
type JobStatus string

const (
	JobPending    JobStatus = "pending"
	JobProcessing JobStatus = "processing"
	JobCompleted  JobStatus = "completed"
	JobFailed     JobStatus = "failed"
)

// Job is a unit of work submitted to the queue.
type Job struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Status    JobStatus              `json:"status"`
	Result    string                 `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	DoneAt    *time.Time             `json:"done_at,omitempty"`
}

// JobHandler is a function that processes a job and returns a result or error.
type JobHandler func(ctx context.Context, job *Job) (string, error)

// QueueService is an in-memory, channel-backed job queue with a worker pool.
type QueueService struct {
	jobs     chan *Job
	handlers map[string]JobHandler
	mu       sync.RWMutex
	registry map[string]*Job // job ID → job (for status lookup)
	counter  atomic.Uint64
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// QueueStats returns queue health metrics.
type QueueStats struct {
	Pending    int `json:"pending"`
	Processing int `json:"processing"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
	Total      int `json:"total"`
	Capacity   int `json:"capacity"`
}

// NewQueueService creates a queue with the given buffer size and starts workers.
func NewQueueService(bufferSize, workerCount int) *QueueService {
	ctx, cancel := context.WithCancel(context.Background())
	q := &QueueService{
		jobs:     make(chan *Job, bufferSize),
		handlers: make(map[string]JobHandler),
		registry: make(map[string]*Job),
		ctx:      ctx,
		cancel:   cancel,
	}
	for i := 0; i < workerCount; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
	return q
}

// RegisterHandler maps a job type to its processing function.
func (q *QueueService) RegisterHandler(jobType string, h JobHandler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[jobType] = h
}

// Submit enqueues a new job and returns it immediately.
func (q *QueueService) Submit(jobType string, payload map[string]interface{}) (*Job, error) {
	id := fmt.Sprintf("job-%d", q.counter.Add(1))
	job := &Job{
		ID:        id,
		Type:      jobType,
		Payload:   payload,
		Status:    JobPending,
		CreatedAt: time.Now(),
	}

	q.mu.Lock()
	q.registry[id] = job
	q.mu.Unlock()

	select {
	case q.jobs <- job:
		return job, nil
	default:
		job.Status = JobFailed
		job.Error = "queue is full"
		return job, fmt.Errorf("queue is full (capacity %d)", cap(q.jobs))
	}
}

// GetJob returns a job by ID.
func (q *QueueService) GetJob(id string) (*Job, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	j, ok := q.registry[id]
	return j, ok
}

// Stats returns current queue statistics.
func (q *QueueService) Stats() QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()

	stats := QueueStats{
		Total:    len(q.registry),
		Capacity: cap(q.jobs),
	}
	for _, j := range q.registry {
		switch j.Status {
		case JobPending:
			stats.Pending++
		case JobProcessing:
			stats.Processing++
		case JobCompleted:
			stats.Completed++
		case JobFailed:
			stats.Failed++
		}
	}
	return stats
}

// Shutdown gracefully stops the queue and waits for workers to finish.
func (q *QueueService) Shutdown() {
	q.cancel()
	close(q.jobs)
	q.wg.Wait()
}

func (q *QueueService) worker(id int) {
	defer q.wg.Done()
	for job := range q.jobs {
		q.processJob(id, job)
	}
}

func (q *QueueService) processJob(workerID int, job *Job) {
	job.Status = JobProcessing

	q.mu.RLock()
	handler, ok := q.handlers[job.Type]
	q.mu.RUnlock()

	if !ok {
		job.Status = JobFailed
		job.Error = fmt.Sprintf("no handler for job type %q", job.Type)
		now := time.Now()
		job.DoneAt = &now
		log.Printf("[worker-%d] job %s FAILED: %s", workerID, job.ID, job.Error)
		return
	}

	ctx, cancel := context.WithTimeout(q.ctx, 30*time.Second)
	defer cancel()

	result, err := handler(ctx, job)
	now := time.Now()
	job.DoneAt = &now

	if err != nil {
		job.Status = JobFailed
		job.Error = err.Error()
		log.Printf("[worker-%d] job %s FAILED: %v", workerID, job.ID, err)
		return
	}

	job.Status = JobCompleted
	job.Result = result
	log.Printf("[worker-%d] job %s COMPLETED", workerID, job.ID)
}
