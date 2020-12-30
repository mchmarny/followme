package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kurrik/oauth1a"
	"github.com/mchmarny/followme/internal/data"
	"github.com/mchmarny/followme/pkg/format"
	"github.com/mchmarny/followme/pkg/id"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

const (
	userIDCookieName = "user_id"
	authIDCookieName = "auth_id"
)

var (
	userCookieDuration = 60 * 60 * 24 * 30 // month in sec
	maxSessionAge      = 5.0               // min
	sessionCookieAge   = 5 * 60            // maxSessionAge in secs
	maxEventLimit      = 100
)

// AuthSession represents the authenticated user session
type AuthSession struct {
	ID     string    `storm:"id" json:"id"`
	On     time.Time `json:"on"`
	Config string    `json:"config"`
}

func (a *App) authLoginHandler(c *gin.Context) {
	uid, _ := c.Cookie(userIDCookieName)
	if uid != "" {
		c.Redirect(http.StatusSeeOther, "/view/dash")
		return
	}
	httpClient := new(http.Client)
	userConfig := &oauth1a.UserConfig{}
	if err := userConfig.GetRequestToken(a.authService, httpClient); err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting request token")
		return
	}

	AuthURL, err := userConfig.GetAuthorizeURL(a.authService)
	if err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting authorization URL")
		return
	}

	authSession := &AuthSession{
		ID:     id.NewID(),
		Config: userConfigToString(userConfig),
		On:     time.Now().UTC(),
	}

	if err := a.db.Save(authSession); err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error saving authentication session")
		return
	}

	c.SetCookie(authIDCookieName, authSession.ID, sessionCookieAge, "/", c.Request.Host, false, true)
	c.Redirect(http.StatusFound, AuthURL)
}

func (a *App) authCallbackHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sessionID, err := c.Cookie(authIDCookieName)
	if err != nil {
		a.viewErrorHandler(c, http.StatusUnauthorized, err, "Error handling callback with no session id")
		return
	}

	var authSession AuthSession
	if err := a.db.One("ID", sessionID, &authSession); err != nil || authSession.ID == "" {
		a.viewErrorHandler(c, http.StatusUnauthorized, err, fmt.Sprintf("Unable to find auth config for this sessions ID: %s", sessionID))
		return
	}

	sessionAge := time.Now().UTC().Sub(authSession.On)
	if sessionAge.Minutes() > maxSessionAge {
		a.viewErrorHandler(c, http.StatusUnauthorized, err, fmt.Sprintf("session %s expired. Age %v, expected %f min", sessionAge, maxSessionAge, maxSessionAge))
		return
	}

	userConfig, err := userConfigFromString(authSession.Config)
	if err != nil {
		a.viewErrorHandler(c, http.StatusUnauthorized, err, "Error decoding user config in sessions storage")
		return
	}

	token, verifier, err := userConfig.ParseAuthorize(c.Request, a.authService)
	if err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Could not parse authorization")
		return
	}

	httpClient := new(http.Client)
	if err = userConfig.GetAccessToken(token, verifier, a.authService, httpClient); err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting access token")
		return
	}

	if err := a.db.DeleteStruct(&authSession); err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error deleting session")
		return
	}

	c.SetCookie(authIDCookieName, "", 0, "/", c.Request.Host, false, true)

	u := &data.User{
		Username:          format.NormalizeString(userConfig.AccessValues.Get("screen_name")),
		AccessTokenKey:    userConfig.AccessTokenKey,
		AccessTokenSecret: userConfig.AccessTokenSecret,
		UpdatedAt:         time.Now().UTC(),
	}

	if err = a.db.Save(u); err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error saving authenticated user")
		return
	}

	p, err := a.twClient.GetUserDetails(ctx, u)
	if err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting user twitter details")
		return
	}

	if err = a.db.Save(p); err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error saving authenticated user profile")
		return
	}

	c.SetCookie(userIDCookieName, u.Username, userCookieDuration, "/", c.Request.Host, false, true)
	c.Redirect(http.StatusSeeOther, "/view/dash")
}

func (a *App) logOutHandler(c *gin.Context) {
	c.SetCookie(userIDCookieName, "", -1, "/", c.Request.Host, false, true)
	c.Redirect(http.StatusSeeOther, "/")
}

func userConfigToString(config *oauth1a.UserConfig) string {
	b, _ := json.Marshal(config)
	return hex.EncodeToString(b)
}

func userConfigFromString(content string) (conf *oauth1a.UserConfig, err error) {
	b, e := hex.DecodeString(content)
	if e != nil {
		return nil, e
	}
	conf = &oauth1a.UserConfig{}
	if e := json.Unmarshal(b, conf); e != nil {
		return nil, e
	}
	return
}

func authRequired(isJSON bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, _ := c.Cookie(userIDCookieName)
		if username == "" {
			if isJSON {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message": "User not authenticated",
					"status":  "Unauthorized",
				})
			} else {
				c.Redirect(http.StatusSeeOther, "/")
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

func (a *App) getUser(c *gin.Context) (*data.User, error) {
	username, _ := c.Cookie(userIDCookieName)
	if username == "" {
		return nil, errors.New("nil auth cookie")
	}

	var usr data.User
	if err := a.db.One("Username", username, &usr); err != nil || usr.Username == "" {
		return nil, errors.Wrapf(err, "error getting authenticated user: %s", username)
	}

	return &usr, nil
}
