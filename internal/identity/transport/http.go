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
		r.Get("/setup-status", h.setupStatus)
		r.Post("/setup", h.setup)
		r.Post("/login", h.login)
		r.Post("/totp/login", h.loginTOTP)
		r.Post("/logout", h.logout)
		r.Post("/passkeys/login/options", h.beginPasskeyLogin)
		r.Post("/passkeys/login/verify", h.finishPasskeyLogin)
		r.Group(func(r chi.Router) {
			r.Use(h.RequireSession)
			r.Get("/me", h.me)
			r.Get("/totp", h.totpStatus)
			r.Post("/totp/setup", h.setupTOTP)
			r.Post("/totp/enable", h.enableTOTP)
			r.Post("/totp/disable", h.disableTOTP)
			r.Post("/passkeys/register/options", h.beginPasskeyRegistration)
			r.Post("/passkeys/register/verify", h.finishPasskeyRegistration)
		})
	})
}

func (h *HTTP) setupStatus(w http.ResponseWriter, r *http.Request) {
	required, err := h.sessions.SetupRequired(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "setup_status_failed", "could not read setup status")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"required": required})
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

func (h *HTTP) loginTOTP(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	p, token, err := h.sessions.LoginWithTOTP(r.Context(), input.Email, input.Code, r.RemoteAddr, r.UserAgent())
	if err != nil {
		if errors.Is(err, application.ErrTOTPRateLimited) {
			httpx.Error(w, http.StatusTooManyRequests, "totp_rate_limited", "尝试次数过多，请 5 分钟后再试")
			return
		}
		httpx.Error(w, http.StatusUnauthorized, "invalid_totp", "邮箱或动态码不正确")
		return
	}
	h.setCookie(w, token)
	httpx.JSON(w, http.StatusOK, p)
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

func (h *HTTP) totpStatus(w http.ResponseWriter, r *http.Request) {
	p, _ := Principal(r.Context())
	status, err := h.sessions.TOTPStatus(r.Context(), p.UserID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "totp_status_failed", "无法读取动态码状态")
		return
	}
	httpx.JSON(w, http.StatusOK, status)
}

func (h *HTTP) setupTOTP(w http.ResponseWriter, r *http.Request) {
	p, _ := Principal(r.Context())
	enrollment, err := h.sessions.BeginTOTPEnrollment(p.Email)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "totp_setup_failed", "无法创建动态码配置")
		return
	}
	httpx.JSON(w, http.StatusOK, enrollment)
}

func (h *HTTP) enableTOTP(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	p, _ := Principal(r.Context())
	codes, err := h.sessions.EnableTOTP(r.Context(), p.UserID, input.Secret, input.Code)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "totp_verification_failed", "动态码验证失败")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"enabled": true, "recovery_codes": codes})
}

func (h *HTTP) disableTOTP(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code string `json:"code"`
	}
	if err := httpx.Decode(w, r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	p, _ := Principal(r.Context())
	if err := h.sessions.DisableTOTP(r.Context(), p.UserID, input.Code); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "totp_verification_failed", "动态码或恢复码不正确")
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
