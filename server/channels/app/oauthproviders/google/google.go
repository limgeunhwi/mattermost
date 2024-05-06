package oauthgoogle

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/mattermost/mattermost/server/v8/einterfaces"
)

type GoogleProvider struct {
}

type GoogleUser struct {
	Sub       string `json:"sub"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
	Name      string `json:"name"`
	Firstname string `json:"given_name"`
	Lastname  string `json:"family_name"`
}

func init() {
	provider := &GoogleProvider{}
	einterfaces.RegisterOAuthProvider(model.ServiceGoogle, provider)
}

func userFromGoogleUser(logger mlog.LoggerIFace, glu *GoogleUser) *model.User {
	user := &model.User{}
	username := glu.Name
	if username == "" {
		username = glu.Email
	}
	user.Username = model.CleanUsername(logger, username)
	user.FirstName = glu.Firstname
	user.LastName = glu.Lastname
	splitName := strings.Split(glu.Name, " ")
	if len(splitName) == 2 && user.FirstName == "" {
		user.FirstName = splitName[0]
		user.LastName = splitName[1]
	} else if len(splitName) >= 2 {
		user.FirstName = splitName[0]
		user.LastName = strings.Join(splitName[1:], " ")
	} else {
		user.FirstName = glu.Name
	}

	user.Email = glu.Email
	user.Email = strings.ToLower(user.Email)
	user.AuthData = &glu.Sub
	user.AuthService = model.ServiceGoogle

	return user
}

func getGoogleUserFromJSON(data io.Reader) (*GoogleUser, error) {
	decoder := json.NewDecoder(data)
	var glu GoogleUser
	err := decoder.Decode(&glu)
	if err != nil {
		return nil, err
	}
	return &glu, nil
}

func (glu *GoogleUser) IsValid() error {
	if glu.Sub == "" {
		return errors.New("user id can't be 0")
	}

	if glu.Email == "" {
		return errors.New("user e-mail should not be empty")
	}

	return nil
}

func (gp *GoogleProvider) GetUserFromJSON(c request.CTX, data io.Reader, tokenUser *model.User) (*model.User, error) {
	glu, err := getGoogleUserFromJSON(data)
	if err != nil {
		return nil, err
	}
	if err = glu.IsValid(); err != nil {
		result, _ := io.ReadAll(data)
		return nil, errors.New("Json data : " + string(result))
		// return nil, err
	}

	return userFromGoogleUser(c.Logger(), glu), nil
}

func (gp *GoogleProvider) GetSSOSettings(_ request.CTX, config *model.Config, service string) (*model.SSOSettings, error) {
	return &config.GoogleSettings, nil
}

func (gp *GoogleProvider) GetUserFromIdToken(_ request.CTX, idToken string) (*model.User, error) {
	return nil, nil
}

func (gp *GoogleProvider) IsSameUser(_ request.CTX, dbUser, oauthUser *model.User) bool {
	return dbUser.AuthData == oauthUser.AuthData
}
