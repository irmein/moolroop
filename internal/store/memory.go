package store

import (
	"sync"
	"time"

	"github.com/google/uuid"
	v1 "moolroop/gen/v1"
	apperrors "moolroop/internal/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrUserNotFound = apperrors.ErrUserNotFound
	ErrInvalidEmail = apperrors.ErrInvalidEmail
)

type MemoryStore struct {
	mu         sync.RWMutex
	profiles   map[string]*v1.UserProfile
	activities map[string][]*v1.ActivityRecord // Append-only activity logs
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		profiles:   make(map[string]*v1.UserProfile),
		activities: make(map[string][]*v1.ActivityRecord),
	}
}

func (s *MemoryStore) CreateProfile(name, email string) (*v1.UserProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := timestamppb.New(time.Now().UTC())
	profile := &v1.UserProfile{
		UserId:    uuid.New().String(),
		FullName:  name,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.profiles[profile.UserId] = profile
	return profile, nil
}

func (s *MemoryStore) GetProfile(userID string) (*v1.UserProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, exists := s.profiles[userID]
	if !exists {
		return nil, ErrUserNotFound
	}
	return p, nil
}

func (s *MemoryStore) PatchProfile(userID string, name, email *string) (*v1.UserProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, exists := s.profiles[userID]
	if !exists {
		return nil, ErrUserNotFound
	}

	if name != nil {
		p.FullName = *name
	}
	if email != nil {
		p.Email = *email
	}
	p.UpdatedAt = timestamppb.New(time.Now().UTC())
	return p, nil
}

// Append-only: creates an immutable historical event record
func (s *MemoryStore) AppendActivity(userID, actionType, description string) (*v1.ActivityRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.profiles[userID]; !exists {
		return nil, ErrUserNotFound
	}

	record := &v1.ActivityRecord{
		ActivityId:  uuid.New().String(),
		UserId:      userID,
		ActionType:  actionType,
		Description: description,
		LoggedAt:    timestamppb.New(time.Now().UTC()),
	}

	s.activities[userID] = append(s.activities[userID], record)
	return record, nil
}

func (s *MemoryStore) ListActivities(userID string) ([]*v1.ActivityRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.profiles[userID]; !exists {
		return nil, ErrUserNotFound
	}

	records := s.activities[userID]
	copied := make([]*v1.ActivityRecord, len(records))
	copy(copied, records)
	return copied, nil
}
