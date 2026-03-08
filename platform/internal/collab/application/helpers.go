package application

import (
	"context"
	"strings"
	"time"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
)

func (s *Service) publish(ctx context.Context, event collabdomain.Event) {
	if s.publisher != nil {
		s.publisher.Publish(ctx, event)
	}
}

func (s *Service) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}
	return s.clock.Now().UTC()
}

func actorTypeFromPrincipal(principal authdomain.Principal) collabdomain.ActorType {
	switch {
	case principal.IsAccount():
		return collabdomain.ActorTypeAccount
	case principal.IsGuest():
		return collabdomain.ActorTypeGuest
	default:
		return collabdomain.ActorTypeSystem
	}
}

func principalAccountPtr(principal authdomain.Principal) *uuid.UUID {
	if !principal.IsAccount() {
		return nil
	}
	return uuidPtr(principal.AccountID)
}

func principalGuestPtr(principal authdomain.Principal) *uuid.UUID {
	if !principal.IsGuest() {
		return nil
	}
	return uuidPtr(principal.GuestID)
}

func uuidPtr(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	v := id
	return &v
}

func shortUUID(id uuid.UUID) string {
	value := strings.ReplaceAll(id.String(), "-", "")
	if len(value) > 12 {
		return value[:12]
	}
	return value
}

func sanitizeSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '-', r == '_', r == ' ':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	out := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
