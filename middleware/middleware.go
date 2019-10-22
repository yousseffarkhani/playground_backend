package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/yousseffarkhani/playground/backend2/server"

	"github.com/yousseffarkhani/playground/backend2/authentication"
)

type MW func(http.Handler) http.Handler

func (m MW) ThenFunc(finalPage func(http.ResponseWriter, *http.Request)) http.Handler {
	return m(http.HandlerFunc(finalPage))
}

func Initialize() map[string]server.Middleware {
	middlewares := make(map[string]server.Middleware)
	middlewares["isLogged"] = use(IsLogged)
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

func IsLogged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
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
