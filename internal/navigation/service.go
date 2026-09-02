package navigation

import (
	"context"
	"time"

	"gatepass/internal/roles"
)

type Service struct {
	roles    *roles.Repository
	registry []Definition
}

func NewService(roleRepo *roles.Repository) *Service {
	return &Service{roles: roleRepo, registry: DefaultRegistry()}
}

// Build resolves the caller's live membership and permissions. Navigation is
// generated on every request so role or permission changes take effect without
// waiting for a token to expire.
func (s *Service) Build(ctx context.Context, userID, roleID int64) (*Response, error) {
	membership, err := s.roles.MembershipFor(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !membership.IsActive() || membership.RoleID != roleID {
		return nil, roles.ErrNotFound
	}

	granted, err := s.roles.PermissionCodesForRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	items := make([]Item, 0, len(s.registry))
	for _, def := range s.registry {
		if !hasAnyPermission(granted, def.AnyPermissions) {
			continue
		}
		items = append(items, def.Item)
	}

	return &Response{Items: items, GeneratedAt: time.Now().UTC()}, nil
}

func hasAnyPermission(granted map[string]bool, required []string) bool {
	if len(required) == 0 {
		return true
	}
	for _, code := range required {
		if granted[code] {
			return true
		}
	}
	return false
}
