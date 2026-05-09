package app

import (
	"sort"
	"sync"
	"time"
)

type capturedEntry struct {
	value       string
	endpointKey string
	at          time.Time
}

type capturedStore struct {
	mu   sync.RWMutex
	data map[string]map[string]capturedEntry
}

func newCapturedStore() *capturedStore {
	return &capturedStore{data: make(map[string]map[string]capturedEntry)}
}

func (s *capturedStore) set(projectID, name, value, endpointKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.data[projectID]
	if !ok {
		bucket = make(map[string]capturedEntry)
		s.data[projectID] = bucket
	}
	bucket[name] = capturedEntry{value: value, endpointKey: endpointKey, at: time.Now().UTC()}
}

func (s *capturedStore) clear(projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, projectID)
}

func (s *capturedStore) pruneByEndpoint(projectID, endpointKey string, keep map[string]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket, ok := s.data[projectID]
	if !ok {
		return
	}
	for name, entry := range bucket {
		if entry.endpointKey == endpointKey && !keep[name] {
			delete(bucket, name)
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
