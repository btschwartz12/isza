// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql"
)

type Post struct {
	ID             int64
	ImageFilenames string
	Caption        string
	Timestamp      string
	Position       int64
	PhotoCount     int64
	IsPosted       int64
	PostedAt       sql.NullString
}
