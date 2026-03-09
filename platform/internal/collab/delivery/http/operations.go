package http

import "github.com/danielgtaylor/huma/v2"

var createChannelOp = huma.Operation{
	OperationID: "create-channel",
	Method:      "POST",
	Path:        "/groups/{group_id}/channels",
	Tags:        []string{"Collab / Channels"},
	Summary:     "Create group channel",
	Description: "Creates a channel inside the target group. The group remains the ACL container, while the channel becomes the container for messages, attachments, guest invites, and conferences.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listChannelsOp = huma.Operation{
	OperationID: "list-channels",
	Method:      "GET",
	Path:        "/groups/{group_id}/channels",
	Tags:        []string{"Collab / Channels"},
	Summary:     "List group channels",
	Description: "Returns the channels visible inside the group for the authenticated actor.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createMessageOp = huma.Operation{
	OperationID: "create-message",
	Method:      "POST",
	Path:        "/channels/{channel_id}/messages",
	Tags:        []string{"Collab / Channels"},
	Summary:     "Create channel message",
	Description: "Creates a message in the channel. Messages may include replies, mentions, and previously uploaded attachment object ids.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listMessagesOp = huma.Operation{
	OperationID: "list-messages",
	Method:      "GET",
	Path:        "/channels/{channel_id}/messages",
	Tags:        []string{"Collab / Channels"},
	Summary:     "List channel messages",
	Description: "Returns the message timeline for the channel, including visible attachments, mentions, and edit state.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateMessageOp = huma.Operation{
	OperationID: "update-message",
	Method:      "PATCH",
	Path:        "/channels/{channel_id}/messages/{message_id}",
	Tags:        []string{"Collab / Channels"},
	Summary:     "Update channel message",
	Description: "Edits an existing message in the channel and records revision history where supported by the domain model.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var deleteMessageOp = huma.Operation{
	OperationID: "delete-message",
	Method:      "DELETE",
	Path:        "/channels/{channel_id}/messages/{message_id}",
	Tags:        []string{"Collab / Channels"},
	Summary:     "Delete channel message",
	Description: "Deletes a message from the channel according to the actor's moderation or ownership permissions.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var uploadAttachmentOp = huma.Operation{
	OperationID: "upload-channel-attachment",
	Method:      "POST",
	Path:        "/channels/{channel_id}/attachments/upload",
	Tags:        []string{"Collab / Files"},
	Summary:     "Upload channel attachment directly",
	Description: "Single-step channel attachment upload using multipart/form-data. Send the file in the `file` field and optional `organizationId` field. The backend uploads the object to S3-compatible storage and returns the stored attachment metadata so its objectId can be used in create-message.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var updateReadCursorOp = huma.Operation{
	OperationID: "update-read-cursor",
	Method:      "POST",
	Path:        "/channels/{channel_id}/read-cursor",
	Tags:        []string{"Collab / Channels"},
	Summary:     "Update channel read cursor",
	Description: "Moves the actor's read cursor forward in the channel so unread counters and read state can be calculated consistently.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var toggleReactionOp = huma.Operation{
	OperationID: "toggle-message-reaction",
	Method:      "POST",
	Path:        "/channels/{channel_id}/reactions",
	Tags:        []string{"Collab / Channels"},
	Summary:     "Toggle message reaction",
	Description: "Adds or removes an emoji reaction for a message in the channel.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createGuestInviteOp = huma.Operation{
	OperationID: "create-guest-invite",
	Method:      "POST",
	Path:        "/channels/{channel_id}/guest-invites",
	Tags:        []string{"Collab / Channels"},
	Summary:     "Create guest invite for channel",
	Description: "Creates a guest invite scoped to a single channel. The resulting magic link can later be exchanged for a guest session.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var exchangeGuestInviteOp = huma.Operation{
	OperationID: "exchange-guest-invite",
	Method:      "POST",
	Path:        "/guest-invites/{token}/exchange",
	Tags:        []string{"Collab / Channels"},
	Summary:     "Exchange guest invite magic link",
	Description: "Exchanges a guest invite token for a guest session scoped to the invited channel and its conferences.",
}

var createConferenceOp = huma.Operation{
	OperationID: "create-conference",
	Method:      "POST",
	Path:        "/channels/{channel_id}/conferences",
	Tags:        []string{"Collab / Conferences"},
	Summary:     "Create channel conference",
	Description: "Creates a conference linked to the channel. Conferences have their own lifecycle, recordings, and transcript flow.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listConferencesOp = huma.Operation{
	OperationID: "list-conferences",
	Method:      "GET",
	Path:        "/channels/{channel_id}/conferences",
	Tags:        []string{"Collab / Conferences"},
	Summary:     "List channel conferences",
	Description: "Returns the conferences that belong to the channel, including current status and lifecycle timestamps.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createConferenceJoinTokenOp = huma.Operation{
	OperationID: "create-conference-join-token",
	Method:      "POST",
	Path:        "/conferences/{conference_id}/join-token",
	Tags:        []string{"Collab / Conferences"},
	Summary:     "Create conference join token",
	Description: "Requests a conference join token. The conference provider is mediasoup, but signaling and join-token issuance are not implemented yet, so the endpoint currently returns 503.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var startConferenceRecordingOp = huma.Operation{
	OperationID: "start-conference-recording",
	Method:      "POST",
	Path:        "/conferences/{conference_id}/recording/start",
	Tags:        []string{"Collab / Conferences"},
	Summary:     "Mark conference recording start",
	Description: "Marks the beginning of conference recording in the platform state. Media ingestion and actual recording backend orchestration are handled separately.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var stopConferenceRecordingOp = huma.Operation{
	OperationID: "stop-conference-recording",
	Method:      "POST",
	Path:        "/conferences/{conference_id}/recording/stop",
	Tags:        []string{"Collab / Conferences"},
	Summary:     "Mark conference recording stop",
	Description: "Marks the end of conference recording in the platform state so recordings can be finalized and exposed in the files subsection.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getConferenceTranscriptOp = huma.Operation{
	OperationID: "get-conference-transcript",
	Method:      "GET",
	Path:        "/conferences/{conference_id}/transcript",
	Tags:        []string{"Collab / Conferences"},
	Summary:     "Get conference transcript",
	Description: "Returns the current transcript for the conference when recording and transcription have produced one.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
