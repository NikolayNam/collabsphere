-- +goose Up

-- +goose StatementBegin
DO
$$
BEGIN
    IF to_regclass('org.cooperation_applications') IS NULL THEN
        RAISE EXCEPTION 'table "org.cooperation_applications" does not exist';
    END IF;
    IF to_regclass('org.organization_legal_documents') IS NULL THEN
        RAISE EXCEPTION 'table "org.organization_legal_documents" does not exist';
    END IF;
    IF to_regclass('org.organization_legal_document_analysis') IS NULL THEN
        RAISE EXCEPTION 'table "org.organization_legal_document_analysis" does not exist';
    END IF;
    IF to_regclass('integration.organization_document_analysis_jobs') IS NULL THEN
        RAISE EXCEPTION 'table "integration.organization_document_analysis_jobs" does not exist';
    END IF;
END
$$;
-- +goose StatementEnd

INSERT INTO org.cooperation_applications (
    id,
    organization_id,
    status,
    confirmation_email,
    company_name,
    represented_categories,
    minimum_order_amount,
    delivery_geography,
    sales_channels,
    storefront_url,
    contact_first_name,
    contact_last_name,
    contact_job_title,
    price_list_object_id,
    contact_email,
    contact_phone,
    partner_code,
    review_note,
    reviewer_account_id,
    submitted_at,
    reviewed_at,
    created_at,
    updated_at
)
VALUES
    (
        '66000000-0000-0000-0000-000000000001',
        '30000000-0000-0000-0000-000000000001',
        'submitted',
        'partners@severny-foods.demo.local',
        'ООО Северный Фудс',
        'Полуфабрикаты, готовые блюда, HoReCa',
        '15000 RUB',
        'Москва и Санкт-Петербург',
        '["horeca", "retail", "marketplace"]'::jsonb,
        'https://severny-foods.demo.local/catalog',
        'Анна',
        'Власова',
        'Коммерческий директор',
        '40000000-0000-0000-0000-000000000010',
        'partners@severny-foods.demo.local',
        '+74950000001',
        'SEV-DEMO-2026',
        NULL,
        NULL,
        '2026-03-08T10:00:00Z',
        NULL,
        '2026-03-08T09:50:00Z',
        '2026-03-08T10:00:00Z'
    ),
    (
        '66000000-0000-0000-0000-000000000002',
        '30000000-0000-0000-0000-000000000002',
        'draft',
        'buyers@gorod-market.demo.local',
        'ООО ГородМаркет',
        'Розничная сеть, соки, напитки',
        '5000 RUB',
        'Санкт-Петербург',
        '["retail", "darkstore"]'::jsonb,
        'https://gorod-market.demo.local/vendors',
        'Мария',
        'Кузнецова',
        'Руководитель закупок',
        NULL,
        'buyers@gorod-market.demo.local',
        '+74950000002',
        'GM-DEMO-2026',
        NULL,
        NULL,
        NULL,
        NULL,
        '2026-03-08T09:55:00Z',
        '2026-03-08T09:55:00Z'
    )
ON CONFLICT (organization_id) DO UPDATE
SET status = EXCLUDED.status,
    confirmation_email = EXCLUDED.confirmation_email,
    company_name = EXCLUDED.company_name,
    represented_categories = EXCLUDED.represented_categories,
    minimum_order_amount = EXCLUDED.minimum_order_amount,
    delivery_geography = EXCLUDED.delivery_geography,
    sales_channels = EXCLUDED.sales_channels,
    storefront_url = EXCLUDED.storefront_url,
    contact_first_name = EXCLUDED.contact_first_name,
    contact_last_name = EXCLUDED.contact_last_name,
    contact_job_title = EXCLUDED.contact_job_title,
    price_list_object_id = EXCLUDED.price_list_object_id,
    contact_email = EXCLUDED.contact_email,
    contact_phone = EXCLUDED.contact_phone,
    partner_code = EXCLUDED.partner_code,
    review_note = EXCLUDED.review_note,
    reviewer_account_id = EXCLUDED.reviewer_account_id,
    submitted_at = EXCLUDED.submitted_at,
    reviewed_at = EXCLUDED.reviewed_at,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at;

INSERT INTO org.organization_legal_documents (
    id,
    organization_id,
    document_type,
    status,
    object_id,
    title,
    uploaded_by_account_id,
    reviewer_account_id,
    review_note,
    created_at,
    updated_at,
    reviewed_at,
    deleted_at
)
VALUES
    (
        '67000000-0000-0000-0000-000000000001',
        '30000000-0000-0000-0000-000000000001',
        'inn_certificate',
        'approved',
        '40000000-0000-0000-0000-000000000011',
        'Свидетельство ИНН',
        '10000000-0000-0000-0000-000000000001',
        '10000000-0000-0000-0000-000000000002',
        'Поля извлечены корректно, документ принят.',
        '2026-03-08T10:05:00Z',
        '2026-03-08T10:20:00Z',
        '2026-03-08T10:20:00Z',
        NULL
    ),
    (
        '67000000-0000-0000-0000-000000000002',
        '30000000-0000-0000-0000-000000000001',
        'ogrn_extract',
        'pending',
        '40000000-0000-0000-0000-000000000012',
        'Выписка ОГРН',
        '10000000-0000-0000-0000-000000000001',
        NULL,
        NULL,
        '2026-03-08T10:06:00Z',
        '2026-03-08T10:06:00Z',
        NULL,
        NULL
    ),
    (
        '67000000-0000-0000-0000-000000000003',
        '30000000-0000-0000-0000-000000000002',
        'charter',
        'pending',
        '40000000-0000-0000-0000-000000000013',
        'Устав компании',
        '10000000-0000-0000-0000-000000000003',
        NULL,
        NULL,
        '2026-03-08T10:07:00Z',
        '2026-03-08T10:07:00Z',
        NULL,
        NULL
    )
ON CONFLICT (id) DO UPDATE
SET organization_id = EXCLUDED.organization_id,
    document_type = EXCLUDED.document_type,
    status = EXCLUDED.status,
    object_id = EXCLUDED.object_id,
    title = EXCLUDED.title,
    uploaded_by_account_id = EXCLUDED.uploaded_by_account_id,
    reviewer_account_id = EXCLUDED.reviewer_account_id,
    review_note = EXCLUDED.review_note,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at,
    reviewed_at = EXCLUDED.reviewed_at,
    deleted_at = EXCLUDED.deleted_at;

INSERT INTO org.organization_legal_document_analysis (
    id,
    document_id,
    organization_id,
    status,
    provider,
    extracted_text,
    summary,
    extracted_fields_json,
    detected_document_type,
    confidence_score,
    requested_at,
    started_at,
    completed_at,
    updated_at,
    last_error
)
VALUES
    (
        '68000000-0000-0000-0000-000000000001',
        '67000000-0000-0000-0000-000000000001',
        '30000000-0000-0000-0000-000000000001',
        'completed',
        'generic-http',
        'ИНН 7701234567. ООО Северный Фудс. Дата регистрации 12.04.2022.',
        'Распознан ИНН поставщика и реквизиты юридического лица.',
        '{"inn":"7701234567","companyName":"ООО Северный Фудс","registrationDate":"2022-04-12"}'::jsonb,
        'inn_certificate',
        0.98,
        '2026-03-08T10:05:30Z',
        '2026-03-08T10:05:35Z',
        '2026-03-08T10:05:40Z',
        '2026-03-08T10:05:40Z',
        NULL
    ),
    (
        '68000000-0000-0000-0000-000000000002',
        '67000000-0000-0000-0000-000000000002',
        '30000000-0000-0000-0000-000000000001',
        'failed',
        'generic-http',
        NULL,
        NULL,
        '{}'::jsonb,
        NULL,
        NULL,
        '2026-03-08T10:06:30Z',
        '2026-03-08T10:06:35Z',
        NULL,
        '2026-03-08T10:06:40Z',
        'Provider timeout while parsing scanned PDF'
    ),
    (
        '68000000-0000-0000-0000-000000000003',
        '67000000-0000-0000-0000-000000000003',
        '30000000-0000-0000-0000-000000000002',
        'pending',
        'generic-http',
        NULL,
        NULL,
        '{}'::jsonb,
        NULL,
        NULL,
        '2026-03-08T10:07:30Z',
        NULL,
        NULL,
        '2026-03-08T10:07:30Z',
        NULL
    )
ON CONFLICT (document_id) DO UPDATE
SET organization_id = EXCLUDED.organization_id,
    status = EXCLUDED.status,
    provider = EXCLUDED.provider,
    extracted_text = EXCLUDED.extracted_text,
    summary = EXCLUDED.summary,
    extracted_fields_json = EXCLUDED.extracted_fields_json,
    detected_document_type = EXCLUDED.detected_document_type,
    confidence_score = EXCLUDED.confidence_score,
    requested_at = EXCLUDED.requested_at,
    started_at = EXCLUDED.started_at,
    completed_at = EXCLUDED.completed_at,
    updated_at = EXCLUDED.updated_at,
    last_error = EXCLUDED.last_error;

INSERT INTO integration.organization_document_analysis_jobs (
    id,
    document_id,
    status,
    provider,
    attempts,
    available_at,
    leased_until,
    completed_at,
    last_error,
    created_at,
    updated_at
)
VALUES
    (
        '69000000-0000-0000-0000-000000000001',
        '67000000-0000-0000-0000-000000000001',
        'completed',
        'generic-http',
        1,
        '2026-03-08T10:05:30Z',
        NULL,
        '2026-03-08T10:05:40Z',
        NULL,
        '2026-03-08T10:05:30Z',
        '2026-03-08T10:05:40Z'
    ),
    (
        '69000000-0000-0000-0000-000000000002',
        '67000000-0000-0000-0000-000000000002',
        'failed',
        'generic-http',
        2,
        '2026-03-08T10:10:00Z',
        NULL,
        NULL,
        'Provider timeout while parsing scanned PDF',
        '2026-03-08T10:06:30Z',
        '2026-03-08T10:06:40Z'
    ),
    (
        '69000000-0000-0000-0000-000000000003',
        '67000000-0000-0000-0000-000000000003',
        'pending',
        'generic-http',
        0,
        '2026-03-08T10:07:30Z',
        NULL,
        NULL,
        NULL,
        '2026-03-08T10:07:30Z',
        '2026-03-08T10:07:30Z'
    )
ON CONFLICT (document_id) DO UPDATE
SET status = EXCLUDED.status,
    provider = EXCLUDED.provider,
    attempts = EXCLUDED.attempts,
    available_at = EXCLUDED.available_at,
    leased_until = EXCLUDED.leased_until,
    completed_at = EXCLUDED.completed_at,
    last_error = EXCLUDED.last_error,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at;

-- +goose Down

DELETE FROM integration.organization_document_analysis_jobs
WHERE id IN (
    '69000000-0000-0000-0000-000000000001',
    '69000000-0000-0000-0000-000000000002',
    '69000000-0000-0000-0000-000000000003'
);

DELETE FROM org.organization_legal_document_analysis
WHERE id IN (
    '68000000-0000-0000-0000-000000000001',
    '68000000-0000-0000-0000-000000000002',
    '68000000-0000-0000-0000-000000000003'
);

DELETE FROM org.organization_legal_documents
WHERE id IN (
    '67000000-0000-0000-0000-000000000001',
    '67000000-0000-0000-0000-000000000002',
    '67000000-0000-0000-0000-000000000003'
);

DELETE FROM org.cooperation_applications
WHERE organization_id IN (
    '30000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000002'
);
