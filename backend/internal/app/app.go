package app

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/gorm"

	"github.com/sskotezhov/maturin/config"
	"github.com/sskotezhov/maturin/internal/auth"
	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/email"
)

type App struct {
	Echo   *echo.Echo
	DB     *gorm.DB
	Redis  *redis.Client
	Config *config.Config
}

func New(cfg *config.Config, db *gorm.DB, rdb *redis.Client) (*App, error) {
	e := echo.New()
	e.HideBanner = true

	if err := migrate(db); err != nil {
		return nil, err
	}

	jwtCfg := cfg.Maturin.JWT
	smtpCfg := cfg.Maturin.SMTP

	//user
	userRepo := user.NewRepository(db)

	//email
	emailSender := email.NewSender(email.Config{
		Host:     smtpCfg.Host,
		Port:     smtpCfg.Port,
		User:     smtpCfg.User,
		Password: smtpCfg.Password,
	})

	//auth
	authSvc := auth.NewService(
		userRepo,
		rdb,
		emailSender,
		jwtCfg.Secret,
		time.Duration(jwtCfg.AccessTokenTTL)*time.Minute,
		time.Duration(jwtCfg.RefreshTokenTTL)*time.Minute,
	)

	//swagger генерю доки в /docs
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api/v1")
	auth.NewHandler(authSvc).Register(api.Group("/auth"))

	return &App{
		Echo:   e,
		DB:     db,
		Redis:  rdb,
		Config: cfg,
	}, nil
}

func migrate(db *gorm.DB) error {
	return user.Migrate(db)
}
