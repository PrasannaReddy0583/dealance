package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/user/internal/domain/entity"
	"github.com/dealance/services/user/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

// ProfileService handles profile CRUD operations.
type ProfileService struct {
	profileRepo repository.ProfileRepository
	settingsRepo repository.SettingsRepository
	cacheRepo   repository.CacheRepository
	log         zerolog.Logger
}

func NewProfileService(
	profileRepo repository.ProfileRepository,
	settingsRepo repository.SettingsRepository,
	cacheRepo repository.CacheRepository,
	log zerolog.Logger,
) *ProfileService {
	return &ProfileService{
		profileRepo: profileRepo,
		settingsRepo: settingsRepo,
		cacheRepo:   cacheRepo,
		log:         log,
	}
}

// CreateProfile creates a new profile for a user (called after signup).
func (s *ProfileService) CreateProfile(ctx context.Context, req entity.CreateProfileRequest) (*entity.ProfileResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid user_id")
	}

	// Check username availability
	exists, err := s.profileRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	if exists {
		return nil, apperrors.ErrConflict(fmt.Sprintf("Username '%s' is already taken", req.Username))
	}

	profile := &entity.Profile{
		ID:          userID,
		Username:    req.Username,
		DisplayName: req.DisplayName,
		IsPublic:    true,
	}

	if err := s.profileRepo.Create(ctx, profile); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Create default settings
	if err := s.settingsRepo.Create(ctx, userID); err != nil {
		s.log.Error().Err(err).Msg("failed to create default settings")
	}

	return s.toProfileResponse(profile, false, false), nil
}

// GetProfile returns a user's profile.
func (s *ProfileService) GetProfile(ctx context.Context, profileID string, viewerID string) (*entity.ProfileResponse, error) {
	id, err := uuid.Parse(profileID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid profile ID")
	}

	// Try cache first
	if cached, err := s.cacheRepo.GetCachedProfile(ctx, profileID); err == nil && cached != nil {
		return s.toProfileResponse(cached, false, false), nil
	}

	profile, err := s.profileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.ErrNotFound("Profile")
	}

	// Cache for next request
	_ = s.cacheRepo.CacheProfile(ctx, profile)

	return s.toProfileResponse(profile, false, false), nil
}

// GetProfileByUsername returns a profile by username.
func (s *ProfileService) GetProfileByUsername(ctx context.Context, username string) (*entity.ProfileResponse, error) {
	profile, err := s.profileRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, apperrors.ErrNotFound("Profile")
	}
	return s.toProfileResponse(profile, false, false), nil
}

// UpdateProfile updates a user's profile.
func (s *ProfileService) UpdateProfile(ctx context.Context, userID string, req entity.UpdateProfileRequest) (*entity.ProfileResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid user ID")
	}

	updates := make(map[string]interface{})
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL
	}
	if req.CoverURL != nil {
		updates["cover_url"] = *req.CoverURL
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.Website != nil {
		updates["website"] = *req.Website
	}
	if req.LinkedInURL != nil {
		updates["linkedin_url"] = *req.LinkedInURL
	}
	if req.TwitterURL != nil {
		updates["twitter_url"] = *req.TwitterURL
	}
	if req.Profession != nil {
		updates["profession"] = *req.Profession
	}
	if req.Company != nil {
		updates["company"] = *req.Company
	}
	if req.ExperienceYears != nil {
		updates["experience_years"] = *req.ExperienceYears
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if len(updates) == 0 {
		return nil, apperrors.ErrValidation("no fields to update")
	}

	if err := s.profileRepo.Update(ctx, id, updates); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Invalidate cache
	_ = s.cacheRepo.InvalidateProfile(ctx, userID)

	// Fetch updated profile
	profile, err := s.profileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	return s.toProfileResponse(profile, false, false), nil
}

// SearchProfiles searches for profiles by name, username, or profession.
func (s *ProfileService) SearchProfiles(ctx context.Context, req entity.SearchUsersRequest) ([]entity.ProfileListItem, error) {
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	profiles, err := s.profileRepo.Search(ctx, req.Query, limit+1, nil, nil)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Trim to limit
	if len(profiles) > limit {
		profiles = profiles[:limit]
	}

	items := make([]entity.ProfileListItem, len(profiles))
	for i, p := range profiles {
		items[i] = entity.ProfileListItem{
			ID:          p.ID.String(),
			Username:    p.Username,
			DisplayName: p.DisplayName,
			AvatarURL:   p.AvatarURL,
			Bio:         p.Bio,
			Profession:  p.Profession,
			IsVerified:  p.IsVerified,
		}
	}

	return items, nil
}

func (s *ProfileService) toProfileResponse(p *entity.Profile, isFollowing, isFollowedBy bool) *entity.ProfileResponse {
	return &entity.ProfileResponse{
		ID:              p.ID.String(),
		Username:        p.Username,
		DisplayName:     p.DisplayName,
		Bio:             p.Bio,
		AvatarURL:       p.AvatarURL,
		CoverURL:        p.CoverURL,
		Location:        p.Location,
		Website:         p.Website,
		LinkedInURL:     p.LinkedInURL,
		TwitterURL:      p.TwitterURL,
		Profession:      p.Profession,
		Company:         p.Company,
		ExperienceYears: p.ExperienceYears,
		IsPublic:        p.IsPublic,
		IsVerified:      p.IsVerified,
		FollowerCount:   p.FollowerCount,
		FollowingCount:  p.FollowingCount,
		PostCount:       p.PostCount,
		IsFollowing:     isFollowing,
		IsFollowedBy:    isFollowedBy,
		CreatedAt:       p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
