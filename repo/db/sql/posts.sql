-- name: InsertPost :one
INSERT INTO
    posts (image_filenames, caption, timestamp, position, photo_count, is_posted)
VALUES
    (?, ?, ?, ?, ?, ?)
RETURNING
    *;

-- name: GetAllPosts :many
SELECT
    *
FROM
    posts;

-- name: GetPostById :one
SELECT
    *
FROM
    posts
WHERE
    id = ?;

-- name: DeletePost :exec
DELETE FROM
    posts
WHERE
    id = ?;

-- name: GetLastPositionOfUnpostedPost :one
SELECT
    position
FROM
    posts
WHERE
    is_posted = 0
ORDER BY
    position DESC
LIMIT
    1;

-- name: UpdatePostCaption :exec
UPDATE
    posts
SET
    caption = ?
WHERE
    id = ?;

-- name: UpdatePostPosition :exec
UPDATE
    posts
SET
    position = ?
WHERE
    id = ?;

-- name: GetPostByPosition :one
SELECT
    *
FROM
    posts
WHERE
    position = ?
AND
    is_posted = 0;

-- name: UpdateIsPostedValueOfPost :exec
UPDATE
    posts
SET
    is_posted = ?,
    position = ?,
    posted_at = ?
WHERE
    id = ?;

-- name: GetPostToPost :one
SELECT
    *
FROM
    posts
WHERE
    is_posted = 0
    AND position = 1
LIMIT
    1;

-- name: GetUnpostedPosts :many
SELECT
    *
FROM
    posts
WHERE
    is_posted = 0
ORDER BY
    position ASC;
