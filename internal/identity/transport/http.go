package transport

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/identity/application"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"net/http"
	"strings"
	"time"
)

type HTTP struct {
	sessions   *application.Service
	passkeys   *application.Passkeys
	cookieName string
	ttl        time.Duration
	secure     bool
}

func NewHTTP(sessions *application.Service, passkeys *application.Passkeys, cookieName string, ttl time.Duration, secure bool) *HTTP {
	return &HTTP{sessions: sessions, passkeys: passkeys, cookieName: cookieName, ttl: ttl, secure: secure}
}
func (h *HTTP) Register(r chi.Router) {
	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/setup", h.setup)
		r.Post("/login", h.login)
		r.Post("/logout", h.logout)
		r.Post("/passkeys/login/options", h.beginPasskeyLogin)
		r.Post("/passkeys/login/verify", h.finishPasskeyLogin)
		r.Group(func(r chi.Router) {
			r.Use(h.RequireSession)
			r.Get("/me", h.me)
			r.Post("/passkeys/register/options", h.beginPasskeyRegistration)
			r.Post("/passkeys/register/verify", h.finishPasskeyRegistration)
		})
	})
}
func (h *HTTP) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(h.cookieName)
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
			return
		}
		p, err := h.sessions.Authenticate(r.Context(), cookie.Value)
		if err != nil {
			h.clearCookie(w)
			httpx.Error(w, http.StatusUnauthorized, "unauthorized", "session is invalid or expired")
			return
		}
		next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
	})
}

type credentials struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

func (h *HTTP) setup(w http.ResponseWriter, r *http.Request) {
	var input credentials
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	p, token, err := h.sessions.Setup(r.Context(), input.Email, input.DisplayName, input.Password)
	if err != nil {
		status := 400
		code := "setup_failed"
		if errors.Is(err, application.ErrSetupComplete) {
			status = 409
			code = "setup_complete"
		}
		httpx.Error(w, status, code, err.Error())
		return
	}
	h.setCookie(w, token)
	httpx.JSON(w, 201, p)
}
func (h *HTTP) login(w http.ResponseWriter, r *http.Request) {
	var input credentials
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	p, token, err := h.sessions.Login(r.Context(), input.Email, input.Password, r.UserAgent())
	if err != nil {
		httpx.Error(w, 401, "invalid_credentials", "email or password is incorrect")
		return
	}
	h.setCookie(w, token)
	httpx.JSON(w, 200, p)
}
func (h *HTTP) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(h.cookieName); err == nil {
		_ = h.sessions.Logout(r.Context(), cookie.Value)
	}
	h.clearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
func (h *HTTP) me(w http.ResponseWriter, r *http.Request) {
	p, _ := Principal(r.Context())
	httpx.JSON(w, 200, p)
}

type emailInput struct {
	Email string `json:"email"`
}

func (h *HTTP) beginPasskeyLogin(w http.ResponseWriter, r *http.Request) {
	var input emailInput
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, 400, "invalid_request", err.Error())
		return
	}
	options, id, err := h.passkeys.BeginLogin(r.Context(), strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		httpx.Error(w, 401, "passkey_unavailable", "passkey login is unavailable")
		return
	}
	httpx.JSON(w, 200, map[string]any{"ceremony_id": id, "options": options})
}
func (h *HTTP) finishPasskeyLogin(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("ceremony_id"))
	if err != nil {
		httpx.Error(w, 400, "invalid_ceremony", "invalid ceremony id")
		return
	}
	p, token, err := h.passkeys.FinishLogin(r.Context(), id, r, h.ttl)
	if err != nil {
		httpx.Error(w, 401, "passkey_verification_failed", "passkey verification failed")
		return
	}
	h.setCookie(w, token)
	httpx.JSON(w, 200, p)
}
func (h *HTTP) beginPasskeyRegistration(w http.ResponseWriter, r *http.Request) {
	p, _ := Principal(r.Context())
	options, id, err := h.passkeys.BeginRegistration(r.Context(), p.UserID)
	if err != nil {
		httpx.Error(w, 500, "passkey_begin_failed", "could not begin passkey registration")
		return
	}
	httpx.JSON(w, 200, map[string]any{"ceremony_id": id, "options": options})
}
func (h *HTTP) finishPasskeyRegistration(w http.ResponseWriter, r *http.Request) {
	p, _ := Principal(r.Context())
	id, err := uuid.Parse(r.URL.Query().Get("ceremony_id"))
	if err != nil {
		httpx.Error(w, 400, "invalid_ceremony", "invalid ceremony id")
		return
	}
	if err = h.passkeys.FinishRegistration(r.Context(), p.UserID, id, r); err != nil {
		httpx.Error(w, 400, "passkey_verification_failed", "passkey registration could not be verified")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *HTTP) setCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: h.cookieName, Value: token, Path: "/", HttpOnly: true, Secure: h.secure, SameSite: http.SameSiteLaxMode, MaxAge: int(h.ttl.Seconds())})
}
func (h *HTTP) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: h.cookieName, Path: "/", HttpOnly: true, Secure: h.secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}
