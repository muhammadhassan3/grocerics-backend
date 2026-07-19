package v1

import (
	"net/url"
	"strings"
	"time"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/util"

	"github.com/gin-gonic/gin"
)

var (
	userSortable     = []string{"created_at", "name"}
	userSortFallback = query.Sort{Column: "created_at", Direction: "desc"}
)

func RegisterUserRoutes(r *gin.Engine, authDeps *middleware.AuthDeps, users *repository.UserRepository) {
	g := r.Group("/v1/users")
	g.Use(middleware.AuthMiddleware(authDeps), middleware.AdminOnly())
	g.GET("", listUsers(users))
	g.POST("/ban", BanUser(users))
}

// @Summary List clients
// @Description Paginated list of mobile clients (name, phone, status).
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param page query int false "page number (default 1)"
// @Param page_size query int false "page size, max 100 (default 20)"
// @Param sort query string false "sort column: created_at | name (default created_at)"
// @Param order query string false "asc | desc (default desc)"
// @Param status query string false "filter by status: active | disabled | banned"
// @Param search query string false "search by name or phone (max 128 chars)"
// @Success 200 {object} dto.Response{data=dto.UserListResponseDTO} "Users list"
// @Failure 400 {object} dto.Response "Bad request"
// @Router /v1/users [get]
func listUsers(repo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		filters, err := parseUserFilters(c)
		if err != nil {
			c.Error(err)
			return
		}
		page := query.PageFromContext(c)
		sort := query.SortFromContext(c, userSortable, userSortFallback)

		items, total, err := repo.List(filters, page, sort)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(200, dto.Response{
			Status: "success",
			Data: dto.UserListResponseDTO{
				Items: usersToDTO(items),
				Meta:  query.BuildMeta(total, page),
			},
		})
	}
}

type BanUserRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

// @Summary Ban a client
// @Description Sets the client's status to "banned". Irreversible.
// @Tags Users
// @Accept json
// @Produce json
// @Param BanUserRequest body BanUserRequest true "Ban User Request"
// @Success 200 {object} dto.Response "User banned successfully"
// @Failure 404 {object} dto.Response "User not found"
// @Router /v1/users/ban [post]
func BanUser(repo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BanUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(errs.BadRequest("VALIDATION", util.ParseValidationError(err).Error()))
			return
		}
		affected, err := repo.SetStatus(req.UserID, domain.UserStatusBanned)
		if err != nil {
			c.Error(err)
			return
		}
		if affected == 0 {
			c.Error(errs.NotFound("USER_NOT_FOUND", "user not found"))
			return
		}
		c.JSON(200, dto.Response{Status: "success", Message: "User banned successfully"})
	}
}

func parseUserFilters(c *gin.Context) (repository.UserFilters, error) {
	q, _ := url.ParseQuery(c.Request.URL.RawQuery)
	var f repository.UserFilters
	if v := q.Get("status"); v != "" {
		st := domain.UserStatus(v)
		if !st.IsValid() {
			return f, errs.BadRequest("VALIDATION", "invalid status")
		}
		f.Status = &st
	}
	if v := strings.TrimSpace(q.Get("search")); v != "" {
		if len(v) > 128 {
			return f, errs.BadRequest("VALIDATION", "search too long")
		}
		f.Search = v
	}
	return f, nil
}

func usersToDTO(us []domain.User) []dto.UserListItemDTO {
	out := make([]dto.UserListItemDTO, 0, len(us))
	for _, u := range us {
		out = append(out, dto.UserListItemDTO{
			ID:        u.ID,
			Name:      u.Name,
			Phone:     u.Phone,
			Status:    string(u.Status),
			CreatedAt: u.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}
