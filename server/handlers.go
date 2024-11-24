package server

import (
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/btschwartz12/isza/assets"
	"github.com/btschwartz12/isza/repo"
	"github.com/go-chi/chi/v5"
)

var (
	homeTmpl = template.Must(template.ParseFS(
		assets.Templates,
		"templates/home.html.tmpl",
	))

	addPostTmpl = template.Must(template.ParseFS(
		assets.Templates,
		"templates/addpost.html.tmpl",
	))

	funcMap = template.FuncMap{
		"add1": func(i int) int {
			return i + 1
		},
	}

	editPostTmpl = template.Must(template.New("editpost.html.tmpl").Funcs(funcMap).ParseFS(
		assets.Templates,
		"templates/editpost.html.tmpl",
	))
)

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	posts, err := s.rpo.GetAllPosts(r.Context())
	if err != nil {
		s.logger.Errorw("error getting all posts", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var queuePosts []repo.Post
	var stackPosts []repo.Post

	for _, post := range posts {
		if post.IsPosted {
			stackPosts = append(stackPosts, post)
		} else {
			queuePosts = append(queuePosts, post)
		}
	}

	sort.Slice(queuePosts, func(i, j int) bool {
		return queuePosts[i].Position < queuePosts[j].Position
	})

	sort.Slice(stackPosts, func(i, j int) bool {
		return stackPosts[i].PostedAt.MustGet().Time.After(stackPosts[j].PostedAt.MustGet().Time)
	})

	data := struct {
		InstagramAccountURL string
		DailyPostTimesEST   string
		QueuePosts          []repo.Post
		StackPosts          []repo.Post
	}{
		InstagramAccountURL: "https://instagram.com/youraccount", // Dummy variable
		DailyPostTimesEST:   "12:00 PM, 6:00 PM",                 // Dummy variable
		QueuePosts:          queuePosts,
		StackPosts:          stackPosts,
	}

	err = homeTmpl.Execute(w, data)
	if err != nil {
		s.logger.Errorw("error rendering home template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.logger.Infow("home page served")
}

func (s *Server) editPostPage(w http.ResponseWriter, r *http.Request) {
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
		s.logger.Errorw("error getting post", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = editPostTmpl.Execute(w, post)
	if err != nil {
		s.logger.Errorw("error rendering edit post template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) editPostHandler(w http.ResponseWriter, r *http.Request) {
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

	caption := r.FormValue("caption")
	if caption == "" {
		http.Error(w, "Caption is required", http.StatusBadRequest)
		return
	}

	err = s.rpo.UpdatePostCaption(r.Context(), id, caption)
	if err != nil {
		s.logger.Errorw("error updating post caption", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) addPostPage(w http.ResponseWriter, r *http.Request) {
	err := addPostTmpl.Execute(w, nil)
	if err != nil {
		s.logger.Errorw("error rendering add post template", "error", err)
	}
	s.logger.Infow("add post page served")
}

func (s *Server) uploadPostHandler(w http.ResponseWriter, r *http.Request) {
	caption := r.FormValue("caption")
	if caption == "" {
		http.Error(w, "Caption is required", http.StatusBadRequest)
		return
	}

	err := r.ParseMultipartForm(MaxPostUploadSize)
	if err != nil {
		s.logger.Errorw("error parsing form", "error", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	files := make([]repo.UploadFile, 0)
	for i := 1; i <= 5; i++ {
		paramName := "file_" + string(rune('0'+i))
		if fheaders, ok := r.MultipartForm.File[paramName]; ok && len(fheaders) > 0 {
			fheader := fheaders[0]
			file, err := fheader.Open()
			if err != nil {
				s.logger.Errorw("error opening file", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			files = append(files, repo.UploadFile{
				Header: fheader,
				File:   &file,
			})
		}
	}

	if len(files) == 0 {
		http.Error(w, "At least one file is required", http.StatusBadRequest)
		return
	}

	_, err = s.rpo.InsertPost(r.Context(), caption, files)
	if err != nil {
		s.logger.Errorw("error inserting post", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) movePostHandler(w http.ResponseWriter, r *http.Request) {
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

	direction := r.FormValue("direction")
	if direction != "up" && direction != "down" {
		http.Error(w, "Invalid direction", http.StatusBadRequest)
		return
	}
	up := direction == "up"

	err = s.rpo.MovePost(r.Context(), id, up)
	if err != nil {
		s.logger.Errorw("error moving post", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) serveImageHandler(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		http.Error(w, "filename is required", http.StatusBadRequest)
	}

	fullUrl := s.rpo.GetPathForPost(filename)
	http.ServeFile(w, r, fullUrl)
}
