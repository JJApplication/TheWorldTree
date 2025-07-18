package models

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	FullName    string    `json:"full_name" db:"full_name"`
	Description string    `json:"description" db:"description"`
	URL         string    `json:"url" db:"url"`
	Language    string    `json:"language" db:"language"`
	Stars       int       `json:"stars" db:"stars"`
	Forks       int       `json:"forks" db:"forks"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	SyncedAt    time.Time `json:"synced_at" db:"synced_at"`
}

type Commit struct {
	ID                 int       `json:"id" db:"id"`
	SHA                string    `json:"sha" db:"sha"`
	Message            string    `json:"message" db:"message"`
	AuthorName         string    `json:"author_name" db:"author_name"`
	AuthorEmail        string    `json:"author_email" db:"author_email"`
	CommitDate         time.Time `json:"commit_date" db:"commit_date"`
	RepositoryFullName string    `json:"repository_full_name" db:"repository_full_name"`
	SyncedAt           time.Time `json:"synced_at" db:"synced_at"`
}

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func (db *DB) createTables() error {
	repoQuery := `
	CREATE TABLE IF NOT EXISTS repositories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		full_name TEXT UNIQUE NOT NULL,
		description TEXT,
		url TEXT NOT NULL,
		language TEXT,
		stars INTEGER DEFAULT 0,
		forks INTEGER DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME,
		synced_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.conn.Exec(repoQuery)
	if err != nil {
		return err
	}

	commitQuery := `
	CREATE TABLE IF NOT EXISTS commits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sha TEXT NOT NULL,
		message TEXT NOT NULL,
		author_name TEXT NOT NULL,
		author_email TEXT NOT NULL,
		commit_date DATETIME NOT NULL,
		repository_full_name TEXT NOT NULL,
		synced_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(sha, repository_full_name),
		FOREIGN KEY (repository_full_name) REFERENCES repositories(full_name)
	);
	`
	_, err = db.conn.Exec(commitQuery)
	return err
}

func (db *DB) SaveRepository(repo *Repository) error {
	query := `
	INSERT OR REPLACE INTO repositories 
	(name, full_name, description, url, language, stars, forks, created_at, updated_at, synced_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query,
		repo.Name, repo.FullName, repo.Description, repo.URL,
		repo.Language, repo.Stars, repo.Forks,
		repo.CreatedAt, repo.UpdatedAt, time.Now())
	return err
}

func (db *DB) GetRepositories() ([]*Repository, error) {
	query := `SELECT id, name, full_name, description, url, language, stars, forks, created_at, updated_at, synced_at FROM repositories ORDER BY stars DESC`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repositories []*Repository
	for rows.Next() {
		repo := &Repository{}
		err := rows.Scan(&repo.ID, &repo.Name, &repo.FullName, &repo.Description,
			&repo.URL, &repo.Language, &repo.Stars, &repo.Forks,
			&repo.CreatedAt, &repo.UpdatedAt, &repo.SyncedAt)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, repo)
	}

	return repositories, nil
}

func (db *DB) GetRepositoryByName(fullName string) (*Repository, error) {
	query := `SELECT id, name, full_name, description, url, language, stars, forks, created_at, updated_at, synced_at FROM repositories WHERE full_name = ?`
	repo := &Repository{}
	err := db.conn.QueryRow(query, fullName).Scan(
		&repo.ID, &repo.Name, &repo.FullName, &repo.Description,
		&repo.URL, &repo.Language, &repo.Stars, &repo.Forks,
		&repo.CreatedAt, &repo.UpdatedAt, &repo.SyncedAt)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (db *DB) SaveCommit(commit *Commit) error {
	query := `
	INSERT OR REPLACE INTO commits 
	(sha, message, author_name, author_email, commit_date, repository_full_name, synced_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query,
		commit.SHA, commit.Message, commit.AuthorName, commit.AuthorEmail,
		commit.CommitDate, commit.RepositoryFullName, time.Now())
	return err
}

func (db *DB) GetCommits(repositoryFullName string, limit, offset int) ([]*Commit, error) {
	query := `SELECT id, sha, message, author_name, author_email, commit_date, repository_full_name, synced_at 
			  FROM commits 
			  WHERE repository_full_name = ? 
			  ORDER BY commit_date DESC 
			  LIMIT ? OFFSET ?`
	rows, err := db.conn.Query(query, repositoryFullName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []*Commit
	for rows.Next() {
		commit := &Commit{}
		err := rows.Scan(&commit.ID, &commit.SHA, &commit.Message, &commit.AuthorName,
			&commit.AuthorEmail, &commit.CommitDate, &commit.RepositoryFullName, &commit.SyncedAt)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func (db *DB) GetCommitCount(repositoryFullName string) (int, error) {
	query := `SELECT COUNT(*) FROM commits WHERE repository_full_name = ?`
	var count int
	err := db.conn.QueryRow(query, repositoryFullName).Scan(&count)
	return count, err
}

func (db *DB) Close() error {
	return db.conn.Close()
}