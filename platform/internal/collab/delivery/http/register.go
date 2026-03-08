package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	secured := huma.Middlewares{authmw.HumaAuthOptional(verifier)}

	createChannel := createChannelOp
	createChannel.Middlewares = secured
	huma.Register(api, createChannel, h.CreateChannel)

	listChannels := listChannelsOp
	listChannels.Middlewares = secured
	huma.Register(api, listChannels, h.ListChannels)

	createMessage := createMessageOp
	createMessage.Middlewares = secured
	huma.Register(api, createMessage, h.CreateMessage)

	listMessages := listMessagesOp
	listMessages.Middlewares = secured
	huma.Register(api, listMessages, h.ListMessages)

	updateMessage := updateMessageOp
	updateMessage.Middlewares = secured
	huma.Register(api, updateMessage, h.UpdateMessage)

	deleteMessage := deleteMessageOp
	deleteMessage.Middlewares = secured
	huma.Register(api, deleteMessage, h.DeleteMessage)

	attachUpload := createAttachmentUploadOp
	attachUpload.Middlewares = secured
	huma.Register(api, attachUpload, h.CreateAttachmentUpload)

	readCursor := updateReadCursorOp
	readCursor.Middlewares = secured
	huma.Register(api, readCursor, h.UpdateReadCursor)

	reaction := toggleReactionOp
	reaction.Middlewares = secured
	huma.Register(api, reaction, h.ToggleReaction)

	guestInvite := createGuestInviteOp
	guestInvite.Middlewares = secured
	huma.Register(api, guestInvite, h.CreateGuestInvite)
	huma.Register(api, exchangeGuestInviteOp, h.ExchangeGuestInvite)

	createConference := createConferenceOp
	createConference.Middlewares = secured
	huma.Register(api, createConference, h.CreateConference)

	listConferences := listConferencesOp
	listConferences.Middlewares = secured
	huma.Register(api, listConferences, h.ListConferences)

	joinToken := createConferenceJoinTokenOp
	joinToken.Middlewares = secured
	huma.Register(api, joinToken, h.CreateConferenceJoinToken)

	startRecording := startConferenceRecordingOp
	startRecording.Middlewares = secured
	huma.Register(api, startRecording, h.StartConferenceRecording)

	stopRecording := stopConferenceRecordingOp
	stopRecording.Middlewares = secured
	huma.Register(api, stopRecording, h.StopConferenceRecording)

	getTranscript := getConferenceTranscriptOp
	getTranscript.Middlewares = secured
	huma.Register(api, getTranscript, h.GetConferenceTranscript)

	huma.Register(api, jitsiWebhookOp, h.JitsiWebhook)
}
