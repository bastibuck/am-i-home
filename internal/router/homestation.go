package router

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// HomeStationClient implements RouterClient for Vodafone HomeStation-like routers
type HomeStationClient struct {
	baseURL string
	user    string
	pass    string
	client  *http.Client
}

func NewHomeStationClient(baseURL, user, pass string) (*HomeStationClient, error) {
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{Jar: jar, Timeout: 10 * time.Second}

	return &HomeStationClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		user:    user,
		pass:    pass,
		client:  httpClient,
	}, nil
}

// pbkdf2Hex computes PBKDF2-SHA256 and returns the result as lowercase hex
// iterations: 1000, keyLen: 16 bytes (128 bits) - matching the router's JS implementation
func pbkdf2Hex(password, salt string) string {
	key := pbkdf2.Key([]byte(password), []byte(salt), 1000, 16, sha256.New)
	return hex.EncodeToString(key)
}

// activateSession makes a request to finalize the login session.
// After successful login, the router requires a follow-up request before
// other API endpoints become accessible. This mimics the browser's behavior
// of reloading the page after login.
func (h *HomeStationClient) activateSession() {
	h.doGet(h.baseURL + "/api/v1/session/menu")
}

type saltResp struct {
	Error     string `json:"error"`
	Salt      string `json:"salt"`
	SaltWebUI string `json:"saltwebui"`
}

// tryLogin performs the two-step login using the salt and hashed password
func (h *HomeStationClient) tryLogin() error {
	loginURL := h.baseURL + "/api/v1/session/login"

	form := url.Values{}
	form.Set("username", h.user)
	form.Set("password", "seeksalthash")
	// Use custom POST so we can set the same headers the browser sends
	_, body, err := h.doPostForm(loginURL, form)
	if err != nil {
		return fmt.Errorf("failed requesting salt: %w", err)
	}

	var saltResponse saltResp
	if err := json.Unmarshal(body, &saltResponse); err != nil {
		return fmt.Errorf("failed parsing salt response: %w", err)
	}

	// We need both salt and saltwebui for the double-PBKDF2 algorithm
	salt := saltResponse.Salt
	saltWebUI := saltResponse.SaltWebUI

	if salt == "" {
		return errors.New("no salt returned from router")
	}
	if saltWebUI == "" {
		return errors.New("no saltwebui returned from router")
	}

	// Compute the double-PBKDF2 hash as per the router's login.js:
	// 1. hash1 = PBKDF2(password, salt, 1000, 128bits) -> hex
	// 2. hash2 = PBKDF2(hash1, saltwebui, 1000, 128bits) -> hex
	hash1 := pbkdf2Hex(h.pass, salt)
	finalHash := pbkdf2Hex(hash1, saltWebUI)

	// Send the login request with the computed hash
	form2 := url.Values{}
	form2.Set("username", h.user)
	form2.Set("password", finalHash)
	_, resp2body, err := h.doPostForm(loginURL, form2)
	if err != nil {
		return fmt.Errorf("failed posting hashed password: %w", err)
	}

	// check JSON response for error=="ok"
	var jr map[string]any
	if err := json.Unmarshal(resp2body, &jr); err == nil {
		if e, ok := jr["error"].(string); ok && e == "ok" {
			// Activate the session (required before other API calls work)
			h.activateSession()

			return nil
		}

		// Router sometimes returns a message code such as MSG_LOGIN_150 when
		// login is blocked (e.g. too many attempts). Check for that and return
		// a clearer error to the caller.
		if msg, ok := jr["message"].(string); ok && strings.Contains(msg, "MSG_LOGIN_150") {
			return fmt.Errorf("An active session exists. Logout first. %s", msg)
		}
	}

	return errors.New("login failed with provided credentials")
}

// doPostForm sends a POST with form-encoded body and returns the response and body bytes
func (h *HomeStationClient) doPostForm(u string, form url.Values) (*http.Response, []byte, error) {
	bodyStr := form.Encode()
	req, err := http.NewRequest("POST", u, strings.NewReader(bodyStr))
	if err != nil {
		return nil, nil, err
	}
	// apply common headers shared between GET/POST requests
	h.setCommonHeaders(req)
	// request-specific header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, b, nil
}

// doGet sends a GET with common headers and returns the response and body bytes
func (h *HomeStationClient) doGet(u string) (*http.Response, []byte, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}
	h.setCommonHeaders(req)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, b, nil
}

func (h *HomeStationClient) setCommonHeaders(req *http.Request) {
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", h.baseURL+"/")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
}

type hostTblResp struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Data    struct {
		HostTbl []struct {
			Physaddress string `json:"physaddress"`
			Ipaddress   string `json:"ipaddress"`
			Hostname    string `json:"hostname"`
			Active      string `json:"active"` // "true" / "false"
		} `json:"hostTbl"`
	} `json:"data"`
	Token string `json:"token"`
}

func (h *HomeStationClient) fetchHostTbl() ([]Device, error) {
	_, body, err := h.doGet(h.baseURL + "/api/v1/host/hostTbl")
	if err != nil {
		return nil, fmt.Errorf("failed fetching host table: %w", err)
	}

	var r hostTblResp
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed parsing host table JSON: %w", err)
	}
	if r.Error != "ok" {
		return nil, fmt.Errorf("host table returned error: %s", r.Error)
	}
	var out []Device
	for _, e := range r.Data.HostTbl {
		active := false
		if e.Active == "true" {
			active = true
		}
		out = append(out, Device{MAC: e.Physaddress, IP: e.Ipaddress, Hostname: e.Hostname, Active: active})
	}
	return out, nil
}

// ListConnected logs in, returns connected devices, and logs out
func (h *HomeStationClient) ListConnected() ([]Device, error) {
	if err := h.tryLogin(); err != nil {
		return nil, err
	}

	devices, err := h.fetchHostTbl()

	h.logout()

	return devices, err
}

// logout ends the current session on the router
func (h *HomeStationClient) logout() {
	logoutURL := h.baseURL + "/api/v1/session/logout"
	h.doPostForm(logoutURL, url.Values{})
}
