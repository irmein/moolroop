package store

import (
	"sync"
	"time"

	"github.com/google/uuid"
	v1 "github.com/moolroop/gen/v1"
	apperrors "github.com/moolroop/internal/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrUserNotFound = apperrors.ErrUserNotFound
	ErrInvalidEmail = apperrors.ErrInvalidEmail
)

type MemoryStore struct {
	profileMu  sync.RWMutex
	profiles   map[string]*v1.UserProfile
	activityMu sync.RWMutex
	activities map[string][]*v1.ActivityRecord // Append-only activity logs
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		profiles:   make(map[string]*v1.UserProfile),
		activities: make(map[string][]*v1.ActivityRecord),
	}
}

func cloneProfile(p *v1.UserProfile) *v1.UserProfile {
	if p == nil {
		return nil
	}
	return proto.Clone(p).(*v1.UserProfile)
}

func cloneActivity(a *v1.ActivityRecord) *v1.ActivityRecord {
	if a == nil {
		return nil
	}
	return proto.Clone(a).(*v1.ActivityRecord)
}

func (s *MemoryStore) CreateProfile(name, email string) (*v1.UserProfile, error) {
	s.profileMu.Lock()
	defer s.profileMu.Unlock()

	now := timestamppb.New(time.Now().UTC())
	profile := &v1.UserProfile{
		UserId:    uuid.New().String(),
		FullName:  name,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.profiles[profile.UserId] = profile
	return cloneProfile(profile), nil
}

func (s *MemoryStore) GetProfile(userID string) (*v1.UserProfile, error) {
	s.profileMu.RLock()
	defer s.profileMu.RUnlock()

	p, exists := s.profiles[userID]
	if !exists {
		return nil, ErrUserNotFound
	}
	return cloneProfile(p), nil
}

func (s *MemoryStore) PatchProfile(userID string, name, email *string) (*v1.UserProfile, error) {
	s.profileMu.Lock()
	defer s.profileMu.Unlock()

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
	return cloneProfile(p), nil
}

// Append-only: creates an immutable historical event record
func (s *MemoryStore) AppendActivity(userID, actionType, description string) (*v1.ActivityRecord, error) {
	s.profileMu.RLock()
	_, exists := s.profiles[userID]
	s.profileMu.RUnlock()

	if !exists {
		return nil, ErrUserNotFound
	}

	s.activityMu.Lock()
	defer s.activityMu.Unlock()

	record := &v1.ActivityRecord{
		ActivityId:  uuid.New().String(),
		UserId:      userID,
		ActionType:  actionType,
		Description: description,
		LoggedAt:    timestamppb.New(time.Now().UTC()),
	}

	s.activities[userID] = append(s.activities[userID], record)
	return cloneActivity(record), nil
}

func (s *MemoryStore) ListActivities(userID string) ([]*v1.ActivityRecord, error) {
	s.profileMu.RLock()
	_, exists := s.profiles[userID]
	s.profileMu.RUnlock()

	if !exists {
		return nil, ErrUserNotFound
	}

	s.activityMu.RLock()
	defer s.activityMu.RUnlock()

	records := s.activities[userID]
	copied := make([]*v1.ActivityRecord, len(records))
	for i, r := range records {
		copied[i] = cloneActivity(r)
	}
	return copied, nil
}
