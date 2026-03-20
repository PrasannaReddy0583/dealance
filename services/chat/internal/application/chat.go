package application

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/chat/internal/domain/entity"
	"github.com/dealance/services/chat/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

type ChatService struct {
	convRepo        repository.ConversationRepository
	participantRepo repository.ParticipantRepository
	messageRepo     repository.MessageRepository
	receiptRepo     repository.ReadReceiptRepository
	cacheRepo       repository.CacheRepository
	log             zerolog.Logger
}

func NewChatService(
	convRepo repository.ConversationRepository, participantRepo repository.ParticipantRepository,
	messageRepo repository.MessageRepository, receiptRepo repository.ReadReceiptRepository,
	cacheRepo repository.CacheRepository, log zerolog.Logger,
) *ChatService {
	return &ChatService{convRepo: convRepo, participantRepo: participantRepo, messageRepo: messageRepo, receiptRepo: receiptRepo, cacheRepo: cacheRepo, log: log}
}

func (s *ChatService) CreateConversation(ctx context.Context, creatorID string, req entity.CreateConversationRequest) (*entity.ConversationResponse, error) {
	cID, _ := uuid.Parse(creatorID)
	conv := &entity.Conversation{
		ID: uuid.New(), ConvType: req.ConvType, CreatedBy: cID,
		Title: sql.NullString{String: req.Title, Valid: req.Title != ""},
	}
	if req.DealID != "" { dID, _ := uuid.Parse(req.DealID); conv.DealID = &dID }

	if err := s.convRepo.Create(ctx, conv); err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }

	// Add creator
	_ = s.participantRepo.Add(ctx, &entity.ConversationParticipant{ConversationID: conv.ID, UserID: cID, Role: "ADMIN"})
	// Add participants
	for _, pid := range req.ParticipantIDs {
		pUID, _ := uuid.Parse(pid)
		_ = s.participantRepo.Add(ctx, &entity.ConversationParticipant{ConversationID: conv.ID, UserID: pUID, Role: "MEMBER"})
	}

	return &entity.ConversationResponse{ID: conv.ID.String(), ConvType: conv.ConvType, Title: req.Title}, nil
}

func (s *ChatService) GetConversations(ctx context.Context, userID string, limit int) ([]entity.ConversationResponse, error) {
	uID, _ := uuid.Parse(userID)
	if limit <= 0 || limit > 50 { limit = 20 }
	convs, err := s.convRepo.GetByUser(ctx, uID, limit)
	if err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }

	participants, _ := s.participantRepo.GetByUser(ctx, uID)
	unreadMap := make(map[string]int)
	for _, p := range participants { unreadMap[p.ConversationID.String()] = p.UnreadCount }

	items := make([]entity.ConversationResponse, len(convs))
	for i, c := range convs {
		items[i] = entity.ConversationResponse{
			ID: c.ID.String(), ConvType: c.ConvType, LastMessageAt: c.LastMessageAt.Format("2006-01-02T15:04:05Z"),
			UnreadCount: unreadMap[c.ID.String()],
		}
		if c.Title.Valid { items[i].Title = c.Title.String }
	}
	return items, nil
}

func (s *ChatService) SendMessage(ctx context.Context, senderID string, req entity.SendMessageRequest) (*entity.MessageResponse, error) {
	uID, _ := uuid.Parse(senderID)
	convID, _ := uuid.Parse(req.ConversationID)

	// Check participant
	isMember, _ := s.participantRepo.IsParticipant(ctx, convID, uID)
	if !isMember { return nil, apperrors.ErrForbidden("NOT_PARTICIPANT", "You are not in this conversation") }

	msgType := "TEXT"
	if req.MessageType != "" { msgType = req.MessageType }

	msg := &entity.Message{
		ID: uuid.New(), ConversationID: convID, SenderID: uID, MessageType: msgType, Body: req.Body,
		MediaURL: sql.NullString{String: req.MediaURL, Valid: req.MediaURL != ""},
	}
	if req.ReplyToID != "" { rID, _ := uuid.Parse(req.ReplyToID); msg.ReplyToID = &rID }

	if err := s.messageRepo.Create(ctx, msg); err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }

	// Update conversation and unread counts
	_ = s.convRepo.UpdateLastMessage(ctx, convID)
	_ = s.participantRepo.IncrementUnread(ctx, convID, uID)

	// Publish via Redis pub/sub
	resp := &entity.MessageResponse{
		ID: msg.ID.String(), SenderID: senderID, MessageType: msgType, Body: req.Body,
		MediaURL: req.MediaURL, CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if payload, err := json.Marshal(entity.WSMessage{Type: "MESSAGE", Payload: resp}); err == nil {
		_ = s.cacheRepo.PublishMessage(ctx, req.ConversationID, payload)
	}

	return resp, nil
}

func (s *ChatService) GetMessages(ctx context.Context, userID, convID string, limit int) ([]entity.MessageResponse, error) {
	uID, _ := uuid.Parse(userID)
	cID, _ := uuid.Parse(convID)
	isMember, _ := s.participantRepo.IsParticipant(ctx, cID, uID)
	if !isMember { return nil, apperrors.ErrForbidden("NOT_PARTICIPANT", "Not in this conversation") }
	if limit <= 0 || limit > 100 { limit = 50 }

	msgs, err := s.messageRepo.GetByConversation(ctx, cID, limit)
	if err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }

	items := make([]entity.MessageResponse, len(msgs))
	for i, m := range msgs {
		items[i] = entity.MessageResponse{
			ID: m.ID.String(), SenderID: m.SenderID.String(), MessageType: m.MessageType,
			Body: m.Body, IsEdited: m.IsEdited, CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if m.MediaURL.Valid { items[i].MediaURL = m.MediaURL.String }
		if m.ReplyToID != nil { items[i].ReplyToID = m.ReplyToID.String() }
	}
	return items, nil
}

func (s *ChatService) EditMessage(ctx context.Context, userID, msgID string, req entity.EditMessageRequest) error {
	uID, _ := uuid.Parse(userID)
	mID, _ := uuid.Parse(msgID)
	msg, err := s.messageRepo.GetByID(ctx, mID)
	if err != nil { return apperrors.ErrNotFound("Message") }
	if msg.SenderID != uID { return apperrors.ErrForbidden("NOT_SENDER", "Only the sender can edit") }
	return s.messageRepo.Update(ctx, mID, req.Body)
}

func (s *ChatService) DeleteMessage(ctx context.Context, userID, msgID string) error {
	uID, _ := uuid.Parse(userID)
	mID, _ := uuid.Parse(msgID)
	msg, err := s.messageRepo.GetByID(ctx, mID)
	if err != nil { return apperrors.ErrNotFound("Message") }
	if msg.SenderID != uID { return apperrors.ErrForbidden("NOT_SENDER", "Only the sender can delete") }
	return s.messageRepo.SoftDelete(ctx, mID)
}

func (s *ChatService) MarkRead(ctx context.Context, userID, convID string) error {
	uID, _ := uuid.Parse(userID)
	cID, _ := uuid.Parse(convID)
	return s.participantRepo.ResetUnread(ctx, cID, uID)
}
