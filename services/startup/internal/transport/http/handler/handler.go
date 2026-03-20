package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/dealance/services/startup/internal/application"
	"github.com/dealance/services/startup/internal/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/middleware"
	"github.com/dealance/shared/pkg/response"
)

var validate = validator.New()

type StartupHandler struct{ svc *application.StartupService }

func NewStartupHandler(svc *application.StartupService) *StartupHandler {
	return &StartupHandler{svc: svc}
}

func (h *StartupHandler) Create(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.CreateStartupRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := validate.Struct(req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	result, err := h.svc.CreateStartup(c.Request.Context(), claims.Subject, req)
	if err != nil { response.Error(c, err); return }
	response.Created(c, result)
}

func (h *StartupHandler) Get(c *gin.Context) {
	id := c.Param("id")
	viewerID := ""
	if claims := middleware.GetClaims(c); claims != nil { viewerID = claims.Subject }
	result, err := h.svc.GetStartup(c.Request.Context(), id, viewerID)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *StartupHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	result, err := h.svc.GetBySlug(c.Request.Context(), slug)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *StartupHandler) Update(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	id := c.Param("id")
	var req entity.UpdateStartupRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	result, err := h.svc.UpdateStartup(c.Request.Context(), id, claims.Subject, req)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *StartupHandler) GetMyStartups(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	result, err := h.svc.GetMyStartups(c.Request.Context(), claims.Subject)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *StartupHandler) Search(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	req := entity.SearchStartupsRequest{
		Query: c.Query("q"), Sector: c.Query("sector"),
		Stage: c.Query("stage"), Country: c.Query("country"), Limit: limit,
	}
	result, err := h.svc.SearchStartups(c.Request.Context(), req)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *StartupHandler) Follow(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	id := c.Param("id")
	if err := h.svc.FollowStartup(c.Request.Context(), claims.Subject, id); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Following startup"})
}

func (h *StartupHandler) Unfollow(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	id := c.Param("id")
	if err := h.svc.UnfollowStartup(c.Request.Context(), claims.Subject, id); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Unfollowed startup"})
}

func (h *StartupHandler) GetFundingRounds(c *gin.Context) {
	id := c.Param("id")
	result, err := h.svc.GetFundingRounds(c.Request.Context(), id)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *StartupHandler) CreateFundingRound(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	id := c.Param("id")
	var req entity.CreateFundingRoundRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.CreateFundingRound(c.Request.Context(), id, claims.Subject, req); err != nil { response.Error(c, err); return }
	response.Created(c, map[string]string{"message": "Funding round created"})
}

func (h *StartupHandler) GetTeam(c *gin.Context) {
	id := c.Param("id")
	result, err := h.svc.GetTeamMembers(c.Request.Context(), id)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *StartupHandler) AddTeamMember(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	id := c.Param("id")
	var req entity.AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.AddTeamMember(c.Request.Context(), id, claims.Subject, req); err != nil { response.Error(c, err); return }
	response.Created(c, map[string]string{"message": "Team member added"})
}

func HealthCheck(c *gin.Context) {
	response.OK(c, map[string]string{"status": "healthy", "service": "dealance-startup"})
}
