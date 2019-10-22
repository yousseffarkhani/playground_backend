package middleware

import (
	"context"
	"fmt"
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
	middlewares["isLogged"] = Use(IsLogged)
	middlewares["refresh"] = Use(IsLogged, RefreshJWT)
	return middlewares
}

func Use(m ...MW) MW {
	return func(finalPage http.Handler) http.Handler {
		for i := len(m) - 1; i >= 0; i-- {
			finalPage = m[i](finalPage)
		}
		return finalPage
	}
}

func IsLogged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		c, err := r.Cookie("Token")
		if err != nil {
			fmt.Println("From middleware.go", err)
		} else {
			claims, token, err := authentication.ParseCookie(c)
			if err != nil || !token.Valid {
				fmt.Println(err)
			} else {
				ctx = context.WithValue(r.Context(), "claims", claims)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RefreshJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := ctx.Value("claims").(*authentication.Claims)
		if ok {
			if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) < 15*time.Minute {
				authentication.SetJwtCookie(w, claims.Username)
				fmt.Println("Refreshed Token")
			} else {
				fmt.Println("Token doesn't need to be refreshed")
			}
		} else {
			fmt.Println("From RefreshJWT : User not connected")
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
