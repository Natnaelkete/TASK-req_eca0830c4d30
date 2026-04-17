package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaMigration_TableName(t *testing.T) {
	assert.Equal(t, "schema_migrations", SchemaMigration{}.TableName())
}

func TestSplitSQLStatements_StripsCommentsAndBlanks(t *testing.T) {
	sql := `-- 001_create.sql
-- another comment
CREATE TABLE foo (id INT);

-- next block
CREATE TABLE bar (id INT);
`
	stmts := splitSQLStatements(sql)
	assert.Len(t, stmts, 2)
	assert.Contains(t, stmts[0], "CREATE TABLE foo")
	assert.Contains(t, stmts[1], "CREATE TABLE bar")
}

func TestSplitSQLStatements_MultipleStatementsSingleLine(t *testing.T) {
	sql := `ALTER TABLE x DROP COLUMN y; ALTER TABLE x DROP COLUMN z;`
	stmts := splitSQLStatements(sql)
	assert.Len(t, stmts, 2)
}

func TestSplitSQLStatements_EmptyInput(t *testing.T) {
	assert.Empty(t, splitSQLStatements(""))
	assert.Empty(t, splitSQLStatements("-- only a comment\n"))
}

func TestSplitSQLStatements_PartitioningBlock(t *testing.T) {
	sql := `ALTER TABLE t PARTITION BY RANGE (TO_DAYS(event_time)) (
    PARTITION p1 VALUES LESS THAN (TO_DAYS('2026-02-01')),
    PARTITION pmax VALUES LESS THAN MAXVALUE
);`
	stmts := splitSQLStatements(sql)
	assert.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], "PARTITION BY RANGE")
}

func TestRunSQLMigrations_NonexistentDirIsNoop(t *testing.T) {
	err := RunSQLMigrations(nil, "nonexistent-dir-that-should-not-exist-xyz")
	assert.NoError(t, err)
}
