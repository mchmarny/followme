package app

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/gin-gonic/gin"
	"github.com/kurrik/oauth1a"
	"github.com/mchmarny/followme/internal/data"
	"github.com/mchmarny/followme/internal/twitter"
	"github.com/pkg/errors"
)

// NewApp creates a new instance of the app
func NewApp(key, secret, url, version string, port int) (*App, error) {
	if key == "" || secret == "" || version == "" {
		return nil, errors.New("key, secret, and version required")
	}

	// log
	logger := log.New(os.Stdout, "app: ", 0)

	// data
	db, err := data.GetDB()
	if err != nil {
		return nil, errors.Wrap(err, "error getting DB")
	}

	// twitter
	t := twitter.NewTwitter(key, secret, logger)

	// oauth
	as := &oauth1a.Service{
		RequestURL:   "https://api.twitter.com/oauth/request_token",
		AuthorizeURL: "https://api.twitter.com/oauth/authorize",
		AccessURL:    "https://api.twitter.com/oauth/access_token",
		ClientConfig: &oauth1a.ClientConfig{
			ConsumerKey:    key,
			ConsumerSecret: secret,
			CallbackURL:    fmt.Sprintf("%s:%d/auth/callback", url, port),
		},
		Signer: new(oauth1a.HmacSha1Signer),
	}

	return &App{
		db:                 db,
		twClient:           t,
		authService:        as,
		logger:             logger,
		appVersion:         version,
		hostPort:           fmt.Sprintf("0.0.0.0:%d", port),
		pageSize:           10,                // TODO: parameterize
		userCookieDuration: 60 * 60 * 24 * 30, // month in sec
		maxSessionAge:      5.0,               // min
		sessionCookieAge:   5 * 60,            // maxSessionAge in secs
	}, nil
}

// App represents the app
type App struct {
	db                 *storm.DB
	twClient           *twitter.Twitter
	logger             *log.Logger
	authService        *oauth1a.Service
	hostPort           string
	appVersion         string
	pageSize           int
	userCookieDuration int
	maxSessionAge      float64
	sessionCookieAge   int
}

// Run starts the app and blocks while running.
func (a *App) Run() error {
	gin.SetMode(gin.ReleaseMode)

	// cleanup
	defer a.db.Close()

	// router
	r := gin.New()
	r.Use(gin.Recovery())

	// templates
	if err := a.setTemplates(r); err != nil {
		return err
	}

	// static
	r.StaticFS("/static", AssetFile())

	// fave
	r.GET("/favicon.ico", func(c *gin.Context) {
		b, err := Asset("web/static/img/favicon.ico")
		if err != nil {
			a.logger.Printf("error reading favicon.ico: %v", err)
			c.Abort()
			return
		}
		reader := bytes.NewReader(b)
		c.Header("Content-Type", "image/x-icon")
		http.ServeContent(c.Writer, c.Request, "favicon.ico", time.Now(), reader)
	})

	// routes
	r.GET("/", a.defaultHandler)

	// auth (authing itself)
	auth := r.Group("/auth")
	{
		auth.GET("/login", a.authLoginHandler)
		auth.GET("/callback", a.authCallbackHandler)
		auth.GET("/logout", a.logOutHandler)
	}

	// authenticated routes
	view := r.Group("/view")
	view.Use(authRequired(false))
	{
		view.GET("/dash", a.dashboardHandler)
		view.GET("/day/:day", a.dayHandler)
	}

	data := r.Group("/data")
	data.Use(authRequired(true))
	{
		data.GET("/dash", a.dashboardQueryHandler)
		data.GET("/day/:day/list/:list/page/:page", a.dayQueryHandler)
	}

	// port
	a.logger.Printf("App starting: %s \n", a.hostPort)
	if err := r.Run(a.hostPort); err != nil {
		return errors.Wrap(err, "error while running app server")
	}
	return nil
}

func (a *App) setTemplates(r *gin.Engine) error {
	templateFiles, err := AssetDir("web/template")
	if err != nil {
		return errors.Wrap(err, "error laoding tempalates")
	}

	mt := template.New("")
	for _, f := range templateFiles {
		p := path.Join("web/template", f)
		// a.logger.Printf("loading template: %s", p)
		b, err := Asset(p)
		if err != nil {
			return errors.Wrapf(err, "error getting asset from: %s", p)
		}
		t, err := mt.New(f).Parse(string(b))
		if err != nil {
			return errors.Wrapf(err, "error parsing tempalate: %s", p)
		}
		r.SetHTMLTemplate(t)
	}
	return nil
}
