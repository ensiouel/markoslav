package app

import (
	"context"
	"fmt"
	"github.com/pressly/goose"
	"log"
	"markoslav/internal/bot"
	"markoslav/internal/bot/handler"
	"markoslav/internal/config"
	"markoslav/internal/service"
	"markoslav/internal/storage"
	"markoslav/internal/usecase"
	"markoslav/pkg/postgres"
	"os/signal"
	"syscall"
)

type App struct {
	conf config.Config
	bot  *bot.Bot
}

func New() *App {
	conf := config.New()

	return &App{
		conf: conf,
		bot:  bot.New(conf.Bot),
	}
}

func (app *App) Run() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pgConfig := postgres.Config{
		Host: app.conf.Postgres.Host, Port: app.conf.Postgres.Port, DB: app.conf.Postgres.DB,
		User: app.conf.Postgres.User, Password: app.conf.Postgres.Password,
	}
	pgClient, err := postgres.NewClient(ctx, pgConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err = migrate("up", "migration", pgConfig.String()); err != nil {
		log.Fatalf("migration error: %s", err)
	}

	captionStorage := storage.NewCaptionStorage(pgClient)
	captionService := service.NewCaptionService(captionStorage)

	imageService := service.NewImageService("static/Lobster-Regular.ttf")

	captionUsecase := usecase.NewCaptionUsecase(captionService, imageService)

	captionHandler := handler.NewCaptionHandler(app.bot.API, captionUsecase, app.conf.Bot.AdminList)

	go app.bot.Handle(captionHandler).
		Run()

	select {
	case <-ctx.Done():
		fmt.Println("graceful shutdown")
	}
}

func migrate(command string, dir string, dbstring string) error {
	db, err := goose.OpenDBWithDriver("postgres", dbstring)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = goose.Run(command, db, dir); err != nil {
		return err
	}

	return nil
}
