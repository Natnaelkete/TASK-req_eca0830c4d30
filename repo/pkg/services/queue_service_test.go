package services

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueueService(t *testing.T) {
	q := NewQueueService(10, 2)
	defer q.Shutdown()
	assert.NotNil(t, q)
}

func TestQueueService_SubmitAndProcess(t *testing.T) {
	q := NewQueueService(10, 1)
	defer q.Shutdown()

	var processed atomic.Bool
	q.RegisterHandler("test_job", func(ctx context.Context, job *Job) (string, error) {
		processed.Store(true)
		return "done", nil
	})

	job, err := q.Submit("test_job", map[string]interface{}{"key": "value"})
	require.NoError(t, err)
	// Status may already be processing/completed due to worker race; verify it was queued
	assert.Contains(t, []JobStatus{JobPending, JobProcessing}, job.Status)

	// Wait for worker to process
	time.Sleep(200 * time.Millisecond)

	assert.True(t, processed.Load())
	found, ok := q.GetJob(job.ID)
	require.True(t, ok)
	assert.Equal(t, JobCompleted, found.Status)
	assert.Equal(t, "done", found.Result)
	assert.NotNil(t, found.DoneAt)
}

func TestQueueService_FailedJob(t *testing.T) {
	q := NewQueueService(10, 1)
	defer q.Shutdown()

	q.RegisterHandler("fail_job", func(ctx context.Context, job *Job) (string, error) {
		return "", assert.AnError
	})

	job, err := q.Submit("fail_job", nil)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	found, ok := q.GetJob(job.ID)
	require.True(t, ok)
	assert.Equal(t, JobFailed, found.Status)
	assert.Contains(t, found.Error, "assert.AnError")
}

func TestQueueService_UnknownHandler(t *testing.T) {
	q := NewQueueService(10, 1)
	defer q.Shutdown()

	job, err := q.Submit("unknown_type", nil)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	found, ok := q.GetJob(job.ID)
	require.True(t, ok)
	assert.Equal(t, JobFailed, found.Status)
	assert.Contains(t, found.Error, "no handler")
}

func TestQueueService_Stats(t *testing.T) {
	q := NewQueueService(10, 1)
	defer q.Shutdown()

	q.RegisterHandler("stat_job", func(ctx context.Context, job *Job) (string, error) {
		return "ok", nil
	})

	q.Submit("stat_job", nil)
	q.Submit("stat_job", nil)

	time.Sleep(200 * time.Millisecond)

	stats := q.Stats()
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 2, stats.Completed)
	assert.Equal(t, 10, stats.Capacity)
}

func TestQueueService_GetJob_NotFound(t *testing.T) {
	q := NewQueueService(10, 1)
	defer q.Shutdown()

	_, ok := q.GetJob("nonexistent")
	assert.False(t, ok)
}

func TestQueueService_QueueFull(t *testing.T) {
	// buffer=1, worker=0 (no workers to drain) — but we can't have 0 workers,
	// so use buffer=1 with a slow handler
	q := NewQueueService(1, 1)
	defer q.Shutdown()

	q.RegisterHandler("slow", func(ctx context.Context, job *Job) (string, error) {
		time.Sleep(500 * time.Millisecond)
		return "ok", nil
	})

	// First job goes into channel and starts processing
	_, err1 := q.Submit("slow", nil)
	require.NoError(t, err1)

	time.Sleep(50 * time.Millisecond) // Let worker pick it up

	// Second job fills the buffer
	_, err2 := q.Submit("slow", nil)
	require.NoError(t, err2)

	// Third job should overflow
	_, err3 := q.Submit("slow", nil)
	assert.Error(t, err3)
}
