package services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const sessionMaxIdle    = 2 * time.Minute
const sessionGCInterval = 30 * time.Second

// StreamSession holds the lifecycle context for one user's active stream.
type StreamSession struct {
	ctx       context.Context
	cancel    context.CancelFunc
	job       *TranscodeJob
	createdAt time.Time
	mu        sync.Mutex
}

// StreamSessionManager tracks active streams keyed by "userID:infoHash".
type StreamSessionManager struct {
	mu       sync.Mutex
	sessions map[string]*StreamSession
}

// Sessions is the process-wide session manager.
var Sessions = newStreamSessionManager()

func newStreamSessionManager() *StreamSessionManager {
	m := &StreamSessionManager{sessions: make(map[string]*StreamSession)}
	go m.runGC()
	return m
}

// Acquire registers a new stream session for (userID, infoHash), replacing any
// existing one. The ffmpeg process is killed automatically when reqCtx is done
// (client disconnect) or when Release is called.
func (m *StreamSessionManager) Acquire(userID int, infoHash string, job *TranscodeJob, reqCtx context.Context) *StreamSession {
	key := fmt.Sprintf("%d:%s", userID, infoHash)
	ctx, cancel := context.WithCancel(context.Background())
	s := &StreamSession{ctx: ctx, cancel: cancel, job: job, createdAt: time.Now()}

	m.mu.Lock()
	if old, ok := m.sessions[key]; ok {
		old.kill()
	}
	m.sessions[key] = s
	m.mu.Unlock()

	go func() {
		select {
		case <-reqCtx.Done():
			m.remove(key)
		case <-ctx.Done():
		}
	}()

	return s
}

// Release removes the session and kills the associated ffmpeg process.
func (m *StreamSessionManager) Release(userID int, infoHash string) {
	m.remove(fmt.Sprintf("%d:%s", userID, infoHash))
}

func (m *StreamSessionManager) remove(key string) {
	m.mu.Lock()
	s, ok := m.sessions[key]
	if ok {
		delete(m.sessions, key)
	}
	m.mu.Unlock()
	if ok {
		s.kill()
	}
}

func (m *StreamSessionManager) runGC() {
	t := time.NewTicker(sessionGCInterval)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		var stale []string
		m.mu.Lock()
		for k, s := range m.sessions {
			if now.Sub(s.createdAt) > sessionMaxIdle {
				stale = append(stale, k)
			}
		}
		for _, k := range stale {
			if s, ok := m.sessions[k]; ok {
				delete(m.sessions, k)
				go s.kill()
			}
		}
		m.mu.Unlock()
	}
}

func (s *StreamSession) kill() {
	s.cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil {
		return
	}
	s.job.Reader.Close()
	if s.job.Cmd != nil && s.job.Cmd.Process != nil {
		s.job.Cmd.Process.Kill()
	}
	s.job = nil
}
