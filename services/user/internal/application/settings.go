package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/user/internal/domain/entity"
	"github.com/dealance/services/user/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

// SettingsService handles user settings.
type SettingsService struct {
	settingsRepo repository.SettingsRepository
	log          zerolog.Logger
}

func NewSettingsService(settingsRepo repository.SettingsRepository, log zerolog.Logger) *SettingsService {
	return &SettingsService{settingsRepo: settingsRepo, log: log}
}

// GetSettings returns the user's settings.
func (s *SettingsService) GetSettings(ctx context.Context, userID string) (*entity.UserSettings, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid user ID")
	}

	settings, err := s.settingsRepo.GetByUserID(ctx, id)
	if err != nil {
		// Create defaults if missing
		_ = s.settingsRepo.Create(ctx, id)
		settings, err = s.settingsRepo.GetByUserID(ctx, id)
		if err != nil {
			return nil, apperrors.ErrInternal().WithInternal(err)
		}
	}

	return settings, nil
}

// UpdateSettings partially updates user settings.
func (s *SettingsService) UpdateSettings(ctx context.Context, userID string, req entity.UpdateSettingsRequest) (*entity.UserSettings, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid user ID")
	}

	updates := make(map[string]interface{})
	if req.NotificationPush != nil {
		updates["notification_push"] = *req.NotificationPush
	}
	if req.NotificationEmail != nil {
		updates["notification_email"] = *req.NotificationEmail
	}
	if req.NotificationSMS != nil {
		updates["notification_sms"] = *req.NotificationSMS
	}
	if req.NotificationDealUpdates != nil {
		updates["notification_deal_updates"] = *req.NotificationDealUpdates
	}
	if req.NotificationNewFollowers != nil {
		updates["notification_new_followers"] = *req.NotificationNewFollowers
	}
	if req.NotificationMessages != nil {
		updates["notification_messages"] = *req.NotificationMessages
	}
	if req.PrivacyShowEmail != nil {
		updates["privacy_show_email"] = *req.PrivacyShowEmail
	}
	if req.PrivacyShowPhone != nil {
		updates["privacy_show_phone"] = *req.PrivacyShowPhone
	}
	if req.PrivacyShowLocation != nil {
		updates["privacy_show_location"] = *req.PrivacyShowLocation
	}
	if req.PrivacyAllowMessages != nil {
		updates["privacy_allow_messages"] = *req.PrivacyAllowMessages
	}
	if req.PrivacyShowInvestments != nil {
		updates["privacy_show_investments"] = *req.PrivacyShowInvestments
	}
	if req.FeedContentLanguage != nil {
		updates["feed_content_language"] = *req.FeedContentLanguage
	}
	if req.FeedSortPreference != nil {
		updates["feed_sort_preference"] = *req.FeedSortPreference
	}
	if req.Theme != nil {
		updates["theme"] = *req.Theme
	}

	if len(updates) == 0 {
		return nil, apperrors.ErrValidation("no fields to update")
	}

	if err := s.settingsRepo.Update(ctx, id, updates); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	return s.settingsRepo.GetByUserID(ctx, id)
}
