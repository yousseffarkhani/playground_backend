package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/yousseffarkhani/playground/backend2/server"

	"github.com/yousseffarkhani/playground/backend2/authentication"
)

type MW func(http.Handler) http.Handler

func (m MW) ThenFunc(finalPage func(http.ResponseWriter, *http.Request)) http.Handler {
	return m(http.HandlerFunc(finalPage))
}

func Initialize() map[string]server.Middleware {
	middlewares := make(map[string]server.Middleware)
	middlewares["isLogged"] = use(isLogged)
	middlewares["refresh"] = use(isLogged, refreshJWT)
	middlewares["authorized"] = use(isLogged, refreshJWT, isAuthorized)
	return middlewares
}

func use(m ...MW) MW {
	return func(finalPage http.Handler) http.Handler {
		for i := len(m) - 1; i >= 0; i-- {
			finalPage = m[i](finalPage)
		}
		return finalPage
	}
}

func isLogged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		c, err := r.Cookie("Token")
		if err != nil {
			log.Println("From middleware.go", err)
		} else {
			claims, token, err := authentication.ParseCookie(c)
			if err != nil || !token.Valid {
				log.Println(err)
			} else {
				ctx = context.WithValue(r.Context(), "claims", claims)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func refreshJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := ctx.Value("claims").(*authentication.Claims)
		if ok {
			if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) < 15*time.Minute {
				authentication.SetJwtCookie(w, claims.Username)
				log.Println("Refreshed Token")
			} else {
				log.Println("Token doesn't need to be refreshed")
			}
		} else {
			log.Println("From RefreshJWT : User not connected")
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isAuthorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, ok := ctx.Value("claims").(*authentication.Claims)
		if !ok {
			log.Println("Access denied")
			http.Redirect(w, r, server.URLLogin, http.StatusFound)
			return
		}
		log.Println("authorized")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
