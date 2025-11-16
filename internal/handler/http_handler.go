package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/T1mof/pr-reviewer-service/internal/domain"
	"github.com/T1mof/pr-reviewer-service/internal/middleware"
	"github.com/T1mof/pr-reviewer-service/internal/service"
)

type Handler struct {
	service    service.ServiceInterface
	adminToken string
}

func NewHandler(svc service.ServiceInterface, adminToken string) *Handler {
	return &Handler{
		service:    svc,
		adminToken: adminToken,
	}
}

// ErrorResponse структура ответа с ошибкой согласно OpenAPI спецификации.
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// sendError отправляет структурированную ошибку клиенту и логирует её.
func (h *Handler) sendError(c *gin.Context, statusCode int, code, message string) {
	slog.Error("Request error",
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"status", statusCode,
		"error_code", code,
		"message", message,
	)

	var resp ErrorResponse
	resp.Error.Code = code
	resp.Error.Message = message
	c.JSON(statusCode, resp)
}

// CreateTeam обрабатывает POST /team/add.
func (h *Handler) CreateTeam(c *gin.Context) {
	var req struct {
		TeamName string `json:"team_name" binding:"required"`
		Members  []struct {
			UserID   string `json:"user_id" binding:"required"`
			Username string `json:"username" binding:"required"`
			IsActive bool   `json:"is_active"`
		} `json:"members" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	team := &domain.Team{
		TeamName: req.TeamName,
		Members:  make([]domain.TeamMember, len(req.Members)),
	}

	for i, m := range req.Members {
		userID, err := uuid.Parse(m.UserID)
		if err != nil {
			h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST",
				"invalid user_id UUID: "+m.UserID)
			return
		}

		team.Members[i] = domain.TeamMember{
			UserID:   userID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	err := h.service.CreateTeam(c.Request.Context(), team)
	if err != nil {
		if err.Error() == "TEAM_EXISTS" {
			h.sendError(c, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
			return
		}
		h.sendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	slog.Info("Team created successfully", "team_name", req.TeamName, "members_count", len(req.Members))
	c.JSON(http.StatusCreated, gin.H{"team": team})
}

// GetTeam обрабатывает GET /team/get?team_name=...
func (h *Handler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	team, err := h.service.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		if err.Error() == "TEAM_NOT_FOUND" {
			h.sendError(c, http.StatusNotFound, "NOT_FOUND", "team not found")
			return
		}
		h.sendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, team)
}

// SetUserActive обрабатывает POST /users/setIsActive.
func (h *Handler) SetUserActive(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid user_id UUID")
		return
	}

	user, err := h.service.SetUserActive(c.Request.Context(), userID, req.IsActive)
	if err != nil {
		if err.Error() == "USER_NOT_FOUND" {
			h.sendError(c, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		h.sendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	slog.Info("User activity changed", "user_id", userID, "is_active", req.IsActive)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// CreatePR обрабатывает POST /pullRequest/create.
func (h *Handler) CreatePR(c *gin.Context) {
	var req struct {
		PullRequestID   string `json:"pull_request_id" binding:"required"`
		PullRequestName string `json:"pull_request_name" binding:"required"`
		AuthorID        string `json:"author_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	prID, err := uuid.Parse(req.PullRequestID)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid pull_request_id UUID")
		return
	}

	authorID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid author_id UUID")
		return
	}

	pr, err := h.service.CreatePR(c.Request.Context(), prID, req.PullRequestName, authorID)
	if err != nil {
		if err.Error() == "PR_EXISTS" {
			h.sendError(c, http.StatusConflict, "PR_EXISTS", "PR id already exists")
			return
		}
		if err.Error() == "USER_NOT_FOUND" {
			h.sendError(c, http.StatusNotFound, "NOT_FOUND", "author or team not found")
			return
		}
		h.sendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	slog.Info("PR created", "pr_id", prID, "author_id", authorID, "reviewers_count", len(pr.AssignedReviewers))
	c.JSON(http.StatusCreated, gin.H{"pr": pr})
}

// MergePR обрабатывает POST /pullRequest/merge.
func (h *Handler) MergePR(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	prID, err := uuid.Parse(req.PullRequestID)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid pull_request_id UUID")
		return
	}

	pr, err := h.service.MergePR(c.Request.Context(), prID)
	if err != nil {
		if err.Error() == "PR_NOT_FOUND" {
			h.sendError(c, http.StatusNotFound, "NOT_FOUND", "PR not found")
			return
		}
		h.sendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	slog.Info("PR merged", "pr_id", prID)
	c.JSON(http.StatusOK, gin.H{"pr": pr})
}

// ReassignReviewer обрабатывает POST /pullRequest/reassign.
func (h *Handler) ReassignReviewer(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id" binding:"required"`
		OldUserID     string `json:"old_user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	prID, err := uuid.Parse(req.PullRequestID)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid pull_request_id UUID")
		return
	}

	oldUserID, err := uuid.Parse(req.OldUserID)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid old_user_id UUID")
		return
	}

	pr, newReviewerID, err := h.service.ReassignReviewer(c.Request.Context(), prID, oldUserID)
	if err != nil {
		switch err.Error() {
		case "PR_MERGED":
			h.sendError(c, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
		case "NOT_ASSIGNED":
			h.sendError(c, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
		case "NO_CANDIDATE":
			h.sendError(c, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
		case "PR_NOT_FOUND":
			h.sendError(c, http.StatusNotFound, "NOT_FOUND", "PR not found")
		default:
			h.sendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	slog.Info("Reviewer reassigned", "pr_id", prID, "old_reviewer", oldUserID, "new_reviewer", newReviewerID)
	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}

// GetUserReviews обрабатывает GET /users/getReview?user_id=...
func (h *Handler) GetUserReviews(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", "user_id is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.sendError(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid user_id UUID")
		return
	}

	prs, err := h.service.GetUserReviews(c.Request.Context(), userID)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

// GetStatistics обрабатывает GET /stats.
func (h *Handler) GetStatistics(c *gin.Context) {
	stats, err := h.service.GetStatistics(c.Request.Context())
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SetupRouter настраивает маршруты для Gin роутера.
func (h *Handler) SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Stats
	r.GET("/stats", h.GetStatistics)

	// Teams
	r.POST("/team/add", h.CreateTeam)
	r.GET("/team/get", h.GetTeam)

	// Users
	r.POST("/users/setIsActive", middleware.AdminAuth(h.adminToken), h.SetUserActive)
	r.GET("/users/getReview", h.GetUserReviews)

	// Pull Requests
	r.POST("/pullRequest/create", h.CreatePR)
	r.POST("/pullRequest/merge", h.MergePR)
	r.POST("/pullRequest/reassign", h.ReassignReviewer)

	return r
}
