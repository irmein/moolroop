package store_test

import (
	"fmt"
	"sync"
	"testing"

	"moolroop/internal/errors"
	"moolroop/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_CreateAndGetProfile(t *testing.T) {
	s := store.NewMemoryStore()

	profile, err := s.CreateProfile("Jane Doe", "jane@example.com")
	require.NoError(t, err)
	require.NotNil(t, profile)
	assert.NotEmpty(t, profile.UserId)
	assert.Equal(t, "Jane Doe", profile.FullName)
	assert.Equal(t, "jane@example.com", profile.Email)
	assert.NotNil(t, profile.CreatedAt)
	assert.NotNil(t, profile.UpdatedAt)

	fetched, err := s.GetProfile(profile.UserId)
	require.NoError(t, err)
	assert.Equal(t, profile.UserId, fetched.UserId)
	assert.Equal(t, "Jane Doe", fetched.FullName)
	assert.Equal(t, "jane@example.com", fetched.Email)

	// Non-existing user
	_, err = s.GetProfile("non-existent-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrUserNotFound)
}

func TestMemoryStore_PatchProfile(t *testing.T) {
	s := store.NewMemoryStore()

	p, err := s.CreateProfile("John Smith", "john@example.com")
	require.NoError(t, err)

	// Update only name
	newName := "Johnathan Smith"
	updated, err := s.PatchProfile(p.UserId, &newName, nil)
	require.NoError(t, err)
	assert.Equal(t, "Johnathan Smith", updated.FullName)
	assert.Equal(t, "john@example.com", updated.Email)

	// Update only email
	newEmail := "johnathan@example.com"
	updated, err = s.PatchProfile(p.UserId, nil, &newEmail)
	require.NoError(t, err)
	assert.Equal(t, "Johnathan Smith", updated.FullName)
	assert.Equal(t, "johnathan@example.com", updated.Email)

	// Update both
	newName2 := "J. Smith"
	newEmail2 := "jsmith@example.com"
	updated, err = s.PatchProfile(p.UserId, &newName2, &newEmail2)
	require.NoError(t, err)
	assert.Equal(t, "J. Smith", updated.FullName)
	assert.Equal(t, "jsmith@example.com", updated.Email)

	// Patch non-existing
	_, err = s.PatchProfile("unknown-user", &newName, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrUserNotFound)
}

func TestMemoryStore_AppendAndListActivities(t *testing.T) {
	s := store.NewMemoryStore()

	p, err := s.CreateProfile("Alice", "alice@example.com")
	require.NoError(t, err)

	// List on new profile has 0 activities
	activities, err := s.ListActivities(p.UserId)
	require.NoError(t, err)
	assert.Empty(t, activities)

	// Append activity
	rec1, err := s.AppendActivity(p.UserId, "LOGIN", "User logged in")
	require.NoError(t, err)
	assert.NotEmpty(t, rec1.ActivityId)
	assert.Equal(t, p.UserId, rec1.UserId)
	assert.Equal(t, "LOGIN", rec1.ActionType)
	assert.Equal(t, "User logged in", rec1.Description)
	assert.NotNil(t, rec1.LoggedAt)

	rec2, err := s.AppendActivity(p.UserId, "SETTINGS_CHANGE", "Updated 2FA")
	require.NoError(t, err)
	assert.NotEmpty(t, rec2.ActivityId)

	// List activities
	activities, err = s.ListActivities(p.UserId)
	require.NoError(t, err)
	require.Len(t, activities, 2)
	assert.Equal(t, rec1.ActivityId, activities[0].ActivityId)
	assert.Equal(t, rec2.ActivityId, activities[1].ActivityId)

	// Verify slice copy immutability
	activities[0] = nil
	activitiesAfter, err := s.ListActivities(p.UserId)
	require.NoError(t, err)
	assert.NotNil(t, activitiesAfter[0])

	// Append on non-existing user
	_, err = s.AppendActivity("unknown-id", "ACTION", "desc")
	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrUserNotFound)

	// List on non-existing user
	_, err = s.ListActivities("unknown-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrUserNotFound)
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	s := store.NewMemoryStore()

	p, err := s.CreateProfile("Bob", "bob@example.com")
	require.NoError(t, err)

	var wg sync.WaitGroup
	numWorkers := 50

	// Concurrent Profile Readers & Updaters
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Read profile
			_, _ = s.GetProfile(p.UserId)

			// Update profile occasionally
			if idx%5 == 0 {
				name := fmt.Sprintf("Bob %d", idx)
				_, _ = s.PatchProfile(p.UserId, &name, nil)
			}
		}(i)
	}

	// Concurrent Activity Loggers & Readers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _ = s.AppendActivity(p.UserId, "EVENT", fmt.Sprintf("Log event %d", idx))
			_, _ = s.ListActivities(p.UserId)
		}(i)
	}

	wg.Wait()

	activities, err := s.ListActivities(p.UserId)
	require.NoError(t, err)
	assert.Len(t, activities, numWorkers)
}
