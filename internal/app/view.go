package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/followme/internal/data"
	"github.com/pkg/errors"
)

func (a *App) defaultHandler(c *gin.Context) {
	uid, _ := c.Cookie(userIDCookieName)
	if uid != "" {
		a.logger.Printf("user already authenticated -> view")
		c.Redirect(http.StatusSeeOther, "/view/dash")
		return
	}

	c.HTML(http.StatusOK, "index", gin.H{
		"version": a.appVersion,
	})
}

func (a *App) dashboardHandler(c *gin.Context) {
	profile, err := a.getUserProfile(c)
	if err != nil {
		a.logger.Printf("error getting profile: %v", err)
		a.logOutHandler(c)
		return
	}

	c.HTML(http.StatusOK, "dash", gin.H{
		"user":    profile,
		"version": a.appVersion,
		"refresh": c.Query("refresh"),
	})
}

func (a *App) getUserProfile(c *gin.Context) (*data.Profile, error) {
	username, _ := c.Cookie(userIDCookieName)
	if username == "" {
		return nil, errors.New("no auth cookie")
	}

	var p data.Profile
	if err := a.db.One("Username", username, &p); err != nil || p.Username == "" {
		return nil, errors.Wrapf(err, "error getting profile for: %v", username)
	}

	return &p, nil
}

func (a *App) viewErrorHandler(c *gin.Context, code int, err error, msg string) {
	a.logger.Printf("Error: %v - Msg: %s", err, msg)
	c.HTML(code, "error", gin.H{
		"code": code,
		"msg":  msg,
	})
	c.Abort()
}
