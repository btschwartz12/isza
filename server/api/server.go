package api

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	"github.com/btschwartz12/isza/repo"
	"github.com/btschwartz12/isza/server/api/swagger"
)

type ApiServer struct {
	router          *chi.Mux
	logger          *zap.SugaredLogger
	rpo             *repo.Repo
	token           string
	instaUsername   string
	instaPassword   string
	instaWorkingDir string
}

func (s *ApiServer) Init(
	logger *zap.SugaredLogger,
	rpo *repo.Repo,
	prefix,
	authToken,
	instaUsername,
	instaPassword,
	instaWorkingDir string,
) error {
	s.logger = logger
	s.router = chi.NewRouter()
	s.rpo = rpo
	s.token = authToken
	s.instaUsername = instaUsername
	s.instaPassword = instaPassword
	
	instaAbsDir, err := filepath.Abs(instaWorkingDir)
	if err != nil {
		return fmt.Errorf("error getting absolute path for instagram working directory: %w", err)
	}
	s.instaWorkingDir = instaAbsDir

	s.router.Get("/", http.RedirectHandler(fmt.Sprintf("%s/swagger/index.html", prefix), http.StatusMovedPermanently).ServeHTTP)
	s.router.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(swagger.SwaggerJSON)
	})
	s.router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(fmt.Sprintf("%s/swagger.json", prefix))))

	s.router.Get("/posts", s.getAllPostsHandler)
	s.router.Get("/posts/{id}", s.getPostHandler)
	s.router.Group(func(rr chi.Router) {
		rr.Use(s.tokenMiddleware)
		rr.Delete("/posts/{id}", s.deletePostHandler)
		rr.Post("/posts/make_post", s.makePostHandler)
		rr.Post("/posts/{id}/unpost", s.setPostAsUnpostedHandler)
		rr.Post("/posts/clean_positions", s.cleanPositionsHandler)
	})

	return nil
}

func (s *ApiServer) GetRouter() chi.Router {
	return s.router
}
