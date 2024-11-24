CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    image_filenames TEXT NOT NULL,
    caption TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    position INTEGER NOT NULL,
    photo_count INTEGER NOT NULL,
    is_posted INTEGER NOT NULL,
    posted_at TEXT DEFAULT NULL
)