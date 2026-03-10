package domain

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type OrganizationDomain struct {
	id             uuid.UUID
	organizationID OrganizationID
	hostname       string
	kind           OrganizationDomainKind
	isPrimary      bool
	verifiedAt     *time.Time
	disabledAt     *time.Time
	createdAt      time.Time
	updatedAt      *time.Time
}

type OrganizationDomainDraft struct {
	Hostname  string
	Kind      string
	IsPrimary bool
}

type NewOrganizationDomainParams struct {
	ID             uuid.UUID
	OrganizationID OrganizationID
	Hostname       string
	Kind           OrganizationDomainKind
	IsPrimary      bool
	VerifiedAt     *time.Time
	DisabledAt     *time.Time
	Now            time.Time
}

type RehydrateOrganizationDomainParams struct {
	ID             uuid.UUID
	OrganizationID OrganizationID
	Hostname       string
	Kind           OrganizationDomainKind
	IsPrimary      bool
	VerifiedAt     *time.Time
	DisabledAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewOrganizationDomain(p NewOrganizationDomainParams) (*OrganizationDomain, error) {
	if p.ID == uuid.Nil {
		return nil, ErrOrganizationDomainIDEmpty
	}
	if p.OrganizationID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.Now.IsZero() {
		return nil, ErrNowRequired
	}
	hostname, err := NormalizeOrganizationHostname(p.Hostname)
	if err != nil {
		return nil, err
	}
	if !p.Kind.IsValid() {
		return nil, ErrOrganizationDomainKindInvalid
	}
	if p.DisabledAt != nil && p.DisabledAt.Before(p.Now) {
		return nil, ErrTimestampsInvalid
	}

	updatedAt := p.Now

	return &OrganizationDomain{
		id:             p.ID,
		organizationID: p.OrganizationID,
		hostname:       hostname,
		kind:           p.Kind,
		isPrimary:      p.IsPrimary,
		verifiedAt:     cloneTimePtr(p.VerifiedAt),
		disabledAt:     cloneTimePtr(p.DisabledAt),
		createdAt:      p.Now,
		updatedAt:      &updatedAt,
	}, nil
}

func RehydrateOrganizationDomain(p RehydrateOrganizationDomainParams) (*OrganizationDomain, error) {
	if p.ID == uuid.Nil {
		return nil, ErrOrganizationDomainIDEmpty
	}
	if p.OrganizationID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return nil, ErrTimestampsMissing
	}
	if p.UpdatedAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}
	hostname, err := NormalizeOrganizationHostname(p.Hostname)
	if err != nil {
		return nil, err
	}
	if !p.Kind.IsValid() {
		return nil, ErrOrganizationDomainKindInvalid
	}
	if p.DisabledAt != nil && p.DisabledAt.Before(p.CreatedAt) {
		return nil, ErrTimestampsInvalid
	}

	updatedAt := p.UpdatedAt

	return &OrganizationDomain{
		id:             p.ID,
		organizationID: p.OrganizationID,
		hostname:       hostname,
		kind:           p.Kind,
		isPrimary:      p.IsPrimary,
		verifiedAt:     cloneTimePtr(p.VerifiedAt),
		disabledAt:     cloneTimePtr(p.DisabledAt),
		createdAt:      p.CreatedAt,
		updatedAt:      &updatedAt,
	}, nil
}

func (d *OrganizationDomain) ID() uuid.UUID                  { return d.id }
func (d *OrganizationDomain) OrganizationID() OrganizationID { return d.organizationID }
func (d *OrganizationDomain) Hostname() string               { return d.hostname }
func (d *OrganizationDomain) Kind() OrganizationDomainKind   { return d.kind }
func (d *OrganizationDomain) IsPrimary() bool                { return d.isPrimary }
func (d *OrganizationDomain) VerifiedAt() *time.Time         { return cloneTimePtr(d.verifiedAt) }
func (d *OrganizationDomain) DisabledAt() *time.Time         { return cloneTimePtr(d.disabledAt) }
func (d *OrganizationDomain) CreatedAt() time.Time           { return d.createdAt }
func (d *OrganizationDomain) UpdatedAt() *time.Time          { return cloneTimePtr(d.updatedAt) }
func (d *OrganizationDomain) IsVerified() bool               { return d.verifiedAt != nil && d.disabledAt == nil }

func NormalizeOrganizationHostname(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrOrganizationDomainHostnameInvalid
	}

	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return "", ErrOrganizationDomainHostnameInvalid
		}
		if parsed.Host == "" {
			return "", ErrOrganizationDomainHostnameInvalid
		}
		if parsed.Path != "" && parsed.Path != "/" {
			return "", ErrOrganizationDomainHostnameInvalid
		}
		if parsed.RawQuery != "" || parsed.Fragment != "" {
			return "", ErrOrganizationDomainHostnameInvalid
		}
		value = parsed.Host
	}

	value = strings.TrimSpace(strings.TrimSuffix(value, "."))
	if value == "" || strings.ContainsAny(value, "/?#@") {
		return "", ErrOrganizationDomainHostnameInvalid
	}

	if host, port, err := net.SplitHostPort(value); err == nil {
		if host == "" || port == "" {
			return "", ErrOrganizationDomainHostnameInvalid
		}
		value = host
	} else if strings.Count(value, ":") > 0 {
		return "", ErrOrganizationDomainHostnameInvalid
	}

	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || len(value) > 253 {
		return "", ErrOrganizationDomainHostnameInvalid
	}

	labels := strings.Split(value, ".")
	if len(labels) < 2 {
		return "", ErrOrganizationDomainHostnameInvalid
	}
	for _, label := range labels {
		if !isValidOrganizationHostnameLabel(label) {
			return "", ErrOrganizationDomainHostnameInvalid
		}
	}

	return value, nil
}

func BuildOrganizationDomains(organizationID OrganizationID, drafts []OrganizationDomainDraft, existing []OrganizationDomain, now time.Time) ([]OrganizationDomain, error) {
	if organizationID.IsZero() {
		return nil, ErrOrganizationIDEmpty
	}
	if now.IsZero() {
		return nil, ErrNowRequired
	}
	if len(drafts) == 0 {
		return []OrganizationDomain{}, nil
	}

	normalized := make([]OrganizationDomainDraft, 0, len(drafts))
	seen := make(map[string]struct{}, len(drafts))
	primaryCount := 0
	for _, draft := range drafts {
		hostname, err := NormalizeOrganizationHostname(draft.Hostname)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[hostname]; ok {
			return nil, ErrOrganizationDomainDuplicate
		}
		seen[hostname] = struct{}{}
		if draft.IsPrimary {
			primaryCount++
		}
		normalized = append(normalized, OrganizationDomainDraft{
			Hostname:  hostname,
			Kind:      strings.TrimSpace(draft.Kind),
			IsPrimary: draft.IsPrimary,
		})
	}

	if primaryCount > 1 {
		return nil, ErrOrganizationDomainPrimaryInvalid
	}
	if primaryCount == 0 && len(normalized) > 0 {
		normalized[0].IsPrimary = true
	}

	existingByHostname := make(map[string]OrganizationDomain, len(existing))
	for _, item := range existing {
		existingByHostname[item.Hostname()] = item
	}

	out := make([]OrganizationDomain, 0, len(normalized))
	for _, draft := range normalized {
		kind := ParseOrganizationDomainKind(draft.Kind)
		if !kind.IsValid() {
			return nil, ErrOrganizationDomainKindInvalid
		}

		var (
			record *OrganizationDomain
			err    error
		)
		if current, ok := existingByHostname[draft.Hostname]; ok {
			verifiedAt := current.VerifiedAt()
			switch kind {
			case OrganizationDomainKindSubdomain:
				if verifiedAt == nil {
					verifiedAt = &now
				}
			case OrganizationDomainKindCustomDomain:
				if current.Kind() != kind {
					verifiedAt = nil
				}
			}
			record, err = RehydrateOrganizationDomain(RehydrateOrganizationDomainParams{
				ID:             current.ID(),
				OrganizationID: organizationID,
				Hostname:       draft.Hostname,
				Kind:           kind,
				IsPrimary:      draft.IsPrimary,
				VerifiedAt:     verifiedAt,
				CreatedAt:      current.CreatedAt(),
				UpdatedAt:      now,
			})
		} else {
			var verifiedAt *time.Time
			if kind == OrganizationDomainKindSubdomain {
				verifiedAt = &now
			}
			record, err = NewOrganizationDomain(NewOrganizationDomainParams{
				ID:             uuid.New(),
				OrganizationID: organizationID,
				Hostname:       draft.Hostname,
				Kind:           kind,
				IsPrimary:      draft.IsPrimary,
				VerifiedAt:     verifiedAt,
				Now:            now,
			})
		}
		if err != nil {
			return nil, err
		}
		out = append(out, *record)
	}

	return out, nil
}

func isValidOrganizationHostnameLabel(label string) bool {
	if label == "" || len(label) > 63 {
		return false
	}
	if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
		return false
	}
	for _, r := range label {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}
