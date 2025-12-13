package services

import (
	"fmt"

	"github.com/nicolailuther/butter/internal/cache"
	"github.com/nicolailuther/butter/internal/config"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/store"
	"github.com/nicolailuther/butter/pkg/kv"
	"go.uber.org/zap"
)

type Service struct {
	store     *store.Store
	logger    *zap.SugaredLogger
	config    *config.AppConfig
	cacheKeys *cache.CacheKeys
	cache     kv.KeyValueStore
}

func NewService(
	store *store.Store,
	logger *zap.SugaredLogger,
	config *config.AppConfig,
	cacheKeys *cache.CacheKeys,
	cache kv.KeyValueStore,
) *Service {
	return &Service{
		store,
		logger,
		config,
		cacheKeys,
		cache,
	}
}

// ResolveCountingUserID resolves the correct user ID for counting purposes.
// For admin and agency users, returns the user ID itself.
// For poster and team_leader users, returns their creator's ID.
func (s *Service) ResolveCountingUserID(userID uint) (uint, error) {
	user, err := s.store.Users.GetByID(userID)
	if err != nil {
		return 0, err
	}

	// If user is admin or agency, count for them directly
	if user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleAgency {
		return userID, nil
	}

	// For poster and team_leader, use their creator's ID for counting
	if user.Role == enums.UserRolePoster || user.Role == enums.UserRoleLeader {
		if user.CreatedBy == 0 {
			return 0, fmt.Errorf("poster/team_leader user %d has no creator set", userID)
		}
		return user.CreatedBy, nil
	}

	// For other roles (user, client), count for them directly
	return userID, nil
}

// ResolveCountingUser resolves the correct user for counting purposes.
// Returns both the user object and the counting user ID.
// For admin and agency users, returns the user itself.
// For poster and team_leader users, returns their creator.
func (s *Service) ResolveCountingUser(user *models.User) (*models.User, uint, error) {
	// If user is admin or agency, count for them directly
	if user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleAgency {
		return user, user.ID, nil
	}

	// For poster and team_leader, use their creator for counting
	if user.Role == enums.UserRolePoster || user.Role == enums.UserRoleLeader {
		if user.CreatedBy == 0 {
			return nil, 0, fmt.Errorf("poster/team_leader user %d has no creator set", user.ID)
		}
		// GetByID already preloads Creator, so we can use it directly
		if user.Creator != nil {
			return user.Creator, user.Creator.ID, nil
		}
		// If Creator is not loaded for some reason, fetch it
		creator, err := s.store.Users.GetByID(user.CreatedBy)
		if err != nil {
			return nil, 0, fmt.Errorf("error fetching creator for user %d: %w", user.ID, err)
		}
		return creator, creator.ID, nil
	}

	// For other roles (user, client), count for them directly
	return user, user.ID, nil
}
