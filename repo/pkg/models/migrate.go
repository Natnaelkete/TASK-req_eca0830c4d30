package models

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SchemaMigration tracks SQL migration files that have been applied.
type SchemaMigration struct {
	Filename  string `gorm:"primaryKey;size:255"`
	AppliedAt int64  `gorm:"not null"`
}

// TableName pins the GORM table name so it is not pluralised.
func (SchemaMigration) TableName() string { return "schema_migrations" }

// RunSQLMigrations applies every `*.sql` file in migrationsDir whose filename
// has not yet been recorded in `schema_migrations`. Files are executed in
// lexicographic order, which mirrors the `NNN_description.sql` convention.
//
// This is deliberately idempotent: an already-applied file is skipped, and
// the table `schema_migrations` is created on demand. Callers invoke this
// once during process start-up, after the GORM connection is established,
// so a clean environment boots with the full prompt-required schema
// (partitioning, archive table, etc.) without manual intervention.
func RunSQLMigrations(db *gorm.DB, migrationsDir string) error {
	info, err := os.Stat(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Missing directory is a no-op: deployments without bundled
			// SQL files (e.g. tests, local dev) rely on AutoMigrate.
			return nil
		}
		return fmt.Errorf("stat migrations dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("migrations path is not a directory: %s", migrationsDir)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		return nil
	}

	if err := db.AutoMigrate(&SchemaMigration{}); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	for _, name := range files {
		var applied SchemaMigration
		err := db.Where("filename = ?", name).First(&applied).Error
		if err == nil {
			continue
		}

		path := filepath.Join(migrationsDir, name)
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read migration %s: %w", name, readErr)
		}

		stmts := splitSQLStatements(string(raw))
		tx := db.Begin()
		for _, stmt := range stmts {
			if err := tx.Exec(stmt).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("apply migration %s: %w", name, err)
			}
		}
		if err := tx.Create(&SchemaMigration{Filename: name, AppliedAt: nowUnix()}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}
	return nil
}

// splitSQLStatements strips `--` line comments and splits on unquoted `;`.
// Our migration files use a small, controlled subset of SQL that does not
// embed semicolons inside string literals, so this is sufficient.
func splitSQLStatements(sql string) []string {
	var cleaned strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		cleaned.WriteString(line)
		cleaned.WriteString("\n")
	}
	parts := strings.Split(cleaned.String(), ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func nowUnix() int64 {
	return time.Now().Unix()
}
