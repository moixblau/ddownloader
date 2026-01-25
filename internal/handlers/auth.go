package handlers

import (
	"ddownload/internal/config"
	"net/http"
)

const (
	AuthCookieName = "ddownloader_auth"
)

func AuthMiddleware(cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !cfg.IsAuthEnabled() {
			next(w, r)
			return
		}

		cookie, err := r.Cookie(AuthCookieName)
		if err == nil && cookie.Value == cfg.AuthPassword {
			next(w, r)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.IsAuthEnabled() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		password := r.FormValue("password")

		if password == h.cfg.AuthPassword {
			http.SetCookie(w, &http.Cookie{
				Name:     AuthCookieName,
				Value:    h.cfg.AuthPassword,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   86400 * 7,
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if err := h.templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Theme": h.cfg.Theme,
			"Error": "Contraseña incorrecta",
		}); err != nil {
			h.cfg.Logger.Error("Error executing login template", "error", err)
		}
		return
	}

	if err := h.templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
		"Theme": h.cfg.Theme,
		"Error": "",
	}); err != nil {
		h.cfg.Logger.Error("Error executing login template", "error", err)
	}
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
