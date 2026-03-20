package application

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/notify/internal/domain/entity"
	"github.com/dealance/services/notify/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

type NotifyService struct {
	notifRepo repository.NotificationRepository
	deviceRepo repository.DeviceTokenRepository
	prefRepo  repository.PreferencesRepository
	log       zerolog.Logger
}

func NewNotifyService(notifRepo repository.NotificationRepository, deviceRepo repository.DeviceTokenRepository, prefRepo repository.PreferencesRepository, log zerolog.Logger) *NotifyService {
	return &NotifyService{notifRepo: notifRepo, deviceRepo: deviceRepo, prefRepo: prefRepo, log: log}
}

func (s *NotifyService) Send(ctx context.Context, req entity.SendNotificationRequest) error {
	uID, _ := uuid.Parse(req.UserID)
	channel := "IN_APP"
	if req.Channel != "" { channel = req.Channel }
	notif := &entity.Notification{
		ID: uuid.New(), UserID: uID, NotifType: req.NotifType, Title: req.Title, Body: req.Body,
		Channel: channel, Status: "SENT",
	}
	if req.EntityType != "" { notif.EntityType = sql.NullString{String: req.EntityType, Valid: true} }
	if req.EntityID != "" { eID, _ := uuid.Parse(req.EntityID); notif.EntityID = &eID }
	return s.notifRepo.Create(ctx, notif)
}

func (s *NotifyService) GetNotifications(ctx context.Context, userID string, limit int) ([]entity.Notification, error) {
	uID, _ := uuid.Parse(userID)
	if limit <= 0 || limit > 50 { limit = 20 }
	return s.notifRepo.GetByUser(ctx, uID, limit)
}

func (s *NotifyService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	uID, _ := uuid.Parse(userID)
	return s.notifRepo.CountUnread(ctx, uID)
}

func (s *NotifyService) MarkRead(ctx context.Context, notifID string) error {
	nID, _ := uuid.Parse(notifID)
	return s.notifRepo.MarkRead(ctx, nID)
}

func (s *NotifyService) MarkAllRead(ctx context.Context, userID string) error {
	uID, _ := uuid.Parse(userID)
	return s.notifRepo.MarkAllRead(ctx, uID)
}

func (s *NotifyService) RegisterDevice(ctx context.Context, userID string, req entity.RegisterDeviceRequest) error {
	uID, _ := uuid.Parse(userID)
	return s.deviceRepo.Upsert(ctx, &entity.DeviceToken{ID: uuid.New(), UserID: uID, DeviceToken: req.DeviceToken, Platform: req.Platform, IsActive: true})
}

func (s *NotifyService) GetPreferences(ctx context.Context, userID string) (*entity.NotificationPreferences, error) {
	uID, _ := uuid.Parse(userID)
	prefs, err := s.prefRepo.GetByUser(ctx, uID)
	if err != nil { return &entity.NotificationPreferences{UserID: uID, PushEnabled: true, EmailEnabled: true, DealUpdates: true, ChatMessages: true, ContentReactions: true, NewFollowers: true}, nil }
	return prefs, nil
}

func (s *NotifyService) UpdatePreferences(ctx context.Context, userID string, req entity.UpdatePreferencesRequest) error {
	uID, _ := uuid.Parse(userID)
	existing, _ := s.GetPreferences(ctx, userID)
	if existing == nil { return apperrors.ErrInternal() }
	if req.PushEnabled != nil { existing.PushEnabled = *req.PushEnabled }
	if req.EmailEnabled != nil { existing.EmailEnabled = *req.EmailEnabled }
	if req.DealUpdates != nil { existing.DealUpdates = *req.DealUpdates }
	if req.ChatMessages != nil { existing.ChatMessages = *req.ChatMessages }
	if req.ContentReactions != nil { existing.ContentReactions = *req.ContentReactions }
	if req.NewFollowers != nil { existing.NewFollowers = *req.NewFollowers }
	if req.Marketing != nil { existing.Marketing = *req.Marketing }
	existing.UserID = uID
	return s.prefRepo.Upsert(ctx, existing)
}
