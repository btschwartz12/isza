package server

import (
	"fmt"

	"github.com/btschwartz12/isza/repo"
	"github.com/btschwartz12/isza/server/api"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	router *chi.Mux
	rpo    *repo.Repo
	logger *zap.SugaredLogger
}

const (
	MaxPostUploadMb   = 50
	MaxPostUploadSize = MaxPostUploadMb << 20
)

func (s *Server) Init(
	logger *zap.SugaredLogger,
	varDir,
	authToken,
	instaUsername,
	instaPassword,
	instaWorkingDir string,
) error {
	r, err := repo.NewRepo(logger, varDir)
	if err != nil {
		return fmt.Errorf("error creating repo: %w", err)
	}
	s.rpo = r
	s.logger = logger
	s.router = chi.NewRouter()
	s.router.Get("/", s.home)
	s.router.Get("/post", s.addPostPage)
	s.router.Post("/post", s.uploadPostHandler)
	s.router.Get("/post/{id}/edit", s.editPostPage)
	s.router.Post("/post/{id}/edit", s.editPostHandler)
	s.router.Get("/post/{id}/move", s.movePostHandler)
	s.router.Get("/static/posts/{filename}", s.serveImageHandler)

	apiServer := &api.ApiServer{}
	err = apiServer.Init(logger, r, "/api", authToken, instaUsername, instaPassword, instaWorkingDir)
	if err != nil {
		return fmt.Errorf("error initializing api server: %w", err)
	}
	s.router.Mount("/api", apiServer.GetRouter())
	return nil
}

func (s *Server) Router() chi.Router {
	return s.router
}
