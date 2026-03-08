package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type conferenceRow struct {
	ID                  uuid.UUID  `gorm:"column:id"`
	ChannelID           uuid.UUID  `gorm:"column:channel_id"`
	Kind                string     `gorm:"column:kind"`
	Status              string     `gorm:"column:status"`
	Provider            string     `gorm:"column:provider"`
	Title               string     `gorm:"column:title"`
	JitsiRoomName       string     `gorm:"column:jitsi_room_name"`
	ScheduledStartAt    *time.Time `gorm:"column:scheduled_start_at"`
	StartedAt           *time.Time `gorm:"column:started_at"`
	EndedAt             *time.Time `gorm:"column:ended_at"`
	RecordingEnabled    bool       `gorm:"column:recording_enabled"`
	RecordingStartedAt  *time.Time `gorm:"column:recording_started_at"`
	RecordingStoppedAt  *time.Time `gorm:"column:recording_stopped_at"`
	TranscriptionStatus string     `gorm:"column:transcription_status"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           *time.Time `gorm:"column:updated_at"`
	CreatedBy           *uuid.UUID `gorm:"column:created_by"`
	UpdatedBy           *uuid.UUID `gorm:"column:updated_by"`
}

type conferenceRecordingRow struct {
	ID           uuid.UUID  `gorm:"column:id"`
	ConferenceID uuid.UUID  `gorm:"column:conference_id"`
	ObjectID     uuid.UUID  `gorm:"column:object_id"`
	DurationSec  *int32     `gorm:"column:duration_sec"`
	MimeType     *string    `gorm:"column:mime_type"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	CreatedBy    *uuid.UUID `gorm:"column:created_by"`
	FileName     *string    `gorm:"column:file_name"`
	Bucket       *string    `gorm:"column:bucket"`
	ObjectKey    *string    `gorm:"column:object_key"`
	ContentType  *string    `gorm:"column:content_type"`
	SizeBytes    *int64     `gorm:"column:size_bytes"`
}

type transcriptRow struct {
	ConferenceID      uuid.UUID       `gorm:"column:conference_id"`
	TranscriptText    string          `gorm:"column:transcript_text"`
	SegmentsJSON      json.RawMessage `gorm:"column:segments_json"`
	LanguageCode      *string         `gorm:"column:language_code"`
	SourceRecordingID *uuid.UUID      `gorm:"column:source_recording_id"`
	CreatedAt         time.Time       `gorm:"column:created_at"`
	UpdatedAt         *time.Time      `gorm:"column:updated_at"`
}

type transcriptionLeaseRow struct {
	JobID         uuid.UUID `gorm:"column:job_id"`
	ConferenceID  uuid.UUID `gorm:"column:conference_id"`
	RecordingID   uuid.UUID `gorm:"column:recording_id"`
	ObjectID      uuid.UUID `gorm:"column:object_id"`
	Bucket        string    `gorm:"column:bucket"`
	ObjectKey     string    `gorm:"column:object_key"`
	MimeType      *string   `gorm:"column:mime_type"`
	Attempts      int       `gorm:"column:attempts"`
	ConferenceCID uuid.UUID `gorm:"column:conference_cid"`
	Status        string    `gorm:"column:status"`
	FileName      string    `gorm:"column:file_name"`
}

func (r *Repo) CreateConference(ctx context.Context, conference collabdomain.Conference) (*collabdomain.Conference, error) {
	if conference.ID == uuid.Nil {
		conference.ID = uuid.New()
	}
	if conference.CreatedAt.IsZero() {
		conference.CreatedAt = time.Now().UTC()
	}
	updatedAt := conference.CreatedAt
	if err := r.dbFrom(ctx).WithContext(ctx).Table("collab.conferences").Create(map[string]any{
		"id":                   conference.ID,
		"channel_id":           conference.ChannelID,
		"kind":                 string(conference.Kind),
		"status":               string(conference.Status),
		"provider":             conference.Provider,
		"title":                conference.Title,
		"jitsi_room_name":      conference.JitsiRoomName,
		"scheduled_start_at":   conference.ScheduledStartAt,
		"started_at":           conference.StartedAt,
		"ended_at":             conference.EndedAt,
		"recording_enabled":    conference.RecordingEnabled,
		"recording_started_at": conference.RecordingStartedAt,
		"recording_stopped_at": conference.RecordingStoppedAt,
		"transcription_status": string(conference.TranscriptionStatus),
		"created_at":           conference.CreatedAt,
		"updated_at":           &updatedAt,
		"created_by":           conference.CreatedBy,
		"updated_by":           conference.UpdatedBy,
	}).Error; err != nil {
		return nil, err
	}
	return r.GetConferenceByID(ctx, conference.ID)
}

func (r *Repo) GetConferenceByID(ctx context.Context, conferenceID uuid.UUID) (*collabdomain.Conference, error) {
	var row conferenceRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.conferences").
		Select("id, channel_id, kind, status, provider, title, jitsi_room_name, scheduled_start_at, started_at, ended_at, recording_enabled, recording_started_at, recording_stopped_at, transcription_status, created_at, updated_at, created_by, updated_by").
		Where("id = ?", conferenceID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapConference(row), nil
}

func (r *Repo) ListConferencesByChannel(ctx context.Context, channelID uuid.UUID) ([]collabdomain.Conference, error) {
	var rows []conferenceRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.conferences").
		Select("id, channel_id, kind, status, provider, title, jitsi_room_name, scheduled_start_at, started_at, ended_at, recording_enabled, recording_started_at, recording_stopped_at, transcription_status, created_at, updated_at, created_by, updated_by").
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]collabdomain.Conference, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapConference(row))
	}
	return out, nil
}

func (r *Repo) UpdateConference(ctx context.Context, conferenceID uuid.UUID, updates map[string]any) (*collabdomain.Conference, error) {
	if len(updates) == 0 {
		return r.GetConferenceByID(ctx, conferenceID)
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Table("collab.conferences").Where("id = ?", conferenceID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetConferenceByID(ctx, conferenceID)
}

func (r *Repo) CreateConferenceRecording(ctx context.Context, conferenceID uuid.UUID, object collabdomain.StorageObject, createdBy *uuid.UUID, durationSec *int32, mimeType *string, now time.Time) (*collabdomain.ConferenceRecording, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if object.ID == uuid.Nil {
		object.ID = uuid.New()
	}
	if object.CreatedAt.IsZero() {
		object.CreatedAt = now
	}
	if err := r.dbFrom(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("storage.objects").Create(map[string]any{
			"id":              object.ID,
			"organization_id": object.OrganizationID,
			"bucket":          object.Bucket,
			"object_key":      object.ObjectKey,
			"file_name":       object.FileName,
			"content_type":    object.ContentType,
			"size_bytes":      object.SizeBytes,
			"checksum_sha256": object.ChecksumSHA256,
			"created_at":      object.CreatedAt,
		}).Error; err != nil {
			return err
		}
		return tx.Table("collab.conference_recordings").Create(map[string]any{
			"id":            uuid.New(),
			"conference_id": conferenceID,
			"object_id":     object.ID,
			"created_at":    now,
			"created_by":    createdBy,
			"duration_sec":  durationSec,
			"mime_type":     mimeType,
		}).Error
	}); err != nil {
		return nil, err
	}
	return r.GetLatestConferenceRecording(ctx, conferenceID)
}

func (r *Repo) GetLatestConferenceRecording(ctx context.Context, conferenceID uuid.UUID) (*collabdomain.ConferenceRecording, error) {
	var row conferenceRecordingRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.conference_recordings AS cr").
		Select("cr.id, cr.conference_id, cr.object_id, cr.duration_sec, cr.mime_type, cr.created_at, cr.created_by, so.file_name, so.bucket, so.object_key, so.content_type, so.size_bytes").
		Joins("JOIN storage.objects AS so ON so.id = cr.object_id").
		Where("cr.conference_id = ?", conferenceID).
		Order("cr.created_at DESC").
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapRecording(row), nil
}

func (r *Repo) UpsertConferenceTranscript(ctx context.Context, transcript collabdomain.ConferenceTranscript, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if len(transcript.SegmentsJSON) == 0 {
		transcript.SegmentsJSON = json.RawMessage(`[]`)
	}
	return r.dbFrom(ctx).WithContext(ctx).Exec(`
		INSERT INTO collab.conference_transcripts (conference_id, transcript_text, segments_json, language_code, source_recording_id, created_at, updated_at)
		VALUES (@conference_id, @transcript_text, CAST(@segments_json AS jsonb), @language_code, @source_recording_id, @created_at, @updated_at)
		ON CONFLICT (conference_id)
		DO UPDATE SET transcript_text = EXCLUDED.transcript_text,
		              segments_json = EXCLUDED.segments_json,
		              language_code = EXCLUDED.language_code,
		              source_recording_id = EXCLUDED.source_recording_id,
		              updated_at = EXCLUDED.updated_at
	`, map[string]any{
		"conference_id":       transcript.ConferenceID,
		"transcript_text":     transcript.TranscriptText,
		"segments_json":       string(transcript.SegmentsJSON),
		"language_code":       transcript.LanguageCode,
		"source_recording_id": transcript.SourceRecordingID,
		"created_at":          now,
		"updated_at":          now,
	}).Error
}

func (r *Repo) GetConferenceTranscript(ctx context.Context, conferenceID uuid.UUID) (*collabdomain.ConferenceTranscript, error) {
	var row transcriptRow
	if err := r.dbFrom(ctx).WithContext(ctx).
		Table("collab.conference_transcripts").
		Select("conference_id, transcript_text, segments_json, language_code, source_recording_id, created_at, updated_at").
		Where("conference_id = ?", conferenceID).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &collabdomain.ConferenceTranscript{
		ConferenceID:      row.ConferenceID,
		TranscriptText:    row.TranscriptText,
		SegmentsJSON:      row.SegmentsJSON,
		LanguageCode:      row.LanguageCode,
		SourceRecordingID: row.SourceRecordingID,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}, nil
}

func (r *Repo) CreateJitsiWebhookInbox(ctx context.Context, providerEventID, eventType string, payload json.RawMessage, now time.Time) (uuid.UUID, bool, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	id := uuid.New()
	err := r.dbFrom(ctx).WithContext(ctx).Table("integration.jitsi_webhook_inbox").Create(map[string]any{
		"id":                id,
		"provider_event_id": providerEventID,
		"event_type":        eventType,
		"payload_json":      payload,
		"received_at":       now,
	}).Error
	if isUniqueViolation(err) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, err
	}
	return id, true, nil
}

func (r *Repo) MarkJitsiWebhookProcessed(ctx context.Context, inboxID uuid.UUID, processedAt time.Time, errMessage *string) error {
	if processedAt.IsZero() {
		processedAt = time.Now().UTC()
	}
	updates := map[string]any{"processed_at": processedAt}
	if errMessage != nil {
		updates["error_message"] = *errMessage
	}
	return r.dbFrom(ctx).WithContext(ctx).Table("integration.jitsi_webhook_inbox").Where("id = ?", inboxID).Updates(updates).Error
}

func (r *Repo) EnqueueTranscriptionJob(ctx context.Context, conferenceID, recordingID uuid.UUID, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return r.dbFrom(ctx).WithContext(ctx).Table("integration.transcription_jobs").Create(map[string]any{
		"id":            uuid.New(),
		"conference_id": conferenceID,
		"recording_id":  recordingID,
		"status":        "pending",
		"provider":      "whisper",
		"attempts":      0,
		"available_at":  now,
		"created_at":    now,
		"updated_at":    now,
	}).Error
}

func (r *Repo) LeaseNextTranscriptionJob(ctx context.Context, now time.Time, leaseFor time.Duration) (*transcriptionLeaseRow, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	leasedUntil := now.Add(leaseFor)
	var job transcriptionLeaseRow
	err := r.dbFrom(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row transcriptionLeaseRow
		if err := tx.Table("integration.transcription_jobs AS tj").
			Select("tj.id AS job_id, tj.conference_id, tj.recording_id, cr.object_id, so.bucket, so.object_key, cr.mime_type, tj.attempts, tj.status, so.file_name").
			Joins("JOIN collab.conference_recordings AS cr ON cr.id = tj.recording_id").
			Joins("JOIN storage.objects AS so ON so.id = cr.object_id").
			Where("tj.status IN ('pending', 'failed') AND tj.available_at <= ? AND (tj.leased_until IS NULL OR tj.leased_until < ?)", now, now).
			Order("tj.available_at ASC, tj.created_at ASC").
			Take(&row).Error; err != nil {
			return err
		}
		if err := tx.Table("integration.transcription_jobs").Where("id = ?", row.JobID).Updates(map[string]any{
			"status":       "leased",
			"leased_until": leasedUntil,
			"attempts":     row.Attempts + 1,
			"updated_at":   now,
		}).Error; err != nil {
			return err
		}
		job = row
		job.Attempts = row.Attempts + 1
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (r *Repo) CompleteTranscriptionJob(ctx context.Context, jobID uuid.UUID, completedAt time.Time) error {
	if completedAt.IsZero() {
		completedAt = time.Now().UTC()
	}
	return r.dbFrom(ctx).WithContext(ctx).Table("integration.transcription_jobs").Where("id = ?", jobID).Updates(map[string]any{
		"status":       "completed",
		"completed_at": completedAt,
		"leased_until": nil,
		"updated_at":   completedAt,
	}).Error
}

func (r *Repo) FailTranscriptionJob(ctx context.Context, jobID uuid.UUID, errMessage string, retryAt time.Time) error {
	if retryAt.IsZero() {
		retryAt = time.Now().UTC().Add(30 * time.Second)
	}
	return r.dbFrom(ctx).WithContext(ctx).Table("integration.transcription_jobs").Where("id = ?", jobID).Updates(map[string]any{
		"status":       "failed",
		"available_at": retryAt,
		"leased_until": nil,
		"last_error":   errMessage,
		"updated_at":   time.Now().UTC(),
	}).Error
}

func mapConference(row conferenceRow) *collabdomain.Conference {
	return &collabdomain.Conference{
		ID:                  row.ID,
		ChannelID:           row.ChannelID,
		Kind:                collabdomain.ConferenceKind(row.Kind),
		Status:              collabdomain.ConferenceStatus(row.Status),
		Provider:            row.Provider,
		Title:               row.Title,
		JitsiRoomName:       row.JitsiRoomName,
		ScheduledStartAt:    row.ScheduledStartAt,
		StartedAt:           row.StartedAt,
		EndedAt:             row.EndedAt,
		RecordingEnabled:    row.RecordingEnabled,
		RecordingStartedAt:  row.RecordingStartedAt,
		RecordingStoppedAt:  row.RecordingStoppedAt,
		TranscriptionStatus: collabdomain.TranscriptionStatus(row.TranscriptionStatus),
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
		CreatedBy:           row.CreatedBy,
		UpdatedBy:           row.UpdatedBy,
	}
}

func mapRecording(row conferenceRecordingRow) *collabdomain.ConferenceRecording {
	return &collabdomain.ConferenceRecording{
		ID:           row.ID,
		ConferenceID: row.ConferenceID,
		ObjectID:     row.ObjectID,
		DurationSec:  row.DurationSec,
		MimeType:     row.MimeType,
		CreatedAt:    row.CreatedAt,
		CreatedBy:    row.CreatedBy,
		FileName:     row.FileName,
		Bucket:       row.Bucket,
		ObjectKey:    row.ObjectKey,
		ContentType:  row.ContentType,
		SizeBytes:    row.SizeBytes,
	}
}

func (r *Repo) AddConferenceParticipant(ctx context.Context, conferenceID uuid.UUID, principalType collabdomain.ActorType, accountID, guestID *uuid.UUID, role string, joinedAt time.Time) error {
	if joinedAt.IsZero() {
		joinedAt = time.Now().UTC()
	}
	return r.dbFrom(ctx).WithContext(ctx).Table("collab.conference_participants").Create(map[string]any{
		"id":            uuid.New(),
		"conference_id": conferenceID,
		"actor_type":    string(principalType),
		"account_id":    accountID,
		"guest_id":      guestID,
		"joined_at":     joinedAt,
		"role":          role,
	}).Error
}
