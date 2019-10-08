package access

import (
	"fmt"
	"net/http"
	"strconv"

	"bitbucket.org/teachingstrategies/authorization-service/internal/pkg/authrlib"
	"github.com/gamegos/jsend"
	"github.com/go-chi/chi"
)

// NewAccessHandler creates new instance of access handler
func NewAccessHandler(ctx *authrlib.AppContext) http.HandlerFunc {
	service := &accessService{&accessConverter{}, &accessRepo{ctx.DbManager.Db()}}
	return (&accessHandler{ctx, service}).handlerFunc()
}

// struct which produces http.HandlerFunc
type accessHandler struct {
	ctx     *authrlib.AppContext
	service Service
}

func (ah *accessHandler) handlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
		logger := ah.ctx.Logger.HandlerLogger(r)

		if err != nil {
			msg := "userID has incorrect value"
			logger.Error().Str("user_id", chi.URLParam(r, "userID")).Msg(msg)
			if _, err = jsend.Wrap(w).Message(msg).Status(http.StatusInternalServerError).Send(); err != nil {
				logger.Warn().Err(err).Msgf("unable to reply: %s", msg)
			}
			return
		}
		resp, err := ah.service.Access(userID)
		if err != nil {
			replyAccessError(userID, err, logger, w)
			return
		}
		if _, err = jsend.Wrap(w).Message("request completed").Data(resp).Status(http.StatusOK).Send(); err != nil {
			logger.Warn().Err(err).Msg("unable to reply success")
		}
	}
}

func replyAccessError(userID int, err error, logger *authrlib.AppLogger, w http.ResponseWriter) {
	if err == errNotFound {
		logger.Warn().Err(err).Int("user_id", userID).Msg("not found")
		if _, err = jsend.Wrap(w).Message(err.Error()).Status(http.StatusNotFound).Send(); err != nil {
			logger.Warn().Err(err).Msg("unable to reply: user not found")
		}
		return
	}
	if err == errNotAllowed {
		logger.Warn().Err(err).Int("user_id", userID).Msg("forbidden")
		if _, err = jsend.Wrap(w).Message(err.Error()).Status(http.StatusForbidden).Send(); err != nil {
			logger.Warn().Err(err).Msg("unable to reply: forbidden")
		}
		return
	}
	err = fmt.Errorf("unable to fetch access data for userID %d; err=%v", userID, err)
	logger.Warn().Err(err).Int("user_id", userID).Msg("unable to fetch access data")
	if _, replyErr := jsend.Wrap(w).Message(err.Error()).Status(http.StatusInternalServerError).Send(); replyErr != nil {
		logger.Warn().Err(err).Msg("unable to reply: unable to fetch access data")
	}
}
