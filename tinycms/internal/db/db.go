package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct{ DB *sql.DB }

type Page struct {
	ID              int64  `json:"id"`
	Slug            string `json:"slug"`
	Path            string `json:"path"`
	Title           string `json:"title"`
	MetaDescription string `json:"meta_description"`
	ContentType     string `json:"content_type"`
	Markdown        string `json:"markdown,omitempty"`
	Published       bool   `json:"published"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type Asset struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(dir(path), 0700); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	st := &Store{DB: db}
	return st, st.migrate()
}

func dir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}

func (s *Store) migrate() error {
	_, err := s.DB.Exec(`
CREATE TABLE IF NOT EXISTS pages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  slug TEXT NOT NULL UNIQUE,
  path TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL,
  meta_description TEXT NOT NULL DEFAULT '',
  content_type TEXT NOT NULL DEFAULT 'page',
  markdown TEXT NOT NULL DEFAULT '',
  published INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_pages_published_slug ON pages(published, slug);
CREATE TABLE IF NOT EXISTS assets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  path TEXT NOT NULL,
  url TEXT NOT NULL,
  size INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
INSERT INTO pages(slug,title,markdown,published)
SELECT 'home','Home','# Welcome to TinyCMS\n\nEdit this page from /admin.',1
WHERE NOT EXISTS (SELECT 1 FROM pages WHERE slug='home');`)
	if err != nil {
		return err
	}
	for _, stmt := range []string{
		`ALTER TABLE pages ADD COLUMN path TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE pages ADD COLUMN meta_description TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE pages ADD COLUMN content_type TEXT NOT NULL DEFAULT 'page'`,
	} {
		if _, err := s.DB.Exec(stmt); err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
	}
	_, err = s.DB.Exec(`
UPDATE pages SET path='/' WHERE slug='home' AND path='';
UPDATE pages SET path='/' || slug WHERE slug <> 'home' AND path='';
CREATE UNIQUE INDEX IF NOT EXISTS idx_pages_path ON pages(path);
CREATE INDEX IF NOT EXISTS idx_pages_published_path ON pages(published, path);`)
	return err
}

func (s *Store) ListPages(ctx context.Context) ([]Page, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id,slug,path,title,meta_description,content_type,published,created_at,updated_at FROM pages ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pages []Page
	for rows.Next() {
		var p Page
		var pub int
		if err := rows.Scan(&p.ID, &p.Slug, &p.Path, &p.Title, &p.MetaDescription, &p.ContentType, &pub, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Published = pub == 1
		pages = append(pages, p)
	}
	return pages, rows.Err()
}

func (s *Store) GetPage(ctx context.Context, slug string) (Page, error) {
	var p Page
	var pub int
	err := s.DB.QueryRowContext(ctx, `SELECT id,slug,path,title,meta_description,content_type,markdown,published,created_at,updated_at FROM pages WHERE slug=?`, slug).Scan(&p.ID, &p.Slug, &p.Path, &p.Title, &p.MetaDescription, &p.ContentType, &p.Markdown, &pub, &p.CreatedAt, &p.UpdatedAt)
	p.Published = pub == 1
	return p, err
}

func (s *Store) GetPublishedByPath(ctx context.Context, path string) (Page, error) {
	var p Page
	var pub int
	err := s.DB.QueryRowContext(ctx, `SELECT id,slug,path,title,meta_description,content_type,markdown,published,created_at,updated_at FROM pages WHERE path=? AND published=1`, path).Scan(&p.ID, &p.Slug, &p.Path, &p.Title, &p.MetaDescription, &p.ContentType, &p.Markdown, &pub, &p.CreatedAt, &p.UpdatedAt)
	p.Published = pub == 1
	return p, err
}

func (s *Store) SavePage(ctx context.Context, p Page) (Page, error) {
	if p.Slug == "" || p.Path == "" || p.Title == "" {
		return Page{}, errors.New("slug, path, and title required")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.DB.ExecContext(ctx, `INSERT INTO pages(slug,path,title,meta_description,content_type,markdown,published,updated_at) VALUES(?,?,?,?,?,?,?,?)
ON CONFLICT(slug) DO UPDATE SET path=excluded.path, title=excluded.title, meta_description=excluded.meta_description, content_type=excluded.content_type, markdown=excluded.markdown, published=excluded.published, updated_at=excluded.updated_at`, p.Slug, p.Path, p.Title, p.MetaDescription, p.ContentType, p.Markdown, boolInt(p.Published), now)
	if err != nil {
		return Page{}, err
	}
	return s.GetPage(ctx, p.Slug)
}

func (s *Store) DeletePage(ctx context.Context, slug string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM pages WHERE slug=? AND slug <> 'home'`, slug)
	return err
}

func (s *Store) InsertAsset(ctx context.Context, name, path, url string, size int64) (Asset, error) {
	res, err := s.DB.ExecContext(ctx, `INSERT INTO assets(name,path,url,size) VALUES(?,?,?,?)`, name, path, url, size)
	if err != nil {
		return Asset{}, err
	}
	id, _ := res.LastInsertId()
	return Asset{ID: id, Name: name, Path: path, URL: url, Size: size, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}, nil
}
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
