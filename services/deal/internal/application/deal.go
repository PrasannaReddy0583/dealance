package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/deal/internal/domain/entity"
	"github.com/dealance/services/deal/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

type DealService struct {
	dealRepo        repository.DealRepository
	participantRepo repository.ParticipantRepository
	docRepo         repository.DocumentRepository
	milestoneRepo   repository.MilestoneRepository
	ndaRepo         repository.NDARepository
	negotiationRepo repository.NegotiationRepository
	cacheRepo       repository.CacheRepository
	log             zerolog.Logger
}

func NewDealService(
	dealRepo repository.DealRepository, participantRepo repository.ParticipantRepository,
	docRepo repository.DocumentRepository, milestoneRepo repository.MilestoneRepository,
	ndaRepo repository.NDARepository, negotiationRepo repository.NegotiationRepository,
	cacheRepo repository.CacheRepository, log zerolog.Logger,
) *DealService {
	return &DealService{
		dealRepo: dealRepo, participantRepo: participantRepo, docRepo: docRepo,
		milestoneRepo: milestoneRepo, ndaRepo: ndaRepo, negotiationRepo: negotiationRepo,
		cacheRepo: cacheRepo, log: log,
	}
}

func (s *DealService) CreateDeal(ctx context.Context, creatorID string, req entity.CreateDealRequest) (*entity.DealResponse, error) {
	cID, _ := uuid.Parse(creatorID)
	sID, err := uuid.Parse(req.StartupID)
	if err != nil { return nil, apperrors.ErrValidation("invalid startup_id") }

	requiresNDA := true
	if req.RequiresNDA != nil { requiresNDA = *req.RequiresNDA }

	deal := &entity.Deal{
		ID: uuid.New(), StartupID: sID, Title: req.Title, Description: req.Description,
		DealType: req.DealType, Status: entity.DealStatusDraft, AmountPaise: req.AmountPaise,
		MinTicketPaise: sql.NullInt64{Int64: req.MinTicketPaise, Valid: req.MinTicketPaise > 0},
		EquityPct: sql.NullFloat64{Float64: req.EquityPct, Valid: req.EquityPct > 0},
		ValuationPaise: sql.NullInt64{Int64: req.ValuationPaise, Valid: req.ValuationPaise > 0},
		TermsSummary: sql.NullString{String: req.TermsSummary, Valid: req.TermsSummary != ""},
		RequiresNDA: requiresNDA, RequiresKYC: true, Currency: "INR", CreatedBy: cID, MaxParticipants: 50,
	}
	if req.FundingRoundID != "" {
		frID, _ := uuid.Parse(req.FundingRoundID)
		deal.FundingRoundID = &frID
	}

	if err := s.dealRepo.Create(ctx, deal); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	return s.toDealResponse(deal), nil
}

func (s *DealService) GetDeal(ctx context.Context, dealID string) (*entity.DealResponse, error) {
	if cached, err := s.cacheRepo.GetCachedDeal(ctx, dealID); err == nil {
		return s.toDealResponse(cached), nil
	}
	dID, _ := uuid.Parse(dealID)
	deal, err := s.dealRepo.GetByID(ctx, dID)
	if err != nil { return nil, apperrors.ErrNotFound("Deal") }
	_ = s.cacheRepo.CacheDeal(ctx, deal)
	return s.toDealResponse(deal), nil
}

func (s *DealService) GetDealsByStartup(ctx context.Context, startupID string) ([]entity.DealListItem, error) {
	sID, _ := uuid.Parse(startupID)
	deals, err := s.dealRepo.GetByStartup(ctx, sID)
	if err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }
	return s.toDealList(deals), nil
}

func (s *DealService) GetMyDeals(ctx context.Context, creatorID string) ([]entity.DealListItem, error) {
	cID, _ := uuid.Parse(creatorID)
	deals, err := s.dealRepo.GetByCreator(ctx, cID)
	if err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }
	return s.toDealList(deals), nil
}

func (s *DealService) UpdateDeal(ctx context.Context, dealID, creatorID string, req entity.UpdateDealRequest) (*entity.DealResponse, error) {
	dID, _ := uuid.Parse(dealID)
	deal, err := s.dealRepo.GetByID(ctx, dID)
	if err != nil { return nil, apperrors.ErrNotFound("Deal") }
	if deal.CreatedBy.String() != creatorID {
		return nil, apperrors.ErrForbidden("NOT_CREATOR", "Only the deal creator can edit")
	}
	updates := make(map[string]interface{})
	if req.Title != nil { updates["title"] = *req.Title }
	if req.Description != nil { updates["description"] = *req.Description }
	if req.Status != nil { updates["status"] = *req.Status }
	if len(updates) == 0 { return nil, apperrors.ErrValidation("no fields to update") }
	if err := s.dealRepo.Update(ctx, dID, updates); err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }
	_ = s.cacheRepo.InvalidateDeal(ctx, dealID)
	updated, _ := s.dealRepo.GetByID(ctx, dID)
	return s.toDealResponse(updated), nil
}

// JoinDeal — express interest in a deal
func (s *DealService) JoinDeal(ctx context.Context, dealID, userID string, req entity.JoinDealRequest) error {
	dID, _ := uuid.Parse(dealID)
	uID, _ := uuid.Parse(userID)
	role := "INVESTOR"
	if req.Role != "" { role = req.Role }
	p := &entity.DealParticipant{ID: uuid.New(), DealID: dID, UserID: uID, Role: role, Status: entity.ParticipantInterested}
	return s.participantRepo.Create(ctx, p)
}

// CommitToDeal — commit investment amount
func (s *DealService) CommitToDeal(ctx context.Context, dealID, userID string, req entity.CommitRequest) error {
	dID, _ := uuid.Parse(dealID)
	uID, _ := uuid.Parse(userID)
	participant, err := s.participantRepo.GetByDealAndUser(ctx, dID, uID)
	if err != nil { return apperrors.ErrNotFound("Participant") }
	return s.participantRepo.Update(ctx, participant.ID, map[string]interface{}{
		"status": entity.ParticipantCommitted, "commitment_paise": req.AmountPaise,
		"committed_at": time.Now(),
	})
}

// SignNDA — sign NDA for a deal
func (s *DealService) SignNDA(ctx context.Context, dealID, userID string, req entity.SignNDARequest, ipAddress, userAgent string) error {
	dID, _ := uuid.Parse(dealID)
	uID, _ := uuid.Parse(userID)
	nda := &entity.DealNDA{ID: uuid.New(), DealID: dID, UserID: uID, Status: "PENDING"}
	_ = s.ndaRepo.Create(ctx, nda)
	existing, err := s.ndaRepo.GetByDealAndUser(ctx, dID, uID)
	if err != nil { return apperrors.ErrInternal().WithInternal(err) }
	now := time.Now()
	return s.ndaRepo.Update(ctx, existing.ID, map[string]interface{}{
		"status": "SIGNED", "signed_at": now,
		"ip_address": ipAddress, "user_agent": userAgent, "signature_hash": req.SignatureHash,
	})
}

// SendNegotiationMessage — send a message/offer in the deal room
func (s *DealService) SendNegotiationMessage(ctx context.Context, dealID, senderID string, req entity.NegotiationMessageRequest) error {
	dID, _ := uuid.Parse(dealID)
	sID, _ := uuid.Parse(senderID)
	neg := &entity.DealNegotiation{
		ID: uuid.New(), DealID: dID, SenderID: sID, MessageType: req.MessageType,
		Body: req.Body, Status: "ACTIVE",
		AmountPaise: sql.NullInt64{Int64: req.AmountPaise, Valid: req.AmountPaise > 0},
		EquityPct: sql.NullFloat64{Float64: req.EquityPct, Valid: req.EquityPct > 0},
	}
	if req.ParentID != "" { pID, _ := uuid.Parse(req.ParentID); neg.ParentID = &pID }
	return s.negotiationRepo.Create(ctx, neg)
}

func (s *DealService) GetNegotiations(ctx context.Context, dealID string, limit int) ([]entity.DealNegotiation, error) {
	dID, _ := uuid.Parse(dealID)
	if limit <= 0 || limit > 100 { limit = 50 }
	return s.negotiationRepo.GetByDeal(ctx, dID, limit)
}

func (s *DealService) GetParticipants(ctx context.Context, dealID string) ([]entity.DealParticipant, error) {
	dID, _ := uuid.Parse(dealID)
	return s.participantRepo.GetByDeal(ctx, dID)
}

func (s *DealService) GetDocuments(ctx context.Context, dealID string) ([]entity.DealDocument, error) {
	dID, _ := uuid.Parse(dealID)
	return s.docRepo.GetByDeal(ctx, dID)
}

func (s *DealService) UploadDocument(ctx context.Context, dealID, uploaderID string, req entity.UploadDocumentRequest) error {
	dID, _ := uuid.Parse(dealID)
	uID, _ := uuid.Parse(uploaderID)
	isConfidential := true
	if req.IsConfidential != nil { isConfidential = *req.IsConfidential }
	accessLevel := "NDA_SIGNED"
	if req.AccessLevel != "" { accessLevel = req.AccessLevel }
	doc := &entity.DealDocument{
		ID: uuid.New(), DealID: dID, UploadedBy: uID, DocType: req.DocType,
		Title: req.Title, FileURL: req.FileURL, IsConfidential: isConfidential, AccessLevel: accessLevel, Version: 1,
	}
	return s.docRepo.Create(ctx, doc)
}

func (s *DealService) GetMilestones(ctx context.Context, dealID string) ([]entity.DealMilestone, error) {
	dID, _ := uuid.Parse(dealID)
	return s.milestoneRepo.GetByDeal(ctx, dID)
}

func (s *DealService) toDealResponse(d *entity.Deal) *entity.DealResponse {
	resp := &entity.DealResponse{
		ID: d.ID.String(), StartupID: d.StartupID.String(), Title: d.Title, Description: d.Description,
		DealType: d.DealType, Status: d.Status, AmountPaise: d.AmountPaise,
		MaxParticipants: d.MaxParticipants, Currency: d.Currency,
		RequiresNDA: d.RequiresNDA, RequiresKYC: d.RequiresKYC,
		CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if d.MinTicketPaise.Valid { resp.MinTicketPaise = d.MinTicketPaise.Int64 }
	if d.EquityPct.Valid { resp.EquityPct = d.EquityPct.Float64 }
	if d.ValuationPaise.Valid { resp.ValuationPaise = d.ValuationPaise.Int64 }
	return resp
}

func (s *DealService) toDealList(deals []entity.Deal) []entity.DealListItem {
	items := make([]entity.DealListItem, len(deals))
	for i, d := range deals {
		items[i] = entity.DealListItem{
			ID: d.ID.String(), Title: d.Title, DealType: d.DealType,
			Status: d.Status, AmountPaise: d.AmountPaise, Currency: d.Currency,
			CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return items
}
