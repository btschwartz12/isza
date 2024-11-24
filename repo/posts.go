package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/btschwartz12/isza/repo/db"
	"github.com/google/uuid"
	"github.com/samber/mo"
)

var (
	EstTimezone         *time.Location
	allowedExtensionsRe = regexp.MustCompile(`\.(jpe?g|png|gif)$`)

	ErrStorageFull      = fmt.Errorf("storage full")
	ErrInvalidExtension = fmt.Errorf("invalid file extension")
	ErrPostNotFound     = fmt.Errorf("post not found")
)

func init() {
	var err error
	EstTimezone, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic(fmt.Errorf("failed to load timezone: %w", err))
	}
}

type EstTime struct {
	time.Time
}

func (t EstTime) String() string {
	return fmt.Sprintf("%s EST", t.In(EstTimezone).Format("2006-01-02 15:04:05"))
}

func (t EstTime) zulu() string {
	return t.Format(time.RFC3339)
}

type Post struct {
	ID             int64
	ImageFilenames []string
	Caption        string
	Timestamp      EstTime
	Position       int64
	PhotoCount     int64
	IsPosted       bool
	PostedAt       mo.Option[EstTime]
}

func (p *Post) fromDb(row *db.Post) {
	p.ID = row.ID
	p.Caption = row.Caption
	p.Position = row.Position
	p.PhotoCount = row.PhotoCount
	p.IsPosted = row.IsPosted == 1
	p.ImageFilenames = strings.Split(row.ImageFilenames, ",")
	t, _ := time.Parse(time.RFC3339, row.Timestamp)
	p.Timestamp = EstTime{t}
	if row.PostedAt.Valid {
		t, _ := time.Parse(time.RFC3339, row.PostedAt.String)
		p.PostedAt = mo.Some(EstTime{t})
	} else {
		p.PostedAt = mo.None[EstTime]()
	}
}

func (p *Post) toDb() db.InsertPostParams {
	return db.InsertPostParams{
		ImageFilenames: strings.Join(p.ImageFilenames, ","),
		Caption:        p.Caption,
		Position:       p.Position,
		PhotoCount:     p.PhotoCount,
		Timestamp:      p.Timestamp.zulu(),
		IsPosted:       boolToInt(p.IsPosted),
	}
}

type UploadFile struct {
	Header *multipart.FileHeader
	File   *multipart.File
}

func (r *Repo) InsertPost(
	ctx context.Context,
	caption string,
	files []UploadFile,
) (*Post, error) {
	if r.storageFull() {
		return nil, ErrStorageFull
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files uploaded")
	}
	fileNames := make([]string, len(files))
	for i, file := range files {
		ext := filepath.Ext(file.Header.Filename)
		if !allowedExtensionsRe.MatchString(ext) {
			return nil, ErrInvalidExtension
		}
		newName := uuid.New().String() + ext
		newPath := filepath.Join(r.varDir, postUploadDir, newName)
		newFile, err := os.Create(newPath)
		if err != nil {
			return nil, fmt.Errorf("error creating file: %w", err)
		}
		defer newFile.Close()
		if _, err := io.Copy(newFile, *file.File); err != nil {
			return nil, fmt.Errorf("error copying file: %w", err)
		}
		fileNames[i] = newName
	}
	position, err := r.GetLastPositionOfUnpostedPost(ctx)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			position = 0
		} else {
			return nil, fmt.Errorf("could not generate position: %w", err)
		}
	}
	post := &Post{
		Caption:        caption,
		Position:       position + 1,
		PhotoCount:     int64(len(files)),
		IsPosted:       false,
		Timestamp:      EstTime{time.Now()},
		ImageFilenames: fileNames,
	}
	q := db.New(r.db)
	row, err := q.InsertPost(ctx, post.toDb())
	if err != nil {
		return nil, fmt.Errorf("error inserting post: %w", err)
	}
	newPost := &Post{}
	newPost.fromDb(&row)
	return newPost, nil
}

func (r *Repo) GetAllPosts(ctx context.Context) ([]Post, error) {
	q := db.New(r.db)
	rows, err := q.GetAllPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting all posts: %w", err)
	}
	posts := make([]Post, len(rows))
	for i, row := range rows {
		posts[i].fromDb(&row)
	}
	return posts, nil
}

func (r *Repo) GetPost(ctx context.Context, id int64) (*Post, error) {
	q := db.New(r.db)
	row, err := q.GetPostById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("error getting post: %w", err)
	}
	post := &Post{}
	post.fromDb(&row)
	return post, nil
}

func (r *Repo) DeletePost(ctx context.Context, id int64) error {
	post, err := r.GetPost(ctx, id)
	if err != nil {
		return err
	}
	q := db.New(r.db)
	if err := q.DeletePost(ctx, id); err != nil {
		return fmt.Errorf("error deleting post: %w", err)
	}
	for _, filename := range post.ImageFilenames {
		err := os.Remove(filepath.Join(r.varDir, postUploadDir, filename))
		if err != nil {
			return fmt.Errorf("error deleting file: %w", err)
		}
	}
	err = r.CleanPositions(ctx)
	if err != nil {
		return fmt.Errorf("error cleaning positions: %w", err)
	}
	return nil
}

func (r *Repo) GetLastPositionOfUnpostedPost(ctx context.Context) (int64, error) {
	q := db.New(r.db)
	pos, err := q.GetLastPositionOfUnpostedPost(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrPostNotFound
		}
		return 0, fmt.Errorf("error getting last position of unposted post: %w", err)
	}
	return pos, nil
}

func (r *Repo) UpdatePostCaption(ctx context.Context, id int64, caption string) error {
	q := db.New(r.db)
	err := q.UpdatePostCaption(ctx, db.UpdatePostCaptionParams{
		Caption: caption,
		ID:      id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPostNotFound
		}
		return fmt.Errorf("error updating post caption: %w", err)
	}
	return nil
}

func (r *Repo) MovePost(ctx context.Context, id int64, up bool) error {
	q := db.New(r.db)
	post, err := r.GetPost(ctx, id)
	if err != nil {
		return err
	}
	lastPosition, err := q.GetLastPositionOfUnpostedPost(ctx)
	if err != nil {
		return fmt.Errorf("error getting last position of unposted post: %w", err)
	}
	var targetPosition int64
	if up {
		if post.Position == 1 {
			return nil
		}
		targetPosition = post.Position - 1
	} else {
		if post.Position == lastPosition {
			return nil
		}
		targetPosition = post.Position + 1
	}
	postAtTarget, err := q.GetPostByPosition(ctx, targetPosition)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("expected post at position %d to exist", targetPosition)
		}
		return fmt.Errorf("error getting post at target position: %w", err)
	}
	err = q.UpdatePostPosition(ctx, db.UpdatePostPositionParams{
		ID:       post.ID,
		Position: targetPosition,
	})
	if err != nil {
		return fmt.Errorf("error updating post position: %w", err)
	}
	err = q.UpdatePostPosition(ctx, db.UpdatePostPositionParams{
		ID:       postAtTarget.ID,
		Position: post.Position,
	})
	if err != nil {
		return fmt.Errorf("error updating post position: %w", err)
	}
	return nil
}

func (r *Repo) SetIsPostedValueOfPost(ctx context.Context, id int64, isPosted bool) error {
	lastPosition, err := r.GetLastPositionOfUnpostedPost(ctx)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			return ErrPostNotFound
		}
		return fmt.Errorf("error getting last position of unposted post: %w", err)
	}
	postedAt := sql.NullString{}
	if isPosted {
		now := EstTime{time.Now()}
		postedAt.Valid = true
		postedAt.String = now.zulu()
	} else {
		postedAt.Valid = false
	}

	q := db.New(r.db)
	err = q.UpdateIsPostedValueOfPost(ctx, db.UpdateIsPostedValueOfPostParams{
		ID:       id,
		IsPosted: boolToInt(isPosted),
		Position: lastPosition + 1,
		PostedAt: postedAt,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPostNotFound
		}
		return fmt.Errorf("error updating is posted value: %w", err)
	}
	err = r.CleanPositions(ctx)
	if err != nil {
		return fmt.Errorf("error cleaning positions: %w", err)
	}
	return nil
}

func (r *Repo) GetPostToPost(ctx context.Context) (*Post, error) {
	q := db.New(r.db)
	row, err := q.GetPostToPost(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, fmt.Errorf("error getting post to post: %w", err)
	}
	post := &Post{}
	post.fromDb(&row)
	return post, nil
}

func (r *Repo) CleanPositions(ctx context.Context) error {
	q := db.New(r.db)
	posts, err := q.GetUnpostedPosts(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("error getting unposted posts: %w", err)
	}
	for i, post := range posts {
		err := q.UpdatePostPosition(ctx, db.UpdatePostPositionParams{
			ID:       post.ID,
			Position: int64(i) + 1,
		})
		if err != nil {
			return fmt.Errorf("error updating post position: %w", err)
		}
	}
	return nil
}

func (r *Repo) GetPathForPost(filename string) string {
	return filepath.Join(r.varDir, postUploadDir, filename)
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
