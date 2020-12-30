package app

import (
	"fmt"
	"log"
	"os"

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
		db:            db,
		twClient:      t,
		authService:   as,
		logger:        logger,
		appVersion:    version,
		hostPort:      fmt.Sprintf("0.0.0.0:%d", port),
		maxEventLimit: 20, // TODO: parameterize
	}, nil
}

// App represents the app
type App struct {
	db            *storm.DB
	twClient      *twitter.Twitter
	logger        *log.Logger
	authService   *oauth1a.Service
	hostPort      string
	appVersion    string
	maxEventLimit int
}

// Run starts the app and blocks while running.
func (a *App) Run() error {
	gin.SetMode(gin.ReleaseMode)

	// cleanup
	defer a.db.Close()

	// router
	r := gin.Default()
	r.Use(gin.Recovery())

	// static
	r.LoadHTMLGlob("./web/template/*")
	r.Static("/static", "./web/static")
	r.StaticFile("/favicon.ico", "./web/static/img/favicon.ico")

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
	}

	// port
	a.logger.Printf("App starting: %s \n", a.hostPort)
	if err := r.Run(a.hostPort); err != nil {
		// TODO: gracefull shotdown
		return errors.Wrap(err, "error while running app server")
	}
	return nil
}
