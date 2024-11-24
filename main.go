package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	flags "github.com/jessevdk/go-flags"
	"go.uber.org/zap"

	"github.com/btschwartz12/isza/server"
)

type arguments struct {
	Port            int    `short:"p" long:"port" description:"Port to listen on" default:"8000"`
	VarDir          string `short:"v" long:"var-dir" env:"ISZA_VAR_DIR" description:"Directory to store data"`
	DevLogging      bool   `short:"d" long:"dev-logging" description:"Enable development logging"`
	AuthToken       string `short:"t" long:"auth-token" env:"ISZA_AUTH_TOKEN" description:"Authorization token"`
	InstaUsername   string `short:"u" long:"insta-username" env:"ISZA_INSTA_USERNAME" description:"Instagram username"`
	InstaPassword   string `short:"w" long:"insta-password" env:"ISZA_INSTA_PASSWORD" description:"Instagram password"`
	InstaWorkingDir string `short:"i" long:"insta-working-dir" env:"ISZA_INSTA_WORKING_DIR" description:"Instagram working directory"`
}

var args arguments

func main() {
	_, err := flags.Parse(&args)
	if err != nil {
		panic(fmt.Errorf("error parsing flags: %s", err))
	}

	if args.VarDir == "" {
		panic("var dir is required")
	}

	if args.AuthToken == "" {
		panic("auth token is required")
	}

	if args.InstaUsername == "" {
		panic("instagram username is required")
	}

	if args.InstaPassword == "" {
		panic("instagram password is required")
	}

	if args.InstaWorkingDir == "" {
		panic("instagram working directory is required")
	}

	

	var l *zap.Logger
	if args.DevLogging {
		l, _ = zap.NewDevelopment()
	} else {
		l, _ = zap.NewProduction()
	}
	logger := l.Sugar()

	s := &server.Server{}
	err = s.Init(logger, args.VarDir, args.AuthToken, args.InstaUsername, args.InstaPassword, args.InstaWorkingDir)
	if err != nil {
		logger.Fatalw("Error initializing server", "error", err)
	}

	r := chi.NewRouter()
	r.Mount("/", s.Router())

	errChan := make(chan error)
	go func() {
		logger.Infow("Starting server", "port", args.Port)
		errChan <- http.ListenAndServe(fmt.Sprintf(":%d", args.Port), r)
	}()
	err = <-errChan
	logger.Fatalw("http server failed", "error", err)
}
