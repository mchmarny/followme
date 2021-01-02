package app

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/gin-gonic/gin"
	"github.com/kurrik/oauth1a"
	"github.com/mchmarny/followme/internal/data"
	"github.com/mchmarny/followme/internal/twitter"
	"github.com/mchmarny/followme/pkg/url"
	"github.com/pkg/errors"
)

// NewApp creates a new instance of the app
func NewApp(dbPath, key, secret, url, version string, port int, dev bool) (*App, error) {
	if key == "" || secret == "" || version == "" {
		return nil, errors.New("key, secret, and version required")
	}

	// log
	logger := log.New(os.Stdout, "", 0)

	// data
	db, err := data.GetDB(dbPath)
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
		appURL:             fmt.Sprintf("%s:%d", url, port),
		devMode:            dev,
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
	appURL             string
	devMode            bool
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
	if err := a.setStaticContent(r); err != nil {
		return err
	}

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
		view.GET("/report", a.reportHandler)
	}

	data := r.Group("/data")
	data.Use(authRequired(true))
	{
		data.GET("/dash", a.dashboardQueryHandler)
		data.GET("/day/:day/list/:list/page/:page", a.dayQueryHandler)
		data.GET("/report/:id", a.reportDataHandler)
	}

	// signals
	done := make(chan os.Signal, 1)
	serverErr := make(chan error, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// start
	go func() {
		a.logger.Printf("Listening: %s \n", a.hostPort)
		if err := r.Run(a.hostPort); err != nil {
			serverErr <- errors.Wrap(err, "error while running app server")
		}
	}()

	time.Sleep(2 * time.Second)
	a.logger.Printf("Opening: %s", a.appURL)
	if err := url.Open(a.appURL); err != nil {
		return errors.Wrap(err, "error opening URL")
	}

	for {
		select {
		case sig := <-done:
			a.logger.Printf("\nClosing: %v", sig)
			return nil
		case err := <-serverErr:
			return err
		}
	}
}

func (a *App) setStaticContent(r *gin.Engine) error {
	if a.devMode {
		a.logger.Printf("loading external static resources")
		r.LoadHTMLGlob("./web/template/*")
		r.Static("/static", "./web/static")
		r.StaticFile("/favicon.ico", "./web/static/img/favicon.ico")
		return nil
	}

	// templates
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

	return nil
}

func (a *App) getState(username, isoDate string) (*data.DailyState, error) {
	key := data.GetDailyStateKeyISO(username, isoDate)
	var s data.DailyState
	if err := a.db.One("Key", key, &s); err != nil {
		if err != storm.ErrNotFound {
			return nil, errors.Wrapf(err, "error getting state for %s", key)
		}
		s = data.DailyState{
			Key:      key,
			Username: username,
			StateOn:  isoDate,
		}
	}
	return &s, nil
}

// errJSONAndAbort throws JSON error and abort prevents pending handlers from being called
func (a *App) errJSONAndAbort(c *gin.Context, err error) {
	a.logger.Printf("error while processing JSON request: %v", err)
	code := http.StatusInternalServerError
	errMsg := strings.ToLower(err.Error())
	msg := err.Error()

	if strings.Contains(errMsg, strings.ToLower("401 unauthorized")) {
		code = http.StatusUnauthorized
		msg = "Unauthorized, please login again."
	}

	if strings.Contains(errMsg, strings.ToLower("429 too many requests")) {
		code = http.StatusTooManyRequests
		msg = "Too Many Requests, please wait a few minutes and try again."
	}

	c.AbortWithStatusJSON(code, gin.H{
		"message": msg,
		"status":  "Error",
	})
}
