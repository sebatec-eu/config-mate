package database

import (
	"context"
	"net/http"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type mockDataDir struct {
	dir string
}

func (m mockDataDir) DataDir() string {
	return m.dir
}

// Mock hostsharing data
func mockHostsharingDomain(dataDir string) dataDirResolver {
	return mockDataDir{dir: dataDir}
}

func TestSetSQLiteDsnDefault(t *testing.T) {
	t.Run("should set default SQLite DSN if empty", func(t *testing.T) {
		config := &Config{Type: SQLite, Dsn: ""}
		setSQLiteDsnDefault(config, nil)

		if config.Dsn != "./data.db" {
			t.Errorf("Expected default DSN to be './data.db', got '%s'", config.Dsn)
		}
	})

	t.Run("should not override custom SQLite DSN", func(t *testing.T) {
		config := &Config{Type: SQLite, Dsn: "/custom/path/data.db"}
		setSQLiteDsnDefault(config, nil)

		if config.Dsn != "/custom/path/data.db" {
			t.Errorf("Expected DSN to remain '/custom/path/data.db', got '%s'", config.Dsn)
		}
	})

	t.Run("should use Hostsharing data directory if available", func(t *testing.T) {
		config := &Config{Type: SQLite, Dsn: ""}
		setSQLiteDsnDefault(config, mockHostsharingDomain("/tmp/hostsharing/data"))

		expected := "/tmp/hostsharing/data/data.db"
		if config.Dsn != expected {
			t.Errorf("Expected DSN to be '%s', got '%s'", expected, config.Dsn)
		}
	})

}

func TestOpen(t *testing.T) {
	t.Run("should open SQLite database with custom DSN", func(t *testing.T) {
		config := Config{Type: SQLite, Dsn: ":memory:", Debug: false}
		db, err := Open(config)
		if err != nil {
			t.Fatalf("Failed to open SQLite database: %v", err)
		}
		if db == nil {
			t.Fatal("Expected database connection, got nil")
		}
	})

	t.Run("should fail if MySQL DSN is empty", func(t *testing.T) {
		config := Config{Type: MySQL, Dsn: "", Debug: false}
		db, err := Open(config)
		if err == nil {
			t.Error("Expected error when MySQL DSN is empty, got nil")
		}
		if db != nil {
			t.Error("Expected nil database connection, got non-nil")
		}
	})

	t.Run("should fail for unsupported database type", func(t *testing.T) {
		config := Config{Type: "postgres", Dsn: "", Debug: false}
		db, err := Open(config)
		if err == nil {
			t.Error("Expected error for unsupported database type, got nil")
		}
		if db != nil {
			t.Error("Expected nil database connection, got non-nil")
		}
	})
}

func TestContextHelpers(t *testing.T) {
	t.Run("should set and get database from context", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to open in-memory SQLite database: %v", err)
		}

		ctx := context.Background()
		ctx = Set(ctx, db)

		retrievedDB := Get(ctx)
		if retrievedDB != db {
			t.Error("Expected retrieved database to match the one set in context")
		}
	})

	t.Run("should panic if database not in context", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when database is not in context, but no panic occurred")
			}
		}()

		ctx := context.Background()
		Get(ctx)
	})
}

func TestSetMiddleware(t *testing.T) {
	t.Run("should set database in request context", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to open in-memory SQLite database: %v", err)
		}

		middleware := SetMiddleware(db)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			retrievedDB := Get(r.Context())
			if retrievedDB != db {
				t.Error("Expected retrieved database to match the one set in middleware")
			}
		}))

		// Simulate an http request
		req := &http.Request{}
		handler.ServeHTTP(nil, req)
	})
}
