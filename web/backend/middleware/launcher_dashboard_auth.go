package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// LauncherDashboardCookieName is the HttpOnly cookie set after a successful token login.
const LauncherDashboardCookieName = "picoclaw_launcher_auth"

// launcherDashboardSessionMaxAgeSec is the session cookie lifetime (7 days).
const launcherDashboardSessionMaxAgeSec = 7 * 24 * 3600

const launcherSessionMACLabel = "picoclaw-launcher-v1"

// SessionCookieValue is the expected cookie value for the given signing key and dashboard token.
func SessionCookieValue(signingKey []byte, dashboardToken string) string {
	mac := hmac.New(sha256.New, signingKey)
	_, _ = mac.Write([]byte(launcherSessionMACLabel))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(dashboardToken))
	return hex.EncodeToString(mac.Sum(nil))
}

// LauncherDashboardAuthConfig holds runtime material for dashboard access checks.
type LauncherDashboardAuthConfig struct {
	ExpectedCookie string
	Token          string
	// SecureCookie sets the session cookie's Secure flag. If nil, DefaultLauncherDashboardSecureCookie is used.
	SecureCookie func(*http.Request) bool
}

// DefaultLauncherDashboardSecureCookie mirrors typical production HTTPS detection (TLS or X-Forwarded-Proto).
func DefaultLauncherDashboardSecureCookie(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// SetLauncherDashboardSessionCookie writes the HttpOnly session cookie after successful dashboard token login.
func SetLauncherDashboardSessionCookie(
	w http.ResponseWriter,
	r *http.Request,
	sessionValue string,
	secure func(*http.Request) bool,
) {
	if secure == nil {
		secure = DefaultLauncherDashboardSecureCookie
	}
	http.SetCookie(w, &http.Cookie{
		Name:     LauncherDashboardCookieName,
		Value:    sessionValue,
		Path:     "/",
		MaxAge:   launcherDashboardSessionMaxAgeSec,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure(r),
	})
}

// ClearLauncherDashboardSessionCookie clears the dashboard session (e.g. logout).
func ClearLauncherDashboardSessionCookie(w http.ResponseWriter, r *http.Request, secure func(*http.Request) bool) {
	if secure == nil {
		secure = DefaultLauncherDashboardSecureCookie
	}
	http.SetCookie(w, &http.Cookie{
		Name:     LauncherDashboardCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure(r),
		Expires:  time.Unix(0, 0),
	})
}

// LauncherDashboardAuth requires a valid session cookie or Authorization: Bearer <token>
// before calling next. Public paths are login page and /api/auth/* handlers.
func LauncherDashboardAuth(cfg LauncherDashboardAuthConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := canonicalAuthPath(r.URL.Path)
		if handled := tryLauncherQueryTokenLogin(w, r, p, cfg); handled {
			return
		}
		if isPublicLauncherDashboardPath(r.Method, p) {
			next.ServeHTTP(w, r)
			return
		}
		if validLauncherDashboardAuth(r, cfg) {
			next.ServeHTTP(w, r)
			return
		}
		rejectLauncherDashboardAuth(w, r, p)
	})
}

// canonicalAuthPath matches path cleaning used for routing decisions so
// prefixes like /assets/../ cannot bypass auth (CVE-class traversal).

// tryLauncherQueryTokenLogin validates ?token= on GET only (non-/api), sets the session
// cookie when correct, and redirects with 303 so the follow-up is a plain GET without side effects.
// Invalid token is rejected like any other unauthenticated browser request.
func tryLauncherQueryTokenLogin(
	w http.ResponseWriter,
	r *http.Request,
	canonicalPath string,
	cfg LauncherDashboardAuthConfig,
) bool {
	if r.Method != http.MethodGet {
		return false
	}
	if canonicalPath == "/api" || strings.HasPrefix(canonicalPath, "/api/") {
		return false
	}
	qToken := strings.TrimSpace(r.URL.Query().Get("token"))
	if qToken == "" {
		return false
	}
	if len(qToken) != len(cfg.Token) || subtle.ConstantTimeCompare([]byte(qToken), []byte(cfg.Token)) != 1 {
		rejectLauncherDashboardAuth(w, r, canonicalPath)
		return true
	}
	SetLauncherDashboardSessionCookie(w, r, cfg.ExpectedCookie, cfg.SecureCookie)
	http.Redirect(w, r, redirectAfterQueryTokenLogin(r, canonicalPath), http.StatusSeeOther)
	return true
}

func redirectAfterQueryTokenLogin(r *http.Request, canonicalPath string) string {
	if canonicalPath == "/launcher-login" {
		return "/"
	}
	q := r.URL.Query()
	q.Del("token")
	enc := q.Encode()
	if enc != "" {
		return canonicalPath + "?" + enc
	}
	return canonicalPath
}

func canonicalAuthPath(raw string) string {
	if raw == "" {
		return "/"
	}
	c := path.Clean(raw)
	switch c {
	case ".", "":
		return "/"
	default:
		if c[0] != '/' {
			return "/" + c
		}
		return c
	}
}

func isPublicLauncherDashboardPath(method, p string) bool {
	if isPublicLauncherDashboardStatic(method, p) {
		return true
	}
	if method == http.MethodGet && (p == "/pico/ws" || p == "/pico/ws/" || p == "/pet/ws" || p == "/pet/ws/") {
		return true
	}
	switch p {
	case "/api/auth/login":
		return method == http.MethodPost
	case "/api/auth/logout":
		return method == http.MethodPost
	case "/api/auth/status":
		return method == http.MethodGet
	}
	return false
}

// isPublicLauncherDashboardStatic allows the SPA login route and embedded
// frontend assets without a session (GET/HEAD only).
func isPublicLauncherDashboardStatic(method, p string) bool {
	if method != http.MethodGet && method != http.MethodHead {
		return false
	}
	if p == "/launcher-login" {
		return true
	}
	if strings.HasPrefix(p, "/assets/") {
		return true
	}
	switch p {
	case "/favicon.ico", "/favicon.svg", "/favicon-96x96.png",
		"/apple-touch-icon.png", "/site.webmanifest", "/robots.txt":
		return true
	default:
		return false
	}
}

func validLauncherDashboardAuth(r *http.Request, cfg LauncherDashboardAuthConfig) bool {
	if isCrossOriginAPIRequest(r) {
		return validLauncherDashboardBearer(r, cfg.Token)
	}

	return validLauncherDashboardSessionCookie(r, cfg.ExpectedCookie) || validLauncherDashboardBearer(r, cfg.Token)
}

func validLauncherDashboardSessionCookie(r *http.Request, expectedCookie string) bool {
	if c, err := r.Cookie(LauncherDashboardCookieName); err == nil {
		return subtle.ConstantTimeCompare([]byte(c.Value), []byte(expectedCookie)) == 1
	}

	return false
}

func validLauncherDashboardBearer(r *http.Request, token string) bool {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if strings.HasPrefix(auth, prefix) {
		bearer := strings.TrimSpace(auth[len(prefix):])
		if len(bearer) == len(token) && subtle.ConstantTimeCompare([]byte(bearer), []byte(token)) == 1 {
			return true
		}
	}

	return false
}

func isCrossOriginAPIRequest(r *http.Request) bool {
	p := canonicalAuthPath(r.URL.Path)
	if p != "/api" && !strings.HasPrefix(p, "/api/") {
		return false
	}

	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Scheme == "" || originURL.Host == "" {
		return true
	}

	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}

	requestHost := strings.TrimSpace(r.Host)
	if requestHost == "" {
		requestHost = r.URL.Host
	}

	return !strings.EqualFold(originURL.Scheme, scheme) || !strings.EqualFold(originURL.Host, requestHost)
}

func rejectLauncherDashboardAuth(w http.ResponseWriter, r *http.Request, canonicalPath string) {
	if strings.HasPrefix(canonicalPath, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}
	http.Redirect(w, r, "/launcher-login", http.StatusFound)
}
