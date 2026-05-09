package app

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"spectra-desktop/internal/repository"
)

type capturedEntry struct {
	value       string
	endpointKey string
	at          time.Time
}

type capturedStore struct {
	mu   sync.RWMutex
	data map[string]map[string]capturedEntry
	repo *repository.CapturedValuesRepository
	ctx  func() context.Context
}

func newCapturedStore(repo *repository.CapturedValuesRepository, ctxFn func() context.Context) *capturedStore {
	return &capturedStore{
		data: make(map[string]map[string]capturedEntry),
		repo: repo,
		ctx:  ctxFn,
	}
}

func (s *capturedStore) ensureLoaded(projectID string) {
	s.mu.RLock()
	_, exists := s.data[projectID]
	s.mu.RUnlock()
	if exists {
		return
	}
	s.loadProject(projectID)
}

func (s *capturedStore) loadProject(projectID string) {
	if s.repo == nil {
		return
	}
	rows, err := s.repo.ListByProject(s.ctx(), projectID)
	if err != nil {
		log.Printf("captured_values load: %v", err)
		return
	}
	bucket := make(map[string]capturedEntry, len(rows))
	for _, r := range rows {
		bucket[r.Name] = capturedEntry{value: r.Value, endpointKey: r.EndpointKey, at: r.CapturedAt}
	}
	s.mu.Lock()
	s.data[projectID] = bucket
	s.mu.Unlock()
}

func (s *capturedStore) set(projectID, name, value, endpointKey string) {
	now := time.Now().UTC()
	s.mu.Lock()
	bucket, ok := s.data[projectID]
	if !ok {
		bucket = make(map[string]capturedEntry)
		s.data[projectID] = bucket
	}
	bucket[name] = capturedEntry{value: value, endpointKey: endpointKey, at: now}
	s.mu.Unlock()
	if s.repo != nil {
		if err := s.repo.Upsert(s.ctx(), projectID, name, value, endpointKey, now); err != nil {
			log.Printf("captured_values upsert: %v", err)
		}
	}
}

func (s *capturedStore) clear(projectID string) {
	s.mu.Lock()
	delete(s.data, projectID)
	s.mu.Unlock()
	if s.repo != nil {
		if err := s.repo.DeleteByProject(s.ctx(), projectID); err != nil {
			log.Printf("captured_values clear: %v", err)
		}
	}
}

func (s *capturedStore) pruneByEndpoint(projectID, endpointKey string, keep map[string]bool) {
	s.mu.Lock()
	bucket, ok := s.data[projectID]
	if ok {
		for name, entry := range bucket {
			if entry.endpointKey == endpointKey && !keep[name] {
				delete(bucket, name)
			}
		}
	}
	s.mu.Unlock()
	if s.repo != nil {
		if err := s.repo.DeleteByEndpoint(s.ctx(), projectID, endpointKey, keep); err != nil {
			log.Printf("captured_values prune: %v", err)
		}
	}
}

func (s *capturedStore) values(projectID string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.data[projectID]
	if !ok {
		return nil
	}
	out := make(map[string]string, len(bucket))
	for k, v := range bucket {
		out[k] = v.value
	}
	return out
}

func (s *capturedStore) list(projectID string) []CapturedValueDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.data[projectID]
	if !ok {
		return []CapturedValueDTO{}
	}
	out := make([]CapturedValueDTO, 0, len(bucket))
	for k, v := range bucket {
		out = append(out, CapturedValueDTO{
			Name:        k,
			Value:       v.value,
			EndpointKey: v.endpointKey,
			CapturedAt:  v.at.Unix(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
