package http

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	authapp "github.com/NikolayNam/collabsphere/internal/collab/application"
	collabdomain "github.com/NikolayNam/collabsphere/internal/collab/domain"
	"github.com/NikolayNam/collabsphere/internal/collab/realtime"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/google/uuid"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type WebSocketHandler struct {
	svc             *authapp.Service
	verifier        authmw.AccessTokenVerifier
	broker          realtime.Broker
	allowQueryToken bool
}

func NewWebSocketHandler(svc *authapp.Service, verifier authmw.AccessTokenVerifier, broker realtime.Broker, allowQueryToken bool) *WebSocketHandler {
	return &WebSocketHandler{svc: svc, verifier: verifier, broker: broker, allowQueryToken: allowQueryToken}
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(strings.TrimSpace(r.URL.Query().Get("channel_id")))
	if err != nil || channelID == uuid.Nil {
		http.Error(w, "channel_id is required", http.StatusBadRequest)
		return
	}
	principal := authenticateWSPrincipal(r, h.verifier, h.allowQueryToken)
	if !principal.Authenticated {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if err := h.svc.AuthorizeChannel(r.Context(), channelID, principal); err != nil {
		http.Error(w, "channel access denied", http.StatusForbidden)
		return
	}
	if !websocketRequested(r) {
		http.Error(w, "websocket upgrade required", http.StatusUpgradeRequired)
		return
	}
	key := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if key == "" {
		http.Error(w, "missing websocket key", http.StatusBadRequest)
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket unsupported", http.StatusInternalServerError)
		return
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "websocket upgrade failed", http.StatusInternalServerError)
		return
	}
	if err := writeHandshake(rw, key); err != nil {
		_ = conn.Close()
		return
	}
	events, unsubscribe := h.broker.Subscribe(channelID)
	defer unsubscribe()
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	go func() {
		defer cancel()
		_ = consumeClientFrames(conn)
	}()

	for {
		select {
		case <-ctx.Done():
			_ = writeCloseFrame(conn)
			return
		case event := <-events:
			if err := writeJSONFrame(conn, event); err != nil {
				return
			}
		}
	}
}

func authenticateWSPrincipal(r *http.Request, verifier authmw.AccessTokenVerifier, allowQueryToken bool) authdomain.Principal {
	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if authz == "" && allowQueryToken {
		if token := strings.TrimSpace(r.URL.Query().Get("access_token")); token != "" {
			authz = "Bearer " + token
		}
	}
	if verifier == nil || authz == "" {
		return authdomain.AnonymousPrincipal()
	}
	token := wsExtractBearer(authz)
	if token == "" {
		return authdomain.AnonymousPrincipal()
	}
	principal, err := verifier.VerifyAccessToken(r.Context(), token)
	if err != nil {
		return authdomain.AnonymousPrincipal()
	}
	return principal
}

func websocketRequested(r *http.Request) bool {
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket") && strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

func writeHandshake(rw *bufio.ReadWriter, key string) error {
	accept := websocketAccept(key)
	_, err := rw.WriteString("HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n")
	if err != nil {
		return err
	}
	return rw.Flush()
}

func websocketAccept(key string) string {
	h := sha1.Sum([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(h[:])
}

func consumeClientFrames(conn net.Conn) error {
	for {
		opcode, payload, err := readFrame(conn)
		if err != nil {
			return err
		}
		switch opcode {
		case 0x8:
			return nil
		case 0x9:
			if err := writeControlFrame(conn, 0xA, payload); err != nil {
				return err
			}
		}
	}
}

func writeJSONFrame(conn net.Conn, event collabdomain.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return writeFrame(conn, 0x1, payload)
}

func writeCloseFrame(conn net.Conn) error {
	return writeControlFrame(conn, 0x8, nil)
}

func writeControlFrame(conn net.Conn, opcode byte, payload []byte) error {
	return writeFrame(conn, opcode, payload)
}

func writeFrame(conn net.Conn, opcode byte, payload []byte) error {
	if len(payload) > 65535 {
		return fmt.Errorf("payload too large")
	}
	head := []byte{0x80 | opcode}
	switch {
	case len(payload) < 126:
		head = append(head, byte(len(payload)))
	case len(payload) <= 65535:
		head = append(head, 126, byte(len(payload)>>8), byte(len(payload)))
	}
	if _, err := conn.Write(head); err != nil {
		return err
	}
	if len(payload) == 0 {
		return nil
	}
	_, err := conn.Write(payload)
	return err
}

func readFrame(conn net.Conn) (byte, []byte, error) {
	head := make([]byte, 2)
	if _, err := io.ReadFull(conn, head); err != nil {
		return 0, nil, err
	}
	opcode := head[0] & 0x0F
	masked := head[1]&0x80 != 0
	length := int(head[1] & 0x7F)
	if length == 126 {
		ext := make([]byte, 2)
		if _, err := io.ReadFull(conn, ext); err != nil {
			return 0, nil, err
		}
		length = int(ext[0])<<8 | int(ext[1])
	} else if length == 127 {
		return 0, nil, errors.New("unsupported websocket frame length")
	}
	var maskKey []byte
	if masked {
		maskKey = make([]byte, 4)
		if _, err := io.ReadFull(conn, maskKey); err != nil {
			return 0, nil, err
		}
	}
	payload := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(conn, payload); err != nil {
			return 0, nil, err
		}
	}
	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}
	return opcode, payload, nil
}

func wsExtractBearer(v string) string {
	const prefix = "Bearer "
	if len(v) < len(prefix)+1 || !strings.EqualFold(v[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(v[len(prefix):])
}
