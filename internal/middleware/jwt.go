package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const UserIDKey = "user_id"

// JWT validates "Authorization: Bearer <token>" and stores the user ID
// (from the "sub" claim) in the Echo context under UserIDKey.
func JWT(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenStr, ok := strings.CutPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
			if !ok || tokenStr == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing bearer token")
			}

			claims := &jwt.RegisteredClaims{}
			_, err := jwt.ParseWithClaims(tokenStr, claims,
				func(*jwt.Token) (any, error) { return secret, nil },
				// Pin the algorithm: prevents "alg: none" / alg-confusion attacks.
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
				jwt.WithExpirationRequired(),
			)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			userID, err := strconv.ParseUint(claims.Subject, 10, 64)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token subject")
			}

			c.Set(UserIDKey, uint(userID))
			return next(c)
		}
	}
}
