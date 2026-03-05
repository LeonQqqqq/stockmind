package store

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"stockmind-go/internal/model"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS experiences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL DEFAULT 'insight',
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			tags TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS opinions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			author TEXT NOT NULL,
			content TEXT NOT NULL,
			tags TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}

	// Add type column to existing experiences table if missing
	db.Exec("ALTER TABLE experiences ADD COLUMN type TEXT NOT NULL DEFAULT 'insight'")
	// Create index after ensuring column exists
	db.Exec("CREATE INDEX IF NOT EXISTS idx_experiences_type ON experiences(type)")

	return nil
}

// === Sessions ===

func (s *SQLiteStore) CreateSession(id, title string) error {
	_, err := s.db.Exec("INSERT INTO sessions (id, title) VALUES (?, ?)", id, title)
	return err
}

func (s *SQLiteStore) GetSession(id string) (*model.Session, error) {
	var sess model.Session
	err := s.db.QueryRow("SELECT id, title, created_at, updated_at FROM sessions WHERE id = ?", id).
		Scan(&sess.ID, &sess.Title, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *SQLiteStore) ListSessions() ([]model.Session, error) {
	rows, err := s.db.Query("SELECT id, title, created_at, updated_at FROM sessions ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []model.Session
	for rows.Next() {
		var sess model.Session
		if err := rows.Scan(&sess.ID, &sess.Title, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

func (s *SQLiteStore) UpdateSessionTitle(id, title string) error {
	_, err := s.db.Exec("UPDATE sessions SET title = ?, updated_at = ? WHERE id = ?", title, time.Now(), id)
	return err
}

func (s *SQLiteStore) DeleteSession(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	tx.Exec("DELETE FROM messages WHERE session_id = ?", id)
	tx.Exec("DELETE FROM sessions WHERE id = ?", id)
	return tx.Commit()
}

// === Messages ===

func (s *SQLiteStore) SaveMessage(sessionID, role, content string) error {
	_, err := s.db.Exec("INSERT INTO messages (session_id, role, content) VALUES (?, ?, ?)",
		sessionID, role, content)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("UPDATE sessions SET updated_at = ? WHERE id = ?", time.Now(), sessionID)
	return err
}

func (s *SQLiteStore) GetMessages(sessionID string) ([]model.Message, error) {
	rows, err := s.db.Query(
		"SELECT id, session_id, role, content, created_at FROM messages WHERE session_id = ? ORDER BY id",
		sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

// === Experiences ===

func (s *SQLiteStore) CreateExperience(expType, title, content, tags string) (int64, error) {
	if expType == "" {
		expType = "insight"
	}
	res, err := s.db.Exec("INSERT INTO experiences (type, title, content, tags) VALUES (?, ?, ?, ?)",
		expType, title, content, tags)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *SQLiteStore) ListExperiences() ([]model.Experience, error) {
	rows, err := s.db.Query("SELECT id, type, title, content, tags, created_at, updated_at FROM experiences ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var exps []model.Experience
	for rows.Next() {
		var e model.Experience
		if err := rows.Scan(&e.ID, &e.Type, &e.Title, &e.Content, &e.Tags, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		exps = append(exps, e)
	}
	return exps, nil
}

func (s *SQLiteStore) UpdateExperience(id int64, title, content, tags string) error {
	_, err := s.db.Exec("UPDATE experiences SET title=?, content=?, tags=?, updated_at=? WHERE id=?",
		title, content, tags, time.Now(), id)
	return err
}

func (s *SQLiteStore) DeleteExperience(id int64) error {
	_, err := s.db.Exec("DELETE FROM experiences WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) SearchExperiences(keyword string) ([]model.Experience, error) {
	rows, err := s.db.Query(
		"SELECT id, type, title, content, tags, created_at, updated_at FROM experiences WHERE title LIKE ? OR content LIKE ? OR tags LIKE ? ORDER BY updated_at DESC",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var exps []model.Experience
	for rows.Next() {
		var e model.Experience
		if err := rows.Scan(&e.ID, &e.Type, &e.Title, &e.Content, &e.Tags, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		exps = append(exps, e)
	}
	return exps, nil
}

// === Opinions ===

func (s *SQLiteStore) CreateOpinion(author, content, tags string) (int64, error) {
	res, err := s.db.Exec("INSERT INTO opinions (author, content, tags) VALUES (?, ?, ?)",
		author, content, tags)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *SQLiteStore) ListOpinions(author string) ([]model.Opinion, error) {
	var rows *sql.Rows
	var err error
	if author != "" {
		rows, err = s.db.Query("SELECT id, author, content, tags, created_at FROM opinions WHERE author = ? ORDER BY created_at DESC", author)
	} else {
		rows, err = s.db.Query("SELECT id, author, content, tags, created_at FROM opinions ORDER BY created_at DESC")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ops []model.Opinion
	for rows.Next() {
		var o model.Opinion
		if err := rows.Scan(&o.ID, &o.Author, &o.Content, &o.Tags, &o.CreatedAt); err != nil {
			return nil, err
		}
		ops = append(ops, o)
	}
	return ops, nil
}

func (s *SQLiteStore) SearchOpinions(keyword string) ([]model.Opinion, error) {
	rows, err := s.db.Query(
		"SELECT id, author, content, tags, created_at FROM opinions WHERE author LIKE ? OR content LIKE ? OR tags LIKE ? ORDER BY created_at DESC",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ops []model.Opinion
	for rows.Next() {
		var o model.Opinion
		if err := rows.Scan(&o.ID, &o.Author, &o.Content, &o.Tags, &o.CreatedAt); err != nil {
			return nil, err
		}
		ops = append(ops, o)
	}
	return ops, nil
}

func (s *SQLiteStore) ListAuthors() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT author FROM opinions ORDER BY author")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var authors []string
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, err
		}
		authors = append(authors, a)
	}
	return authors, nil
}

func (s *SQLiteStore) DeleteOpinion(id int64) error {
	_, err := s.db.Exec("DELETE FROM opinions WHERE id = ?", id)
	return err
}
