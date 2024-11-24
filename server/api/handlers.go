package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/btschwartz12/isza/instagram"
	"github.com/btschwartz12/isza/repo"
	"github.com/go-chi/chi/v5"
	_ "github.com/samber/mo"
)

// getAllPostsHandler godoc
// @Summary Get all posts
// @Description Get all posts
// @Tags posts
// @Produce json
// @Router /api/posts [get]
// @Success 200
func (s *ApiServer) getAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := s.rpo.GetAllPosts(r.Context())
	if err != nil {
		s.logger.Errorw("error getting all posts", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.MarshalIndent(posts, "", "\t")
	if err != nil {
		s.logger.Errorw("error marshalling posts", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// getPostHandler godoc
// @Summary Get a post
// @Description Get a post
// @Tags posts
// @Produce json
// @Param id path int true "Post ID"
// @Router /api/posts/{id} [get]
// @Success 200
func (s *ApiServer) getPostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Post ID", http.StatusBadRequest)
		return
	}

	post, err := s.rpo.GetPost(r.Context(), id)
	if err != nil {
		if err == repo.ErrPostNotFound {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("error getting post", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.MarshalIndent(post, "", "\t")
	if err != nil {
		s.logger.Errorw("error marshalling post", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// deletePostHandler godoc
// @Summary Delete a post
// @Description Delete a post
// @Tags posts
// @Param id path int true "Post ID"
// @Router /api/posts/{id} [delete]
// @Security Bearer
// @Success 204
func (s *ApiServer) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Post ID", http.StatusBadRequest)
		return
	}

	err = s.rpo.DeletePost(r.Context(), id)
	if err != nil {
		if err == repo.ErrPostNotFound {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("error deleting post", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	s.logger.Infow("post deleted", "id", id)
}

// makePostHandler godoc
// @Summary Set a post as posted
// @Description Set a post as posted
// @Tags posts
// @Router /api/posts/make_post [post]
// @Security Bearer
// @Success 204
func (s *ApiServer) makePostHandler(w http.ResponseWriter, r *http.Request) {
	post, err := s.rpo.GetPostToPost(r.Context())
	if err != nil {
		if err == repo.ErrPostNotFound {
			http.Error(w, "Nothing to post", http.StatusNotFound)
			return
		}
		s.logger.Errorw("error getting post to post", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = instagram.ExecutePost(r.Context(), s.logger, s.rpo, s.instaWorkingDir, s.instaUsername, s.instaPassword, post)
	if err != nil {
		s.logger.Errorw("error executing post", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = s.rpo.SetIsPostedValueOfPost(r.Context(), post.ID, true)
	if err != nil {
		if err == repo.ErrPostNotFound {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("error setting post as posted", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	s.logger.Infow("post set as posted", "id", post.ID)
}

// setPostAsUnpostedHandler godoc
// @Summary Set a post as unposted
// @Description Set a post as unposted
// @Tags posts
// @Param id path int true "Post ID"
// @Router /api/posts/{id}/unpost [post]
// @Security Bearer
// @Success 204
func (s *ApiServer) setPostAsUnpostedHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Post ID", http.StatusBadRequest)
		return
	}

	err = s.rpo.SetIsPostedValueOfPost(r.Context(), id, false)
	if err != nil {
		if err == repo.ErrPostNotFound {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("error setting post as unposted", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	s.logger.Infow("post set as unposted", "id", id)
}

// cleanPositionsHandler godoc
// @Summary Clean post positions
// @Description Clean post positions
// @Tags posts
// @Router /api/posts/clean_positions [post]
// @Security Bearer
// @Success 204
func (s *ApiServer) cleanPositionsHandler(w http.ResponseWriter, r *http.Request) {
	err := s.rpo.CleanPositions(r.Context())
	if err != nil {
		s.logger.Errorw("error cleaning post positions", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	s.logger.Infow("post positions cleaned")
}
