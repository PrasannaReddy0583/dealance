package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/dealance/services/content/internal/application"
	"github.com/dealance/services/content/internal/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/middleware"
	"github.com/dealance/shared/pkg/response"
)

var validate = validator.New()

// --- Post Handler ---

type PostHandler struct{ svc *application.PostService }

func NewPostHandler(svc *application.PostService) *PostHandler { return &PostHandler{svc: svc} }

func (h *PostHandler) Create(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	var req entity.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.CreatePost(c.Request.Context(), claims.Subject, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

func (h *PostHandler) Get(c *gin.Context) {
	postID := c.Param("id")
	viewerID := ""
	if claims := middleware.GetClaims(c); claims != nil {
		viewerID = claims.Subject
	}
	result, err := h.svc.GetPost(c.Request.Context(), postID, viewerID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *PostHandler) Update(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	postID := c.Param("id")
	var req entity.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.UpdatePost(c.Request.Context(), postID, claims.Subject, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *PostHandler) Delete(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	postID := c.Param("id")
	if err := h.svc.DeletePost(c.Request.Context(), postID, claims.Subject); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Post deleted"})
}

func (h *PostHandler) GetUserPosts(c *gin.Context) {
	userID := c.Param("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetUserPosts(c.Request.Context(), userID, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *PostHandler) GetFeed(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetFeed(c.Request.Context(), limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *PostHandler) GetByHashtag(c *gin.Context) {
	tag := c.Param("tag")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetByHashtag(c.Request.Context(), tag, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *PostHandler) Save(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	var req entity.SavePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.SavePost(c.Request.Context(), claims.Subject, req.PostID, req.Collection); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Post saved"})
}

func (h *PostHandler) Unsave(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	postID := c.Param("id")
	if err := h.svc.UnsavePost(c.Request.Context(), claims.Subject, postID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Post unsaved"})
}

func (h *PostHandler) TrendingHashtags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetTrendingHashtags(c.Request.Context(), limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// --- Comment Handler ---

type CommentHandler struct{ svc *application.CommentService }

func NewCommentHandler(svc *application.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

func (h *CommentHandler) Create(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	var req entity.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.CreateComment(c.Request.Context(), claims.Subject, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, result)
}

func (h *CommentHandler) GetByPost(c *gin.Context) {
	postID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetComments(c.Request.Context(), postID, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *CommentHandler) GetReplies(c *gin.Context) {
	commentID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetReplies(c.Request.Context(), commentID, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *CommentHandler) Update(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	commentID := c.Param("id")
	var req entity.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.UpdateComment(c.Request.Context(), commentID, claims.Subject, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *CommentHandler) Delete(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	commentID := c.Param("id")
	if err := h.svc.DeleteComment(c.Request.Context(), commentID, claims.Subject); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Comment deleted"})
}

// --- Reaction Handler ---

type ReactionHandler struct{ svc *application.ReactionService }

func NewReactionHandler(svc *application.ReactionService) *ReactionHandler {
	return &ReactionHandler{svc: svc}
}

func (h *ReactionHandler) React(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	var req entity.ReactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.React(c.Request.Context(), claims.Subject, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Reaction added"})
}

func (h *ReactionHandler) Unreact(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	var req entity.UnreactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.Unreact(c.Request.Context(), claims.Subject, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Reaction removed"})
}

func (h *ReactionHandler) Report(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrValidation("authentication required"))
		return
	}
	var req entity.ReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.ReportContent(c.Request.Context(), claims.Subject, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Report submitted"})
}

// --- Health ---

func HealthCheck(c *gin.Context) {
	response.OK(c, map[string]string{"status": "healthy", "service": "dealance-content"})
}
