package access

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	jwt "github.com/dgrijalva/jwt-go"

	"bitbucket.org/teachingstrategies/authorization-service/internal/pkg/authrlib"
	"bitbucket.org/teachingstrategies/go-svc-bootstrap/authorization"
)

// NewAccessValidationMiddlewares creates a middleware function to check jwt client id and userID from path
func NewAccessValidationMiddlewares(ctx *authrlib.AppContext) []func(next http.Handler) http.Handler {
	return (&accessValidationMiddleware{ctx.ConfigService.Config().App.KeysServer}).middlewares()
}

type accessValidationMiddleware struct {
	keysServerURL string
}

func (m *accessValidationMiddleware) middlewares() []func(next http.Handler) http.Handler {
	return []func(next http.Handler) http.Handler{
		authorization.FindTokenMiddleware(),
		authorization.VerifyTokenMiddleware(m.keysServerURL, authorization.ClaimsValidator(m.validateClaims)),
	}
}

func (*accessValidationMiddleware) validateClaims(claims jwt.MapClaims, r *http.Request) (*http.Request, error) {
	var subscription string
	var ok bool
	if sub, ok := claims["sub"]; !ok || sub == nil {
		return r, authorization.ErrSubNotFound
	}
	if subscription, ok = claims["sub"].(string); !ok {
		return r, authorization.ErrSubNotFound
	}

	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		return r, fmt.Errorf("internal server error: %v", err)
	}

	if strconv.Itoa(userID) != subscription {
		return r, authorization.ErrSubInvalid
	}

	return r, nil
}
