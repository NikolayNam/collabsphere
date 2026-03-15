-- +goose Up
WITH upserted_level AS (
    INSERT INTO kyc.levels (id, scope, code, name, rank, is_active, created_at, updated_at)
    VALUES (
        gen_random_uuid(),
        'organization',
        'public_directory_org_verified',
        'Public directory verified organization',
        200,
        true,
        now(),
        now()
    )
    ON CONFLICT (scope, code) DO UPDATE
    SET
        name = EXCLUDED.name,
        rank = EXCLUDED.rank,
        is_active = EXCLUDED.is_active,
        updated_at = now()
    RETURNING id
)
INSERT INTO kyc.level_requirements (level_id, document_type, min_count, created_at, updated_at)
SELECT id, 'founding_document', 1, now(), now()
FROM upserted_level
ON CONFLICT (level_id, document_type) DO UPDATE
SET
    min_count = EXCLUDED.min_count,
    updated_at = now();

-- +goose Down
DELETE FROM kyc.level_requirements
WHERE level_id IN (
    SELECT id FROM kyc.levels WHERE scope = 'organization' AND code = 'public_directory_org_verified'
)
AND document_type = 'founding_document';

DELETE FROM kyc.levels
WHERE scope = 'organization'
  AND code = 'public_directory_org_verified';
