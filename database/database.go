// Package database provides database connectivity and context management.
//
// It supports both SQLite and MySQL databases with automatic configuration
// for Hostsharing environments. The package provides context-based database
// access through HTTP middleware integration.
package database

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/sebatec-eu/config-mate/hostsharing"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DBType represents the database type.
type DBType string

const (
	// SQLite represents a SQLite database type.
	SQLite DBType = "sqlite"
	// MySQL represents a MySQL database type.
	MySQL DBType = "mysql"
)

// Config holds the database configuration.
//
// Type specifies the database backend (SQLite or MySQL).
// Dsn is the data source name (connection string). For SQLite, if empty,
// it defaults to "./data.db" or a path within the Hostsharing data directory.
// Debug enables SQL query logging.
type Config struct {
	Type  DBType
	Dsn   string
	Debug bool
}

type dataDirResolver interface {
	DataDir() string
}

func getDataDirResolver() dataDirResolver {
	dom, err := hostsharing.DomainByWorkingDir()
	if err != nil {
		return nil
	}

	return dom
}

var serviceNameFunc = hostsharing.ServiceName

func setSQLiteDsnDefault(c *Config, dataDirResolver dataDirResolver) {
	if c.Dsn != "" {
		return
	}

	c.Dsn = "./data.db"

	if dataDirResolver != nil {
		s, err := serviceNameFunc()
		if err != nil {
			s = "data"
		}
		c.Dsn = filepath.Join(dataDirResolver.DataDir(), fmt.Sprintf("%s.db", s))
	}

	dir := filepath.Dir(c.Dsn)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		panic(fmt.Errorf("cannot create directory %s: %w", dir, err))
	}
}

// Open opens a database connection based on the provided configuration.
//
// If no database type is specified, it defaults to SQLite.
// For SQLite, if the DSN is empty, it defaults to "./data.db" in the current
// directory, or within the Hostsharing data directory if available.
// For MySQL, a DSN must be provided.
//
// The returned *gorm.DB can be used to execute queries, create migrations,
// or be injected into context via Set() for use in HTTP handlers.
func Open(c Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	if c.Type == "" {
		c.Type = SQLite
	}

	switch c.Type {
	case MySQL:
		if c.Dsn == "" {
			return nil, fmt.Errorf("dsn is required for MySQL")
		}
		dialector = mysql.Open(c.Dsn)
	case SQLite:
		setSQLiteDsnDefault(&c, getDataDirResolver())
		dialector = sqlite.Open(c.Dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", c.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if c.Debug {
		db = db.Debug()
	}

	return db, nil
}

// SetMiddleware returns an HTTP middleware that injects the database connection
// into the request context. This allows handlers to access the database via Get().
//
// Example:
//
//	db, err := database.Open(database.Config{Type: database.SQLite})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	router.Use(database.SetMiddleware(db))
//	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
//	    db := database.Get(r.Context())
//	    // Use db to query users
//	})
func SetMiddleware(tx *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(Set(r.Context(), tx)))
		})
	}
}

type ctxDbKeyType struct{}

var ctxDbKey = ctxDbKeyType{}

// Get retrieves the database connection from the request context.
// It panics if the database connection was not previously set using Set().
//
// Example:
//
//	func handleGetUser(w http.ResponseWriter, r *http.Request) {
//	    db := database.Get(r.Context())
//	    var user User
//	    if err := db.First(&user, id).Error; err != nil {
//	        http.Error(w, err.Error(), http.StatusInternalServerError)
//	        return
//	    }
//	    // Render user response
//	}
func Get(ctx context.Context) *gorm.DB {
	raw, ok := ctx.Value(ctxDbKey).(*gorm.DB)
	if !ok {
		panic("database connection does not exist on context")
	}
	return raw
}

// Set stores the database connection in the context and returns the new context.
// This is typically called by SetMiddleware(), but can be used manually for testing.
//
// Example:
//
//	db, err := database.Open(database.Config{Type: database.SQLite})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	ctx := database.Set(context.Background(), db)
//	// Now Get(ctx) will return the database connection
func Set(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxDbKey, tx)
}
