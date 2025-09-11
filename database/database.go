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

type DBType string

const (
	SQLite DBType = "sqlite"
	MySQL  DBType = "mysql"
)

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

func setSQLiteDsnDefault(c *Config, dataDirResolver dataDirResolver) {
	if c.Dsn != "" {
		return
	}

	c.Dsn = "./data.db"

	if dataDirResolver != nil {
		c.Dsn = filepath.Join(dataDirResolver.DataDir(), "data.db")
	}

	dir := filepath.Dir(c.Dsn)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		panic(fmt.Errorf("cannot create directory %s: %w", dir, err))
	}
}

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

func SetMiddleware(tx *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(Set(r.Context(), tx)))
		})
	}
}

type ctxDbKeyType struct{}

var ctxDbKey = ctxDbKeyType{}

func Get(ctx context.Context) *gorm.DB {
	raw, ok := ctx.Value(ctxDbKey).(*gorm.DB)
	if !ok {
		panic("database connection does not exist on context")
	}
	return raw
}

func Set(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxDbKey, tx)
}
