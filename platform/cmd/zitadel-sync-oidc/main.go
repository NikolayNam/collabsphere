package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func logf(format string, args ...any) {
	fmt.Printf("[zitadel-sync-oidc] "+format+"\n", args...)
}

func resolvePath(envKey string, candidates ...string) string {
	if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
		return filepath.Clean(v)
	}
	for _, c := range candidates {
		p := filepath.Clean(c)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if len(candidates) > 0 {
		return filepath.Clean(candidates[0])
	}
	return ""
}

func parseEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return out, nil
}

func upsertEnvKey(path, key, value string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(data), "\n")
	needle := key + "="
	replaced := false
	changed := false
	for i, line := range lines {
		if strings.HasPrefix(line, needle) {
			replaced = true
			newLine := needle + value
			if line != newLine {
				lines[i] = newLine
				changed = true
			}
		}
	}
	if !replaced {
		lines = append(lines, needle+value)
		changed = true
	}
	if !changed {
		return false, nil
	}
	result := strings.Join(lines, "\n")
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	if err := os.WriteFile(path, []byte(result), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func readSecret(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func writeSecret(path, secret string) (bool, error) {
	current, err := readSecret(path)
	if err == nil && current == secret {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	if err := os.WriteFile(path, []byte(secret+"\n"), 0o600); err != nil {
		return false, err
	}
	return true, nil
}

func requestJSON(
	client *http.Client,
	method, url, token string,
	payload any,
) (map[string]any, error) {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if strings.Contains(url, "/management/") &&
			(resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
			return nil, fmt.Errorf(
				"%s %s failed with %d: %s\n%s",
				method,
				url,
				resp.StatusCode,
				string(raw),
				describeManagementAuthFailure(resp.StatusCode),
			)
		}
		return nil, fmt.Errorf("%s %s failed with %d: %s", method, url, resp.StatusCode, string(raw))
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return map[string]any{}, nil
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%s %s returned non-json: %w", method, url, err)
	}
	return out, nil
}

func describeManagementAuthFailure(statusCode int) string {
	switch statusCode {
	case http.StatusUnauthorized:
		return "ZITADEL management API rejected admin PAT (401).\n" +
			"Check AUTH_ZITADEL_ADMIN_TOKEN or deploy/secrets/identity/zitadel_admin_token: token must be a valid raw PAT for the current instance (no JSON wrapper, no client_secret/password)."
	case http.StatusForbidden:
		return "ZITADEL management API denied permissions for admin PAT (403).\n" +
			"Use a PAT from a service account/user that has rights to list/create OIDC apps in the target project."
	default:
		return "ZITADEL management API authentication/authorization failed."
	}
}

func parseOAuthError(raw []byte) string {
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(payload["error"]))
}

func validateExistingCredentials(
	client *http.Client,
	issuerURL, clientID, clientSecret string,
) (bool, string, error) {
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return false, "missing_client_credentials", nil
	}
	tokenURL := strings.TrimRight(issuerURL, "/") + "/oauth/v2/token"
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", "openid")
	req, err := http.NewRequest(
		http.MethodPost,
		tokenURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return false, "", err
	}
	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	req.SetBasicAuth(clientID, clientSecret)
	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, "ok", nil
	}
	oauthErr := parseOAuthError(raw)
	if oauthErr == "invalid_client" {
		return false, oauthErr, nil
	}
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		// Non-invalid_client errors (e.g. unauthorized_client) still mean
		// client authentication passed, so keep existing credentials.
		if oauthErr != "" {
			return true, oauthErr, nil
		}
	}
	return false, "", fmt.Errorf(
		"credential validation returned status %d: %s",
		resp.StatusCode,
		string(raw),
	)
}

func waitReady(client *http.Client, issuerURL string, timeout time.Duration) error {
	baseURL := strings.TrimRight(issuerURL, "/")
	readyURLs := []string{
		baseURL + "/ready",
		baseURL + "/debug/healthz",
		baseURL + "/.well-known/openid-configuration",
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, readyURL := range readyURLs {
			req, _ := http.NewRequest(http.MethodGet, readyURL, nil)
			resp, err := client.Do(req)
			if err == nil && resp != nil {
				resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					return nil
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("zitadel was not ready at %s within %s", strings.Join(readyURLs, ", "), timeout)
}

func asList(v any) []map[string]any {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func pickProject(projects []map[string]any, wanted string) map[string]any {
	for _, project := range projects {
		if strings.TrimSpace(fmt.Sprint(project["name"])) == wanted {
			return project
		}
	}
	if len(projects) > 0 {
		return projects[0]
	}
	return nil
}

func pickApp(apps []map[string]any, wanted string) map[string]any {
	for _, app := range apps {
		if strings.TrimSpace(fmt.Sprint(app["name"])) == wanted {
			return app
		}
	}
	return nil
}

func extractClientID(app map[string]any) string {
	if cfg, ok := app["oidcConfig"].(map[string]any); ok {
		for _, key := range []string{"clientId", "clientID", "client_id"} {
			if id := strings.TrimSpace(fmt.Sprint(cfg[key])); id != "" && id != "<nil>" {
				return id
			}
		}
	}
	if id := strings.TrimSpace(fmt.Sprint(app["clientId"])); id != "" && id != "<nil>" {
		return id
	}
	return ""
}

func resolveAdminToken(envValues map[string]string) (string, error) {
	if inline := strings.TrimSpace(envValues["AUTH_ZITADEL_ADMIN_TOKEN"]); inline != "" {
		return inline, nil
	}
	source := strings.TrimSpace(envValues["AUTH_ZITADEL_ADMIN_TOKEN_SOURCE_FILE"])
	candidates := []string{}
	if source != "" {
		candidates = append(candidates, resolvePath("", source, filepath.Join("deploy", source)))
	}
	candidates = append(
		candidates,
		filepath.Join("deploy", "secrets", "identity", "zitadel_admin_token"),
		filepath.Join("..", "deploy", "secrets", "identity", "zitadel_admin_token"),
	)
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			token, err := readSecret(p)
			if err == nil && token != "" {
				return token, nil
			}
		}
	}
	return "", errors.New("zitadel admin token not found")
}

func run() error {
	envPath := resolvePath(
		"ZITADEL_SYNC_ENV_FILE",
		filepath.Join("deploy", "env", ".env.zitadel.dev"),
		filepath.Join("..", "deploy", "env", ".env.zitadel.dev"),
	)
	secretPath := resolvePath(
		"ZITADEL_SYNC_SECRET_FILE",
		filepath.Join("deploy", "secrets", "identity", "zitadel_client_secret"),
		filepath.Join("..", "deploy", "secrets", "identity", "zitadel_client_secret"),
	)
	waitSeconds := 180
	if v := strings.TrimSpace(os.Getenv("ZITADEL_SYNC_WAIT_TIMEOUT_SEC")); v != "" {
		fmt.Sscanf(v, "%d", &waitSeconds)
	}
	projectName := strings.TrimSpace(os.Getenv("ZITADEL_SYNC_PROJECT_NAME"))
	if projectName == "" {
		projectName = "CollabSphere"
	}
	appName := strings.TrimSpace(os.Getenv("ZITADEL_SYNC_APP_NAME"))
	if appName == "" {
		appName = "CollabSphere Backend OIDC"
	}
	forceRegenerate := strings.EqualFold(strings.TrimSpace(os.Getenv("ZITADEL_SYNC_FORCE_REGENERATE")), "true") ||
		strings.TrimSpace(os.Getenv("ZITADEL_SYNC_FORCE_REGENERATE")) == "1"

	envValues, err := parseEnvFile(envPath)
	if err != nil {
		return fmt.Errorf("read env file: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(envValues["AUTH_ZITADEL_ENABLED"]), "true") {
		logf("AUTH_ZITADEL_ENABLED is not true; skipping")
		return nil
	}
	issuerURL := strings.TrimSpace(envValues["AUTH_ZITADEL_ISSUER_URL"])
	redirectURL := strings.TrimSpace(envValues["AUTH_ZITADEL_REDIRECT_URL"])
	if issuerURL == "" || redirectURL == "" {
		return errors.New("AUTH_ZITADEL_ISSUER_URL and AUTH_ZITADEL_REDIRECT_URL are required")
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err == nil && strings.HasSuffix(strings.ToLower(strings.TrimSpace(host)), ".localhost") {
					addr = net.JoinHostPort("127.0.0.1", port)
				}
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}
	if err := waitReady(client, issuerURL, time.Duration(waitSeconds)*time.Second); err != nil {
		return err
	}

	envClientID := strings.TrimSpace(envValues["AUTH_ZITADEL_CLIENT_ID"])
	existingSecret, existingSecretErr := readSecret(secretPath)
	if !forceRegenerate && envClientID != "" && existingSecretErr == nil && existingSecret != "" {
		valid, reason, err := validateExistingCredentials(
			client,
			issuerURL,
			envClientID,
			existingSecret,
		)
		if err != nil {
			return err
		}
		if valid {
			logf("existing OIDC credentials are valid (check=%s), skipping regenerate", reason)
			return nil
		}
		logf("existing OIDC credentials are not valid (check=%s), syncing via management API", reason)
	}

	adminToken, err := resolveAdminToken(envValues)
	if err != nil {
		return err
	}
	base := strings.TrimRight(issuerURL, "/")

	projectsResp, err := requestJSON(
		client,
		http.MethodPost,
		base+"/management/v1/projects/_search",
		adminToken,
		map[string]any{},
	)
	if err != nil {
		return err
	}
	project := pickProject(asList(projectsResp["result"]), projectName)
	if project == nil {
		return errors.New("no projects available in zitadel response")
	}
	projectID := strings.TrimSpace(fmt.Sprint(project["id"]))
	if projectID == "" || projectID == "<nil>" {
		return errors.New("project id is empty")
	}

	appsResp, err := requestJSON(
		client,
		http.MethodPost,
		base+"/management/v1/projects/"+projectID+"/apps/_search",
		adminToken,
		map[string]any{},
	)
	if err != nil {
		return err
	}
	app := pickApp(asList(appsResp["result"]), appName)

	clientID := ""
	clientSecret := ""
	appID := ""
	if app == nil {
		logf("OIDC app '%s' not found; creating", appName)
		created, err := requestJSON(
			client,
			http.MethodPost,
			base+"/management/v1/projects/"+projectID+"/apps/oidc",
			adminToken,
			map[string]any{
				"name":                   appName,
				"redirectUris":           []string{redirectURL},
				"responseTypes":          []string{"OIDC_RESPONSE_TYPE_CODE"},
				"grantTypes":             []string{"OIDC_GRANT_TYPE_AUTHORIZATION_CODE", "OIDC_GRANT_TYPE_REFRESH_TOKEN"},
				"appType":                "OIDC_APP_TYPE_WEB",
				"authMethodType":         "OIDC_AUTH_METHOD_TYPE_BASIC",
				"postLogoutRedirectUris": []string{redirectURL},
				"version":                "OIDC_VERSION_1_0",
				"devMode":                true,
			},
		)
		if err != nil {
			return err
		}
		appID = strings.TrimSpace(fmt.Sprint(created["appId"]))
		clientID = strings.TrimSpace(fmt.Sprint(created["clientId"]))
		clientSecret = strings.TrimSpace(fmt.Sprint(created["clientSecret"]))
	} else {
		appID = strings.TrimSpace(fmt.Sprint(app["id"]))
		clientID = extractClientID(app)
		if appID == "" || appID == "<nil>" || clientID == "" || clientID == "<nil>" {
			return errors.New("existing OIDC app response is missing id/client_id")
		}
		needGenerate := forceRegenerate
		generateReason := "force_regenerate"
		if !needGenerate {
			needGenerate = existingSecretErr != nil || existingSecret == ""
			if needGenerate {
				generateReason = "missing_local_secret"
			}
		}
		if !needGenerate {
			valid, reason, err := validateExistingCredentials(
				client,
				issuerURL,
				clientID,
				existingSecret,
			)
			if err != nil {
				return err
			}
			if !valid {
				needGenerate = true
				generateReason = reason
			} else {
				logf(
					"existing OIDC credentials are valid (check=%s), skipping regenerate",
					reason,
				)
				if envClientID != clientID {
					logf("syncing AUTH_ZITADEL_CLIENT_ID to current app client_id")
				}
				clientSecret = existingSecret
			}
		}
		if needGenerate {
			logf("regenerating OIDC client secret (reason=%s)", generateReason)
			genResp, err := requestJSON(
				client,
				http.MethodPost,
				base+"/management/v1/projects/"+projectID+"/apps/"+appID+"/oidc_config/_generate_client_secret",
				adminToken,
				nil,
			)
			if err != nil {
				return err
			}
			clientSecret = strings.TrimSpace(fmt.Sprint(genResp["clientSecret"]))
		}
	}
	if clientID == "" || clientID == "<nil>" {
		return errors.New("resolved empty client id")
	}
	if clientSecret == "" || clientSecret == "<nil>" {
		if existing, err := readSecret(secretPath); err == nil && existing != "" {
			clientSecret = existing
		} else {
			return errors.New("resolved empty client secret")
		}
	}

	envChanged, err := upsertEnvKey(envPath, "AUTH_ZITADEL_CLIENT_ID", clientID)
	if err != nil {
		return err
	}
	secretChanged, err := writeSecret(secretPath, clientSecret)
	if err != nil {
		return err
	}
	if envChanged || secretChanged {
		logf("updated local ZITADEL OIDC credentials")
	} else {
		logf("local ZITADEL OIDC credentials are up to date")
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "[zitadel-sync-oidc] ERROR: %v\n", err)
		os.Exit(1)
	}
}
