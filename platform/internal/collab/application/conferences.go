package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/google/uuid"
)

func (s *Service) CreateConference(ctx context.Context, cmd CreateConferenceCmd) (*collabdomain.Conference, error) {
	access, channel, err := s.requireChannelAccess(ctx, cmd.ChannelID, cmd.Actor)
	if err != nil {
		return nil, err
	}
	if !cmd.Actor.IsAccount() || !access.CanManage {
		return nil, fault.Forbidden("Only channel managers can create conferences")
	}
	kind := strings.TrimSpace(strings.ToLower(cmd.Kind))
	if kind != string(collabdomain.ConferenceKindAudio) && kind != string(collabdomain.ConferenceKindVideo) {
		return nil, fault.Validation("kind must be audio or video")
	}
	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		title = fmt.Sprintf("%s call", strings.Title(kind))
	}
	provider := s.ConferenceProvider()
	now := s.now()
	conference, err := s.repo.CreateConference(ctx, collabdomain.Conference{
		ID:                  uuid.New(),
		ChannelID:           channel.ID,
		Kind:                collabdomain.ConferenceKind(kind),
		Status:              collabdomain.ConferenceStatusScheduled,
		Provider:            provider,
		Title:               title,
		JitsiRoomName:       buildConferenceRoomName(channel.GroupID, channel.ID),
		ScheduledStartAt:    cmd.ScheduledStartAt,
		RecordingEnabled:    cmd.RecordingEnabled,
		TranscriptionStatus: chooseTranscriptionStatus(cmd.RecordingEnabled),
		CreatedBy:           uuidPtr(cmd.Actor.AccountID),
		UpdatedBy:           uuidPtr(cmd.Actor.AccountID),
		CreatedAt:           now,
	})
	if err != nil {
		return nil, fault.Internal("Create conference failed", fault.WithCause(err))
	}
	_, _ = s.repo.CreateMessage(ctx, collabdomain.Message{ID: uuid.New(), ChannelID: channel.ID, Type: collabdomain.MessageTypeSystem, AuthorType: collabdomain.ActorTypeSystem, Body: fmt.Sprintf("Conference scheduled: %s", conference.Title), CreatedAt: now}, nil, nil)
	s.publish(ctx, collabdomain.Event{Type: "conference.created", ChannelID: channel.ID, Payload: conference})
	return conference, nil
}

func (s *Service) ListConferences(ctx context.Context, q ListConferencesQuery) ([]collabdomain.Conference, error) {
	if _, _, err := s.requireChannelAccess(ctx, q.ChannelID, q.Actor); err != nil {
		return nil, err
	}
	items, err := s.repo.ListConferencesByChannel(ctx, q.ChannelID)
	if err != nil {
		return nil, fault.Internal("List conferences failed", fault.WithCause(err))
	}
	return items, nil
}

func (s *Service) CreateConferenceJoinToken(ctx context.Context, cmd CreateConferenceJoinTokenCmd) (*CreateConferenceJoinTokenResult, error) {
	conference, err := s.repo.GetConferenceByID(ctx, cmd.ConferenceID)
	if err != nil {
		return nil, fault.Internal("Load conference failed", fault.WithCause(err))
	}
	if conference == nil {
		return nil, fault.NotFound("Conference not found")
	}
	if strings.ToLower(strings.TrimSpace(conference.Provider)) != "jitsi" || s.jitsi == nil {
		return nil, fault.Unavailable(fmt.Sprintf("Conference join flow is not implemented for provider %s", conference.Provider))
	}
	access, _, err := s.requireChannelAccess(ctx, conference.ChannelID, cmd.Actor)
	if err != nil {
		return nil, err
	}
	moderator := cmd.Actor.IsAccount() && (access.CanManage || access.CanModerate)
	displayName, err := s.displayNameForPrincipal(ctx, cmd.Actor)
	if err != nil {
		return nil, err
	}
	expiresAt := s.now().Add(2 * time.Hour)
	token, err := s.jitsi.GenerateJoinToken(ctx, conference.JitsiRoomName, displayName, moderator, expiresAt)
	if err != nil {
		return nil, fault.Internal("Generate conference join token failed", fault.WithCause(err))
	}
	role := "participant"
	if moderator {
		role = "moderator"
	}
	_ = s.repo.AddConferenceParticipant(ctx, conference.ID, actorTypeFromPrincipal(cmd.Actor), principalAccountPtr(cmd.Actor), principalGuestPtr(cmd.Actor), role, s.now())
	return &CreateConferenceJoinTokenResult{Token: token, RoomName: conference.JitsiRoomName, JoinURL: s.jitsi.JoinURL(conference.JitsiRoomName, token), ExpiresAt: expiresAt}, nil
}

func (s *Service) StartConferenceRecording(ctx context.Context, cmd UpdateConferenceRecordingCmd) (*collabdomain.Conference, error) {
	conference, err := s.loadManageableConference(ctx, cmd.ConferenceID, cmd.Actor)
	if err != nil {
		return nil, err
	}
	now := s.now()
	updated, err := s.repo.UpdateConference(ctx, conference.ID, map[string]any{"recording_started_at": now, "updated_at": now, "updated_by": principalAccountPtr(cmd.Actor)})
	if err != nil {
		return nil, fault.Internal("Start conference recording failed", fault.WithCause(err))
	}
	_, _ = s.repo.CreateMessage(ctx, collabdomain.Message{ID: uuid.New(), ChannelID: conference.ChannelID, Type: collabdomain.MessageTypeSystem, AuthorType: collabdomain.ActorTypeSystem, Body: fmt.Sprintf("Recording started for conference: %s", conference.Title), CreatedAt: now}, nil, nil)
	s.publish(ctx, collabdomain.Event{Type: "conference.recording.started", ChannelID: conference.ChannelID, Payload: updated})
	return updated, nil
}

func (s *Service) StopConferenceRecording(ctx context.Context, cmd UpdateConferenceRecordingCmd) (*collabdomain.Conference, error) {
	conference, err := s.loadManageableConference(ctx, cmd.ConferenceID, cmd.Actor)
	if err != nil {
		return nil, err
	}
	now := s.now()
	updated, err := s.repo.UpdateConference(ctx, conference.ID, map[string]any{"recording_stopped_at": now, "updated_at": now, "updated_by": principalAccountPtr(cmd.Actor)})
	if err != nil {
		return nil, fault.Internal("Stop conference recording failed", fault.WithCause(err))
	}
	_, _ = s.repo.CreateMessage(ctx, collabdomain.Message{ID: uuid.New(), ChannelID: conference.ChannelID, Type: collabdomain.MessageTypeSystem, AuthorType: collabdomain.ActorTypeSystem, Body: fmt.Sprintf("Recording stopped for conference: %s", conference.Title), CreatedAt: now}, nil, nil)
	s.publish(ctx, collabdomain.Event{Type: "conference.recording.stopped", ChannelID: conference.ChannelID, Payload: updated})
	return updated, nil
}

func (s *Service) GetConferenceTranscript(ctx context.Context, q GetConferenceTranscriptQuery) (*collabdomain.ConferenceTranscript, error) {
	conference, err := s.repo.GetConferenceByID(ctx, q.ConferenceID)
	if err != nil {
		return nil, fault.Internal("Load conference failed", fault.WithCause(err))
	}
	if conference == nil {
		return nil, fault.NotFound("Conference not found")
	}
	if _, _, err := s.requireChannelAccess(ctx, conference.ChannelID, q.Actor); err != nil {
		return nil, err
	}
	transcript, err := s.repo.GetConferenceTranscript(ctx, q.ConferenceID)
	if err != nil {
		return nil, fault.Internal("Load transcript failed", fault.WithCause(err))
	}
	if transcript == nil {
		return nil, fault.NotFound("Transcript not found")
	}
	return transcript, nil
}

func (s *Service) HandleJitsiWebhook(ctx context.Context, cmd JitsiWebhookCmd) error {
	eventID, inserted, err := s.repo.CreateJitsiWebhookInbox(ctx, cmd.ProviderEventID, cmd.EventType, cmd.Payload, s.now())
	if err != nil {
		return fault.Internal("Persist Jitsi webhook failed", fault.WithCause(err))
	}
	if !inserted {
		return nil
	}

	var payload jitsiWebhookPayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		message := err.Error()
		_ = s.repo.MarkJitsiWebhookProcessed(ctx, eventID, s.now(), &message)
		return fault.Validation("Invalid Jitsi webhook payload")
	}
	conferenceID, err := uuid.Parse(strings.TrimSpace(payload.ConferenceID))
	if err != nil || conferenceID == uuid.Nil {
		message := "conferenceId is required"
		_ = s.repo.MarkJitsiWebhookProcessed(ctx, eventID, s.now(), &message)
		return fault.Validation("conferenceId is required")
	}
	conference, err := s.repo.GetConferenceByID(ctx, conferenceID)
	if err != nil {
		return fault.Internal("Load conference failed", fault.WithCause(err))
	}
	if conference == nil {
		return fault.NotFound("Conference not found")
	}
	if strings.ToLower(strings.TrimSpace(conference.Provider)) != "jitsi" {
		message := fmt.Sprintf("conference provider %s does not accept jitsi events", conference.Provider)
		_ = s.repo.MarkJitsiWebhookProcessed(ctx, eventID, s.now(), &message)
		return fault.Validation("Conference provider does not accept Jitsi events")
	}
	occurredAt := s.now()
	if payload.OccurredAt != nil {
		occurredAt = payload.OccurredAt.UTC()
	}

	switch cmd.EventType {
	case "conference.live":
		_, err = s.repo.UpdateConference(ctx, conference.ID, map[string]any{"status": string(collabdomain.ConferenceStatusLive), "started_at": occurredAt, "updated_at": occurredAt})
		if err == nil {
			_, _ = s.repo.CreateMessage(ctx, collabdomain.Message{ID: uuid.New(), ChannelID: conference.ChannelID, Type: collabdomain.MessageTypeSystem, AuthorType: collabdomain.ActorTypeSystem, Body: fmt.Sprintf("Conference live: %s", conference.Title), CreatedAt: occurredAt}, nil, nil)
		}
	case "conference.ended":
		_, err = s.repo.UpdateConference(ctx, conference.ID, map[string]any{"status": string(collabdomain.ConferenceStatusEnded), "ended_at": occurredAt, "updated_at": occurredAt})
		if err == nil {
			_, _ = s.repo.CreateMessage(ctx, collabdomain.Message{ID: uuid.New(), ChannelID: conference.ChannelID, Type: collabdomain.MessageTypeSystem, AuthorType: collabdomain.ActorTypeSystem, Body: fmt.Sprintf("Conference ended: %s", conference.Title), CreatedAt: occurredAt}, nil, nil)
		}
	case "recording.ready":
		if payload.Recording == nil {
			message := "recording payload is required"
			_ = s.repo.MarkJitsiWebhookProcessed(ctx, eventID, occurredAt, &message)
			return fault.Validation("recording payload is required")
		}
		object, duration, mimeType, buildErr := buildRecordingObject(*payload.Recording, occurredAt)
		if buildErr != nil {
			message := buildErr.Error()
			_ = s.repo.MarkJitsiWebhookProcessed(ctx, eventID, occurredAt, &message)
			return fault.Validation(buildErr.Error())
		}
		recording, recErr := s.repo.CreateConferenceRecording(ctx, conference.ID, object, nil, duration, mimeType, occurredAt)
		if recErr != nil {
			err = recErr
			break
		}
		if conference.RecordingEnabled {
			_ = s.repo.EnqueueTranscriptionJob(ctx, conference.ID, recording.ID, occurredAt)
			_, _ = s.repo.UpdateConference(ctx, conference.ID, map[string]any{"transcription_status": string(collabdomain.TranscriptionStatusPending), "updated_at": occurredAt})
		}
		_, _ = s.repo.CreateMessage(ctx, collabdomain.Message{ID: uuid.New(), ChannelID: conference.ChannelID, Type: collabdomain.MessageTypeSystem, AuthorType: collabdomain.ActorTypeSystem, Body: fmt.Sprintf("Recording ready for conference: %s", conference.Title), CreatedAt: occurredAt}, nil, nil)
	case "transcript.ready":
		if payload.Transcript == nil {
			message := "transcript payload is required"
			_ = s.repo.MarkJitsiWebhookProcessed(ctx, eventID, occurredAt, &message)
			return fault.Validation("transcript payload is required")
		}
		transcript := collabdomain.ConferenceTranscript{ConferenceID: conference.ID, TranscriptText: payload.Transcript.Text, SegmentsJSON: payload.Transcript.SegmentsJSON, LanguageCode: payload.Transcript.LanguageCode}
		if err = s.repo.UpsertConferenceTranscript(ctx, transcript, occurredAt); err == nil {
			_, _ = s.repo.UpdateConference(ctx, conference.ID, map[string]any{"transcription_status": string(collabdomain.TranscriptionStatusReady), "updated_at": occurredAt})
			_, _ = s.repo.CreateMessage(ctx, collabdomain.Message{ID: uuid.New(), ChannelID: conference.ChannelID, Type: collabdomain.MessageTypeSystem, AuthorType: collabdomain.ActorTypeSystem, Body: fmt.Sprintf("Transcript ready for conference: %s", conference.Title), CreatedAt: occurredAt}, nil, nil)
		}
	default:
		message := "unsupported event type"
		_ = s.repo.MarkJitsiWebhookProcessed(ctx, eventID, occurredAt, &message)
		return fault.Validation("Unsupported Jitsi webhook event")
	}
	if err != nil {
		message := err.Error()
		_ = s.repo.MarkJitsiWebhookProcessed(ctx, eventID, occurredAt, &message)
		return fault.Internal("Process Jitsi webhook failed", fault.WithCause(err))
	}
	return s.repo.MarkJitsiWebhookProcessed(ctx, eventID, occurredAt, nil)
}

func (s *Service) ProcessNextTranscriptionJob(ctx context.Context) (bool, error) {
	if s.transcriber == nil || s.storage == nil {
		return false, nil
	}
	lease, err := s.repo.LeaseNextTranscriptionJob(ctx, s.now(), 2*time.Minute)
	if err != nil {
		return false, fault.Internal("Lease transcription job failed", fault.WithCause(err))
	}
	if lease == nil {
		return false, nil
	}
	body, err := s.storage.ReadObject(ctx, lease.Bucket, lease.ObjectKey)
	if err != nil {
		_ = s.repo.FailTranscriptionJob(ctx, lease.JobID, err.Error(), s.now().Add(30*time.Second))
		return true, fault.Internal("Read recording object failed", fault.WithCause(err))
	}
	defer body.Close()
	result, err := s.transcriber.Transcribe(ctx, lease.FileName, lease.MimeType, body)
	if err != nil {
		_ = s.repo.FailTranscriptionJob(ctx, lease.JobID, err.Error(), s.now().Add(30*time.Second))
		return true, fault.Internal("Transcription failed", fault.WithCause(err))
	}
	text := strings.TrimSpace(result.Text)
	if text == "" {
		text = "(empty transcript)"
	}
	transcript := collabdomain.ConferenceTranscript{ConferenceID: lease.ConferenceID, TranscriptText: text, SegmentsJSON: result.SegmentsJSON, LanguageCode: result.LanguageCode}
	if err := s.repo.UpsertConferenceTranscript(ctx, transcript, s.now()); err != nil {
		_ = s.repo.FailTranscriptionJob(ctx, lease.JobID, err.Error(), s.now().Add(30*time.Second))
		return true, fault.Internal("Store transcript failed", fault.WithCause(err))
	}
	_, _ = s.repo.UpdateConference(ctx, lease.ConferenceID, map[string]any{"transcription_status": string(collabdomain.TranscriptionStatusReady), "updated_at": s.now()})
	_ = s.repo.CompleteTranscriptionJob(ctx, lease.JobID, s.now())
	conference, _ := s.repo.GetConferenceByID(ctx, lease.ConferenceID)
	if conference != nil {
		_, _ = s.repo.CreateMessage(ctx, collabdomain.Message{ID: uuid.New(), ChannelID: conference.ChannelID, Type: collabdomain.MessageTypeSystem, AuthorType: collabdomain.ActorTypeSystem, Body: fmt.Sprintf("Transcript ready for conference: %s", conference.Title), CreatedAt: s.now()}, nil, nil)
		s.publish(ctx, collabdomain.Event{Type: "conference.transcript.ready", ChannelID: conference.ChannelID, Payload: map[string]any{"conferenceId": lease.ConferenceID}})
	}
	return true, nil
}

func (s *Service) GetConferenceByID(ctx context.Context, id uuid.UUID, actor authdomain.Principal) (*collabdomain.Conference, error) {
	conference, err := s.repo.GetConferenceByID(ctx, id)
	if err != nil {
		return nil, fault.Internal("Load conference failed", fault.WithCause(err))
	}
	if conference == nil {
		return nil, fault.NotFound("Conference not found")
	}
	if _, _, err := s.requireChannelAccess(ctx, conference.ChannelID, actor); err != nil {
		return nil, err
	}
	return conference, nil
}

func (s *Service) displayNameForPrincipal(ctx context.Context, principal authdomain.Principal) (string, error) {
	if principal.IsGuest() {
		guest, err := s.repo.GetGuestIdentityByID(ctx, principal.GuestID)
		if err != nil {
			return "Guest", fault.Internal("Load guest identity failed", fault.WithCause(err))
		}
		if guest != nil && strings.TrimSpace(guest.DisplayName) != "" {
			return guest.DisplayName, nil
		}
		return "Guest", nil
	}
	if principal.IsAccount() && s.accounts != nil {
		accountID, err := accdomain.AccountIDFromUUID(principal.AccountID)
		if err == nil {
			acc, accErr := s.accounts.GetByID(ctx, accountID)
			if accErr != nil {
				return "Account", fault.Internal("Load account failed", fault.WithCause(accErr))
			}
			if acc != nil {
				if name := acc.DisplayName(); name != nil && strings.TrimSpace(*name) != "" {
					return *name, nil
				}
				return acc.Email().String(), nil
			}
		}
	}
	return "Account", nil
}

func (s *Service) loadManageableConference(ctx context.Context, conferenceID uuid.UUID, actor authdomain.Principal) (*collabdomain.Conference, error) {
	conference, err := s.repo.GetConferenceByID(ctx, conferenceID)
	if err != nil {
		return nil, fault.Internal("Load conference failed", fault.WithCause(err))
	}
	if conference == nil {
		return nil, fault.NotFound("Conference not found")
	}
	access, _, err := s.requireChannelAccess(ctx, conference.ChannelID, actor)
	if err != nil {
		return nil, err
	}
	if !actor.IsAccount() || !access.CanManage {
		return nil, fault.Forbidden("Conference management is not allowed")
	}
	return conference, nil
}

func buildConferenceRoomName(groupID, channelID uuid.UUID) string {
	return strings.ToLower(fmt.Sprintf("grp-%s-ch-%s-%s", shortUUID(groupID), shortUUID(channelID), shortUUID(uuid.New())))
}

func buildRecordingObject(payload jitsiRecordingPayload, now time.Time) (collabdomain.StorageObject, *int32, *string, error) {
	if strings.TrimSpace(payload.Bucket) == "" || strings.TrimSpace(payload.ObjectKey) == "" || strings.TrimSpace(payload.FileName) == "" {
		return collabdomain.StorageObject{}, nil, nil, fmt.Errorf("recording bucket, objectKey and fileName are required")
	}
	var organizationID *uuid.UUID
	if payload.OrganizationID != nil && strings.TrimSpace(*payload.OrganizationID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*payload.OrganizationID))
		if err != nil {
			return collabdomain.StorageObject{}, nil, nil, fmt.Errorf("recording organizationId is invalid")
		}
		organizationID = &parsed
	}
	return collabdomain.StorageObject{ID: uuid.New(), OrganizationID: organizationID, Bucket: strings.TrimSpace(payload.Bucket), ObjectKey: strings.TrimSpace(payload.ObjectKey), FileName: sanitizeFileName(payload.FileName), ContentType: normalizeOptional(payload.ContentType), SizeBytes: payload.SizeBytes, ChecksumSHA256: normalizeOptional(payload.ChecksumSHA256), CreatedAt: now}, payload.DurationSec, normalizeOptional(payload.ContentType), nil
}

func chooseTranscriptionStatus(recordingEnabled bool) collabdomain.TranscriptionStatus {
	if !recordingEnabled {
		return collabdomain.TranscriptionStatusDisabled
	}
	return collabdomain.TranscriptionStatusPending
}
