package postgres

import (
	"context"

	appports "github.com/NikolayNam/collabsphere/internal/organizations/application/ports"
)

func (r *OrganizationRepo) ListPublicKYCDirectoryOrganizations(ctx context.Context, kycLevelCode string, limit int) ([]appports.PublicKYCDirectoryOrganization, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	var rows []appports.PublicKYCDirectoryOrganization
	err := r.dbFrom(ctx).WithContext(ctx).Raw(`
SELECT
  o.id,
  o.name,
  o.slug,
  o.description,
  o.website,
  o.industry,
  d.hostname AS primary_domain,
  p.kyc_level_code,
  l.name AS kyc_level_name
FROM org.organizations o
JOIN kyc.organization_profiles p ON p.organization_id = o.id
JOIN kyc.organization_documents kd
  ON kd.organization_id = o.id
 AND kd.deleted_at IS NULL
 AND kd.status = 'verified'
 AND kd.document_type = 'founding_document'
JOIN org.organization_domains d
  ON d.organization_id = o.id
 AND d.disabled_at IS NULL
 AND d.verified_at IS NOT NULL
 AND d.is_primary = TRUE
LEFT JOIN kyc.levels l
  ON l.scope = 'organization'
 AND l.code = p.kyc_level_code
WHERE o.is_active = TRUE
  AND p.status = 'approved'
  AND p.kyc_level_code = ?
  AND btrim(COALESCE(p.legal_name, '')) <> ''
GROUP BY o.id, o.name, o.slug, o.description, o.website, o.industry, d.hostname, p.kyc_level_code, l.name
ORDER BY o.updated_at DESC NULLS LAST, o.created_at DESC
LIMIT ?`, kycLevelCode, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
