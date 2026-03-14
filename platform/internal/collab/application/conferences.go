package application

import (
	"context"
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
		RoomName:            buildConferenceRoomName(channel.GroupID, channel.ID),
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
	s.bestEffort(ctx, "conference.create.system_message", func() error {
		_, err := s.repo.CreateMessage(ctx, collabdomain.Message{
			ID:         uuid.New(),
			ChannelID:  channel.ID,
			Type:       collabdomain.MessageTypeSystem,
			AuthorType: collabdomain.ActorTypeSystem,
			Body:       fmt.Sprintf("Conference scheduled: %s", conference.Title),
			CreatedAt:  now,
		}, nil, nil)
		return err
	})
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
	if _, _, err := s.requireChannelAccess(ctx, conference.ChannelID, cmd.Actor); err != nil {
		return nil, err
	}
	return nil, fault.Unavailable(fmt.Sprintf("Conference join flow is not implemented for provider %s", conference.Provider))
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
	s.bestEffort(ctx, "conference.recording.started.system_message", func() error {
		_, err := s.repo.CreateMessage(ctx, collabdomain.Message{
			ID:         uuid.New(),
			ChannelID:  conference.ChannelID,
			Type:       collabdomain.MessageTypeSystem,
			AuthorType: collabdomain.ActorTypeSystem,
			Body:       fmt.Sprintf("Recording started for conference: %s", conference.Title),
			CreatedAt:  now,
		}, nil, nil)
		return err
	})
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
	s.bestEffort(ctx, "conference.recording.stopped.system_message", func() error {
		_, err := s.repo.CreateMessage(ctx, collabdomain.Message{
			ID:         uuid.New(),
			ChannelID:  conference.ChannelID,
			Type:       collabdomain.MessageTypeSystem,
			AuthorType: collabdomain.ActorTypeSystem,
			Body:       fmt.Sprintf("Recording stopped for conference: %s", conference.Title),
			CreatedAt:  now,
		}, nil, nil)
		return err
	})
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
		s.bestEffort(ctx, "conference.transcription.fail_job.read_object", func() error {
			return s.repo.FailTranscriptionJob(ctx, lease.JobID, err.Error(), s.now().Add(30*time.Second))
		})
		return true, fault.Internal("Read recording object failed", fault.WithCause(err))
	}
	defer body.Close()
	result, err := s.transcriber.Transcribe(ctx, lease.FileName, lease.MimeType, body)
	if err != nil {
		s.bestEffort(ctx, "conference.transcription.fail_job.transcribe", func() error {
			return s.repo.FailTranscriptionJob(ctx, lease.JobID, err.Error(), s.now().Add(30*time.Second))
		})
		return true, fault.Internal("Transcription failed", fault.WithCause(err))
	}
	text := strings.TrimSpace(result.Text)
	if text == "" {
		text = "(empty transcript)"
	}
	transcript := collabdomain.ConferenceTranscript{ConferenceID: lease.ConferenceID, TranscriptText: text, SegmentsJSON: result.SegmentsJSON, LanguageCode: result.LanguageCode}
	if err := s.repo.UpsertConferenceTranscript(ctx, transcript, s.now()); err != nil {
		s.bestEffort(ctx, "conference.transcription.fail_job.upsert_transcript", func() error {
			return s.repo.FailTranscriptionJob(ctx, lease.JobID, err.Error(), s.now().Add(30*time.Second))
		})
		return true, fault.Internal("Store transcript failed", fault.WithCause(err))
	}
	s.bestEffort(ctx, "conference.transcription.update_status", func() error {
		_, err := s.repo.UpdateConference(ctx, lease.ConferenceID, map[string]any{
			"transcription_status": string(collabdomain.TranscriptionStatusReady),
			"updated_at":           s.now(),
		})
		return err
	})
	s.bestEffort(ctx, "conference.transcription.complete_job", func() error {
		return s.repo.CompleteTranscriptionJob(ctx, lease.JobID, s.now())
	})
	conference, conferenceErr := s.repo.GetConferenceByID(ctx, lease.ConferenceID)
	if conferenceErr != nil {
		s.bestEffort(ctx, "conference.transcription.load_conference", func() error {
			return conferenceErr
		})
	}
	if conference != nil {
		s.bestEffort(ctx, "conference.transcription.ready.system_message", func() error {
			_, err := s.repo.CreateMessage(ctx, collabdomain.Message{
				ID:         uuid.New(),
				ChannelID:  conference.ChannelID,
				Type:       collabdomain.MessageTypeSystem,
				AuthorType: collabdomain.ActorTypeSystem,
				Body:       fmt.Sprintf("Transcript ready for conference: %s", conference.Title),
				CreatedAt:  s.now(),
			}, nil, nil)
			return err
		})
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

func chooseTranscriptionStatus(recordingEnabled bool) collabdomain.TranscriptionStatus {
	if !recordingEnabled {
		return collabdomain.TranscriptionStatusDisabled
	}
	return collabdomain.TranscriptionStatusPending
}
