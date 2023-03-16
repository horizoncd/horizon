package util

import (
	"net/http"

	"github.com/gorilla/sessions"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/core/common"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
)

func SetSession(ss *sessions.Session,
	request *http.Request, response http.ResponseWriter,
	user *usermodels.User) error {
	ss.Values[common.SessionKeyAuthUser] = &userauth.DefaultInfo{
		Name:     user.Name,
		FullName: user.FullName,
		ID:       user.ID,
		Email:    user.Email,
		Admin:    user.Admin,
	}

	if err := ss.Save(request, response); err != nil {
		return perror.Wrapf(herrors.ErrSessionSaveFailed,
			"err = %v", err)
	}
	return nil
}

func GetSession(store sessions.Store, request *http.Request) (*sessions.Session, error) {
	session, err := store.Get(request, common.CookieKeyAuth)
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrSessionNotFound,
			"session name = %s\n err = %v", common.CookieKeyAuth, err)
	}
	return session, nil
}
