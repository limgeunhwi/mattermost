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

// Name: A person's name. If the name is a mononym, the family name is empty.
type Name struct {
	// DisplayName: Output only. The display name formatted according to the locale
	// specified by the viewer's account or the `Accept-Language` HTTP header.
	DisplayName string `json:"displayName,omitempty"`
	// DisplayNameLastFirst: Output only. The display name with the last name first
	// formatted according to the locale specified by the viewer's account or the
	// `Accept-Language` HTTP header.
	DisplayNameLastFirst string `json:"displayNameLastFirst,omitempty"`
	// FamilyName: The family name.
	FamilyName string `json:"familyName,omitempty"`
	// GivenName: The given name.
	GivenName string `json:"givenName,omitempty"`
}

type Photo struct {
	// Default: True if the photo is a default photo; false if the photo is a
	// user-provided photo.
	Default bool `json:"default,omitempty"`
	// Url: The URL of the photo. You can change the desired size by appending a
	// query parameter `sz={size}` at the end of the url, where {size} is the size
	// in pixels. Example:
	// https://lh3.googleusercontent.com/-T_wVWLlmg7w/AAAAAAAAAAI/AAAAAAAABa8/00gzXvDBYqw/s100/photo.jpg?sz=50
	Url string `json:"url,omitempty"`
}

// EmailAddress: A person's email address.
type EmailAddress struct {
	// DisplayName: The display name of the email.
	DisplayName string `json:"displayName,omitempty"`
	// Value: The email address.
	Value string `json:"value,omitempty"`
}

// This is subset of Person struct defined in google-api-go-client/people/v1/people-gen.go
type GoogleUser struct {
	// EmailAddresses: The person's email addresses. For `people.connections.list`
	// and `otherContacts.list` the number of email addresses is limited to 100. If
	// a Person has more email addresses the entire set can be obtained by calling
	// GetPeople.
	EmailAddresses []*EmailAddress `json:"emailAddresses,omitempty"`
	// Etag: The HTTP entity tag (https://en.wikipedia.org/wiki/HTTP_ETag) of the
	// resource. Used for web cache validation.
	Etag string `json:"etag,omitempty"`
	// Names: The person's names. This field is a singleton for contact sources.
	Names []*Name `json:"names,omitempty"`
	// Photos: Output only. The person's photos.
	Photos []*Photo `json:"photos,omitempty"`
	// ResourceName: The resource name for the person, assigned by the server. An
	// ASCII string in the form of `people/{person_id}`.
	ResourceName string `json:"resourceName,omitempty"`
}

func init() {
	provider := &GoogleProvider{}
	einterfaces.RegisterOAuthProvider(model.ServiceGoogle, provider)
}

func userFromGoogleUser(logger mlog.LoggerIFace, glu *GoogleUser) *model.User {
	user := &model.User{}
	user.Id = glu.ResourceName
	glu_name := glu.Names[0]
	glu_mail := glu.EmailAddresses[0]
	username := glu_name.DisplayName
	if username == "" {
		username = glu_mail.Value
	}
	user.Username = model.CleanUsername(logger, username)
	user.FirstName = glu_name.GivenName
	user.LastName = glu_name.FamilyName

	user.Email = glu_mail.Value
	user.Email = strings.ToLower(user.Email)
	user.AuthData = &glu.ResourceName
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
	// result, _ := io.ReadAll(data)
	// return nil, errors.New("Json data : " + string(result))
	glu, err := getGoogleUserFromJSON(data)
	if err != nil {
		return nil, err
	}
	if err = glu.IsValid(); err != nil {
		return nil, err
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
