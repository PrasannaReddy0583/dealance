package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/dealance/services/user/internal/application"
	"github.com/dealance/services/user/internal/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/middleware"
	"github.com/dealance/shared/pkg/response"
)

var validate = validator.New()

// --- Profile Handler ---

type ProfileHandler struct {
	svc *application.ProfileService
}

func NewProfileHandler(svc *application.ProfileService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

func (h *ProfileHandler) Create(c *gin.Context) {
	var req entity.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.CreateProfile(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

func (h *ProfileHandler) GetByID(c *gin.Context) {
	profileID := c.Param("id")
	viewerID := ""
	if claims := middleware.GetClaims(c); claims != nil {
		viewerID = claims.Subject
	}
	result, err := h.svc.GetProfile(c.Request.Context(), profileID, viewerID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *ProfileHandler) GetByUsername(c *gin.Context) {
	username := c.Param("username")
	result, err := h.svc.GetProfileByUsername(c.Request.Context(), username)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *ProfileHandler) GetMe(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}
	result, err := h.svc.GetProfile(c.Request.Context(), claims.Subject, claims.Subject)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *ProfileHandler) Update(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}
	var req entity.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.UpdateProfile(c.Request.Context(), claims.Subject, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *ProfileHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.Error(c, apperrors.ErrValidation("query parameter 'q' is required"))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	req := entity.SearchUsersRequest{Query: q, Limit: limit}
	result, err := h.svc.SearchProfiles(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// --- Follow Handler ---

type FollowHandler struct {
	svc *application.FollowService
}

func NewFollowHandler(svc *application.FollowService) *FollowHandler {
	return &FollowHandler{svc: svc}
}

func (h *FollowHandler) Follow(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}
	var req entity.FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.Follow(c.Request.Context(), claims.Subject, req.TargetUserID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Followed successfully"})
}

func (h *FollowHandler) Unfollow(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}
	var req entity.FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.Unfollow(c.Request.Context(), claims.Subject, req.TargetUserID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Unfollowed successfully"})
}

func (h *FollowHandler) GetFollowers(c *gin.Context) {
	userID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetFollowers(c.Request.Context(), userID, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *FollowHandler) GetFollowing(c *gin.Context) {
	userID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetFollowing(c.Request.Context(), userID, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *FollowHandler) GetCounts(c *gin.Context) {
	userID := c.Param("id")
	result, err := h.svc.GetFollowCounts(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *FollowHandler) Block(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}
	var req entity.BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.BlockUser(c.Request.Context(), claims.Subject, req.TargetUserID, req.Reason); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "User blocked"})
}

func (h *FollowHandler) Unblock(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}
	targetID := c.Param("id")
	if err := h.svc.UnblockUser(c.Request.Context(), claims.Subject, targetID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "User unblocked"})
}

// --- Settings Handler ---

type SettingsHandler struct {
	svc *application.SettingsService
}

func NewSettingsHandler(svc *application.SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

func (h *SettingsHandler) Get(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}
	result, err := h.svc.GetSettings(c.Request.Context(), claims.Subject)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *SettingsHandler) Update(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}
	var req entity.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.UpdateSettings(c.Request.Context(), claims.Subject, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// --- Health ---

func HealthCheck(c *gin.Context) {
	response.OK(c, map[string]string{"status": "healthy", "service": "dealance-user"})
}
