package application

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog"

	"github.com/dealance/services/startup/internal/domain/entity"
	"github.com/dealance/services/startup/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

type StartupService struct {
	startupRepo  repository.StartupRepository
	fundingRepo  repository.FundingRoundRepository
	teamRepo     repository.TeamMemberRepository
	followRepo   repository.StartupFollowRepository
	cacheRepo    repository.CacheRepository
	log          zerolog.Logger
}

func NewStartupService(
	startupRepo repository.StartupRepository,
	fundingRepo repository.FundingRoundRepository,
	teamRepo repository.TeamMemberRepository,
	followRepo repository.StartupFollowRepository,
	cacheRepo repository.CacheRepository,
	log zerolog.Logger,
) *StartupService {
	return &StartupService{
		startupRepo: startupRepo, fundingRepo: fundingRepo, teamRepo: teamRepo,
		followRepo: followRepo, cacheRepo: cacheRepo, log: log,
	}
}

func (s *StartupService) CreateStartup(ctx context.Context, founderID string, req entity.CreateStartupRequest) (*entity.StartupResponse, error) {
	fID, err := uuid.Parse(founderID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid founder ID")
	}

	stage := req.Stage
	if stage == "" {
		stage = entity.StageIdea
	}

	slug := slugRegex.ReplaceAllString(strings.ToLower(req.Name), "-")
	slug = strings.Trim(slug, "-")

	startup := &entity.Startup{
		ID:        uuid.New(),
		FounderID: fID,
		Name:      req.Name,
		Slug:      slug + "-" + uuid.New().String()[:8],
		Tagline:   sql.NullString{String: req.Tagline, Valid: req.Tagline != ""},
		Description: req.Description,
		Sector:    req.Sector,
		Stage:     stage,
		BusinessModel: sql.NullString{String: req.BusinessModel, Valid: req.BusinessModel != ""},
		Website:   sql.NullString{String: req.Website, Valid: req.Website != ""},
		FoundedYear: sql.NullInt32{Int32: int32(req.FoundedYear), Valid: req.FoundedYear > 0},
		Headquarters: sql.NullString{String: req.Headquarters, Valid: req.Headquarters != ""},
		Country:   sql.NullString{String: req.Country, Valid: req.Country != ""},
		IncorporationType: sql.NullString{String: req.IncorporationType, Valid: req.IncorporationType != ""},
		Tags:      pq.StringArray(req.Tags),
		Status:    "ACTIVE",
	}

	if err := s.startupRepo.Create(ctx, startup); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	return s.toStartupResponse(startup, false), nil
}

func (s *StartupService) GetStartup(ctx context.Context, startupID, viewerID string) (*entity.StartupResponse, error) {
	sID, err := uuid.Parse(startupID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid startup ID")
	}

	if cached, err := s.cacheRepo.GetCachedStartup(ctx, startupID); err == nil {
		isFollowing := false
		if viewerID != "" {
			vID, _ := uuid.Parse(viewerID)
			isFollowing, _ = s.followRepo.IsFollowing(ctx, vID, sID)
		}
		_ = s.startupRepo.IncrementCounter(ctx, sID, "view_count", 1)
		return s.toStartupResponse(cached, isFollowing), nil
	}

	startup, err := s.startupRepo.GetByID(ctx, sID)
	if err != nil {
		return nil, apperrors.ErrNotFound("Startup")
	}
	_ = s.cacheRepo.CacheStartup(ctx, startup)
	_ = s.startupRepo.IncrementCounter(ctx, sID, "view_count", 1)

	isFollowing := false
	if viewerID != "" {
		vID, _ := uuid.Parse(viewerID)
		isFollowing, _ = s.followRepo.IsFollowing(ctx, vID, sID)
	}
	return s.toStartupResponse(startup, isFollowing), nil
}

func (s *StartupService) GetBySlug(ctx context.Context, slug string) (*entity.StartupResponse, error) {
	startup, err := s.startupRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperrors.ErrNotFound("Startup")
	}
	return s.toStartupResponse(startup, false), nil
}

func (s *StartupService) UpdateStartup(ctx context.Context, startupID, founderID string, req entity.UpdateStartupRequest) (*entity.StartupResponse, error) {
	sID, err := uuid.Parse(startupID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid startup ID")
	}
	startup, err := s.startupRepo.GetByID(ctx, sID)
	if err != nil {
		return nil, apperrors.ErrNotFound("Startup")
	}
	if startup.FounderID.String() != founderID {
		return nil, apperrors.ErrForbidden("NOT_FOUNDER", "Only the founder can edit this startup")
	}

	updates := make(map[string]interface{})
	if req.Name != nil { updates["name"] = *req.Name }
	if req.Tagline != nil { updates["tagline"] = *req.Tagline }
	if req.Description != nil { updates["description"] = *req.Description }
	if req.LogoURL != nil { updates["logo_url"] = *req.LogoURL }
	if req.CoverURL != nil { updates["cover_url"] = *req.CoverURL }
	if req.Sector != nil { updates["sector"] = *req.Sector }
	if req.Stage != nil { updates["stage"] = *req.Stage }
	if req.BusinessModel != nil { updates["business_model"] = *req.BusinessModel }
	if req.Website != nil { updates["website"] = *req.Website }
	if req.TeamSize != nil { updates["team_size"] = *req.TeamSize }
	if req.IncorporationType != nil { updates["incorporation_type"] = *req.IncorporationType }
	if req.CINNumber != nil { updates["cin_number"] = *req.CINNumber }
	if req.GSTIN != nil { updates["gstin"] = *req.GSTIN }
	if req.Tags != nil { updates["tags"] = pq.StringArray(req.Tags) }

	if len(updates) == 0 {
		return nil, apperrors.ErrValidation("no fields to update")
	}

	if err := s.startupRepo.Update(ctx, sID, updates); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	_ = s.cacheRepo.InvalidateStartup(ctx, startupID)

	updated, _ := s.startupRepo.GetByID(ctx, sID)
	return s.toStartupResponse(updated, false), nil
}

func (s *StartupService) GetMyStartups(ctx context.Context, founderID string) ([]entity.StartupListItem, error) {
	fID, err := uuid.Parse(founderID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid founder ID")
	}
	startups, err := s.startupRepo.GetByFounder(ctx, fID)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	return s.toStartupList(startups), nil
}

func (s *StartupService) SearchStartups(ctx context.Context, req entity.SearchStartupsRequest) ([]entity.StartupListItem, error) {
	limit := req.Limit
	if limit <= 0 || limit > 50 { limit = 20 }
	startups, err := s.startupRepo.Search(ctx, req.Query, req.Sector, req.Stage, req.Country, limit, nil)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	return s.toStartupList(startups), nil
}

func (s *StartupService) FollowStartup(ctx context.Context, userID, startupID string) error {
	uID, _ := uuid.Parse(userID)
	sID, _ := uuid.Parse(startupID)
	already, _ := s.followRepo.IsFollowing(ctx, uID, sID)
	if already { return nil }
	if err := s.followRepo.Follow(ctx, uID, sID); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}
	_ = s.startupRepo.IncrementCounter(ctx, sID, "follower_count", 1)
	_ = s.cacheRepo.InvalidateStartup(ctx, startupID)
	return nil
}

func (s *StartupService) UnfollowStartup(ctx context.Context, userID, startupID string) error {
	uID, _ := uuid.Parse(userID)
	sID, _ := uuid.Parse(startupID)
	following, _ := s.followRepo.IsFollowing(ctx, uID, sID)
	if !following { return nil }
	if err := s.followRepo.Unfollow(ctx, uID, sID); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}
	_ = s.startupRepo.IncrementCounter(ctx, sID, "follower_count", -1)
	_ = s.cacheRepo.InvalidateStartup(ctx, startupID)
	return nil
}

func (s *StartupService) CreateFundingRound(ctx context.Context, startupID, founderID string, req entity.CreateFundingRoundRequest) error {
	sID, _ := uuid.Parse(startupID)
	startup, err := s.startupRepo.GetByID(ctx, sID)
	if err != nil {
		return apperrors.ErrNotFound("Startup")
	}
	if startup.FounderID.String() != founderID {
		return apperrors.ErrForbidden("NOT_FOUNDER", "Only the founder can add funding rounds")
	}
	round := &entity.FundingRound{
		ID: uuid.New(), StartupID: sID, RoundType: req.RoundType, AmountPaise: req.AmountPaise,
		ValuationPaise: sql.NullInt64{Int64: req.ValuationPaise, Valid: req.ValuationPaise > 0},
		TargetPaise: sql.NullInt64{Int64: req.TargetPaise, Valid: req.TargetPaise > 0},
		MinTicketPaise: sql.NullInt64{Int64: req.MinTicketPaise, Valid: req.MinTicketPaise > 0},
		EquityOffered: sql.NullFloat64{Float64: req.EquityOffered, Valid: req.EquityOffered > 0},
		InstrumentType: sql.NullString{String: req.InstrumentType, Valid: req.InstrumentType != ""},
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Currency: "INR", Status: "OPEN",
	}
	return s.fundingRepo.Create(ctx, round)
}

func (s *StartupService) GetFundingRounds(ctx context.Context, startupID string) ([]entity.FundingRound, error) {
	sID, _ := uuid.Parse(startupID)
	return s.fundingRepo.GetByStartup(ctx, sID)
}

func (s *StartupService) AddTeamMember(ctx context.Context, startupID, founderID string, req entity.AddTeamMemberRequest) error {
	sID, _ := uuid.Parse(startupID)
	startup, err := s.startupRepo.GetByID(ctx, sID)
	if err != nil { return apperrors.ErrNotFound("Startup") }
	if startup.FounderID.String() != founderID {
		return apperrors.ErrForbidden("NOT_FOUNDER", "Only the founder can manage team")
	}
	member := &entity.TeamMember{
		ID: uuid.New(), StartupID: sID, Name: req.Name, Role: req.Role,
		Title: sql.NullString{String: req.Title, Valid: req.Title != ""},
		Bio: req.Bio, LinkedInURL: sql.NullString{String: req.LinkedInURL, Valid: req.LinkedInURL != ""},
		IsFounder: req.IsFounder,
	}
	return s.teamRepo.Create(ctx, member)
}

func (s *StartupService) GetTeamMembers(ctx context.Context, startupID string) ([]entity.TeamMember, error) {
	sID, _ := uuid.Parse(startupID)
	return s.teamRepo.GetByStartup(ctx, sID)
}

func (s *StartupService) toStartupResponse(st *entity.Startup, isFollowing bool) *entity.StartupResponse {
	resp := &entity.StartupResponse{
		ID: st.ID.String(), FounderID: st.FounderID.String(), Name: st.Name, Slug: st.Slug,
		Description: st.Description, LogoURL: st.LogoURL, CoverURL: st.CoverURL,
		Sector: st.Sector, Stage: st.Stage, TeamSize: st.TeamSize, Status: st.Status,
		IsVerified: st.IsVerified, ViewCount: st.ViewCount, FollowerCount: st.FollowerCount,
		IsFollowing: isFollowing, CreatedAt: st.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if st.Tagline.Valid { resp.Tagline = st.Tagline.String }
	if st.Website.Valid { resp.Website = st.Website.String }
	if st.BusinessModel.Valid { resp.BusinessModel = st.BusinessModel.String }
	return resp
}

func (s *StartupService) toStartupList(startups []entity.Startup) []entity.StartupListItem {
	items := make([]entity.StartupListItem, len(startups))
	for i, st := range startups {
		items[i] = entity.StartupListItem{
			ID: st.ID.String(), Name: st.Name, Slug: st.Slug, LogoURL: st.LogoURL,
			Sector: st.Sector, Stage: st.Stage, IsVerified: st.IsVerified,
		}
		if st.Tagline.Valid { items[i].Tagline = st.Tagline.String }
	}
	return items
}
