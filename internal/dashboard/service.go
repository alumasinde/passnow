package dashboard

import (
	"context"
	"time"

	"gatepass/internal/roles"
)

type Service struct {
	repo     *Repository
	roles    *roles.Repository
	registry []WidgetDefinition
}

func NewService(repo *Repository, roleRepo *roles.Repository) *Service {
	return &Service{repo: repo, roles: roleRepo, registry: DefaultRegistry()}
}

// BuildDashboard resolves the caller's current permissions, computes the
// tenant-scoped metrics, and returns only widgets the caller may access.
func (s *Service) BuildDashboard(ctx context.Context, userID, roleID int64) (*Dashboard, error) {
	membership, err := s.roles.MembershipFor(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !membership.IsActive() || membership.RoleID != roleID {
		return nil, roles.ErrNotFound
	}

	permissions, err := s.roles.PermissionCodesForRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	summary, err := s.repo.Summary(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Widget, 0, len(s.registry))
	for _, def := range s.registry {
		if !hasPermissions(permissions, def.Permissions) {
			continue
		}
		value, data := def.Build(summary)
		out = append(out, Widget{
			Code: def.Code, Type: def.Type, Title: def.Title, Icon: def.Icon,
			Accent: def.Accent, Size: def.Size, Order: def.Order,
			Value: value, Data: data,
		})
	}

	return &Dashboard{Widgets: out, GeneratedAt: time.Now().UTC()}, nil
}

func hasPermissions(granted map[string]bool, required []string) bool {
	for _, code := range required {
		if !granted[code] {
			return false
		}
	}
	return true
}
