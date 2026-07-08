package v1

import (
	"net/url"
	"strings"
	"time"

	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/query"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	userSortable     = []string{"created_at", "name", "email", "role"}
	userSortFallback = query.Sort{Column: "created_at", Direction: "desc"}
)

func RegisterUserRoutes(r *gin.Engine, jwt *auth.JWTService, users *repository.UserRepository) {
	g := r.Group("/v1/users")
	g.Use(middleware.AuthMiddleware(jwt, users))
	g.GET("", listUsers(users))
}

// @Summary List Users
// @Description Paginated list of users. Admins see every tenant (optionally filter via ?company_id or ?unassigned=true). Non-admins are auto-scoped to their own company.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param page query int false "page number (default 1)"
// @Param page_size query int false "page size, max 100 (default 20)"
// @Param sort query string false "sort column: created_at | name | email | role (default created_at)"
// @Param order query string false "asc | desc (default desc)"
// @Param company_id query string false "filter by company UUID (admin-only override; non-admin mismatching value → 403)"
// @Param role query string false "filter by role: admin | client_manager | client"
// @Param status query string false "filter by status: active | disabled"
// @Param search query string false "ILIKE on name + email, max 128 chars"
// @Param unassigned query bool false "admin-only: list users with company_id IS NULL"
// @Success 200 {object} dto.Response{data=dto.UserListResponseDTO} "Users list"
// @Failure 400 {object} dto.Response "Bad request"
// @Failure 401 {object} dto.Response "Unauthorized"
// @Failure 403 {object} dto.Response "Forbidden"
// @Router /v1/users [get]
func listUsers(repo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		scope := auth.MustScope(c)
		filters, err := parseUserFilters(c, scope)
		if err != nil {
			c.Error(err)
			return
		}
		page := query.PageFromContext(c)
		sort := query.SortFromContext(c, userSortable, userSortFallback)

		items, total, err := repo.List(scope, filters, page, sort)
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

func parseUserFilters(c *gin.Context, scope auth.Scope) (repository.UserFilters, error) {
	q, _ := url.ParseQuery(c.Request.URL.RawQuery)
	var f repository.UserFilters

	if v := q.Get("company_id"); v != "" {
		if _, err := uuid.Parse(v); err != nil {
			return f, errs.BadRequest("VALIDATION", "company_id is not a valid UUID")
		}
		if scope.Role != domain.RoleAdmin {
			if scope.CompanyID == nil || *scope.CompanyID != v {
				return f, errs.Forbidden("FORBIDDEN", "cannot filter to another company")
			}
		}
		f.CompanyID = &v
	}
	if v := q.Get("role"); v != "" {
		role := domain.Role(v)
		if !role.IsValid() {
			return f, errs.BadRequest("VALIDATION", "invalid role")
		}
		f.Role = &role
	}
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
	if v := q.Get("unassigned"); v == "true" || v == "1" {
		if scope.Role != domain.RoleAdmin {
			return f, errs.Forbidden("FORBIDDEN", "only admins can list unassigned users")
		}
		f.Unassigned = true
	}
	return f, nil
}

func usersToDTO(us []domain.User) []dto.UserListItemDTO {
	out := make([]dto.UserListItemDTO, 0, len(us))
	for _, u := range us {
		out = append(out, dto.UserListItemDTO{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      string(u.Role),
			Status:    string(u.Status),
			CreatedAt: u.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}
