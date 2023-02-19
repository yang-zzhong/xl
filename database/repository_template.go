package database

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

var (
	ErrModelRequired = errors.New("template.Model required")
)

type RepositoryTemplate struct {
	RootDirectory string
	Package       string
	Model         string
}

func (template RepositoryTemplate) Create() error {
	if template.Model == "" {
		return ErrModelRequired
	}
	var dir string
	if err := template.makePackageDirectory(&dir); err != nil {
		return err
	}
	if err := template.makeModelFile(dir); err != nil {
		return err
	}
	if err := template.makeRepositoryFile(dir); err != nil {
		return err
	}
	if err := template.makeRepositoryTestFile(dir); err != nil {
		return err
	}
	return nil
}

func (template RepositoryTemplate) makePackageDirectory(dir *string) error {
	pathfile := template.Package
	if template.RootDirectory != "" {
		pathfile = path.Join(template.RootDirectory, pathfile)
	}
	if template.isExists(pathfile) {
		return fmt.Errorf("directory [%s] has been existed", pathfile)
	}
	if err := os.MkdirAll(pathfile, 0755); err != nil {
		return fmt.Errorf("create directory [%s]: %w", pathfile, err)
	}
	*dir = pathfile
	return nil
}

func (template RepositoryTemplate) makeFile(pathfile, content string) error {
	f, err := os.OpenFile(pathfile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("create file [%s]: %w", pathfile, err)
	}
	if _, err := f.Write([]byte(content)); err != nil {
		return fmt.Errorf("write file content [%s]: %w", pathfile, err)
	}
	return nil
}

func (template RepositoryTemplate) isExists(pathfile string) bool {
	if _, err := os.Open(pathfile); err == nil || !os.IsNotExist(err) {
		return true
	}
	return false
}

func (template RepositoryTemplate) makeModelFile(dir string) error {
	return template.makeFile(path.Join(dir, "model.go"), template.modelContent())
}

func (template RepositoryTemplate) makeRepositoryFile(dir string) error {
	return template.makeFile(path.Join(dir, "repository.go"), template.repositoryContent())
}

func (template RepositoryTemplate) makeRepositoryTestFile(dir string) error {
	return template.makeFile(path.Join(dir, "repository_test.go"), template.repositoryTestContent())
}

func (template RepositoryTemplate) modelContent() string {
	ctt := `
package __package_name__

import (
	"context"
	"time"
	"gorm.io/gorm"
	"udious.com/mockingbird/pkg/database"
	"github.com/google/uuid"
)

// __model__
type __model__ struct {
	ID        string 		__sql_quote__gorm:"primarykey"__sql_quote__
	CreatedAt time.Time		__sql_quote__gorm:"<-:create,autoCreateTime"__sql_quote__
	UpdatedAt time.Time		__sql_quote__gorm:"autoUpdateTime"__sql_quote__
	DeletedAt gorm.DeletedAt __sql_quote__gorm:"index"__sql_quote__
}

func (m *__model__) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

// ID match id
func ID(id string) database.MatchOption {
	return func(opts *database.MatchOptions) { opts.EQ("id", id).SetLimit(1) }
}

func OR(subopts ...database.MatchOption) database.MatchOption {
	return func(opts *database.MatchOptions) {
		sub := database.MatchOptions{}
		opts.OR(sub.Apply(subopts...))
	}
}

func AND(subopts ...database.MatchOption) database.MatchOption {
	return func(opts *database.MatchOptions) {
		sub := database.MatchOptions{}
		opts.AND(sub.Apply(subopts...))
	}
}

// Sort
func Sort(cols ...string) database.MatchOption {
	return func(opts *database.MatchOptions) { opts.Sort = cols }
}

// Offset this is the fetch option
func Offset(offset int) database.MatchOption {
	return func(opts *database.MatchOptions) { opts.Offset = &offset }
}

// Limit this is the fetch option
func Limit(limit int) database.MatchOption {
	return func(opts *database.MatchOptions) { opts.Limit = &limit }
}

// __model__Repository
type Repository interface {
	// Create __model__
	Create(ctx context.Context, m *__model__) error
	// Update __model__
	Update(ctx context.Context, m *__model__) error
	// Fetch __model__[s]
	Fetch(ctx context.Context, result *[]__model__, opts ...database.MatchOption) error
	// First
	First(ctx context.Context, result *__model__, opts ...database.MatchOption) error
	// Total of __model__
	Count(ctx context.Context, result *int64, opts ...database.MatchOption) error
	// Delete __model__
	Delete(ctx context.Context, opts ...database.MatchOption) error
}
	
	`
	ctt = strings.ReplaceAll(ctt, "__sql_quote__", "`")
	ctt = strings.ReplaceAll(ctt, "__package_name__", template.packageName())
	return strings.ReplaceAll(ctt, "__model__", template.Model)
}

func (template RepositoryTemplate) repositoryContent() string {
	ctt := `
package __package_name__

import (
	"context"
	"udious.com/mockingbird/pkg/database"
)

// repository
type repository struct {
	database.Repository
}

// NewRepository
func NewRepository(db database.Repository) Repository {
	return &repository{Repository: db}
}

// Create
func (repo *repository) Create(ctx context.Context, m *__model__) error {
	return repo.Repository.Create(ctx, m)
}

// Update
func (repo *repository) Update(ctx context.Context, m *__model__) error {
	return repo.Repository.Update(ctx, m)
}

// Fetch
func (repo *repository) Fetch(ctx context.Context, result *[]__model__, opts ...database.MatchOption) error {
	return repo.Repository.Find(ctx, result, opts...)
}

// Count
func (repo *repository) Count(ctx context.Context, result *int64, opts ...database.MatchOption) error {
	return repo.Repository.Count(ctx, &__model__{}, result, opts...)
}

func (repo *repository) First(ctx context.Context, result *__model__, opts ...database.MatchOption) error {
	return repo.Repository.First(ctx, result, opts...)
}

// Delete
func (repo *repository) Delete(ctx context.Context, opts ...database.MatchOption) error {
	return repo.Repository.Delete(ctx, &__model__{}, opts...)
}

	`
	ctt = strings.ReplaceAll(ctt, "__package_name__", template.packageName())
	ctt = strings.ReplaceAll(ctt, "__sql_quote__", "`")
	return strings.ReplaceAll(ctt, "__model__", template.Model)
}

func (template RepositoryTemplate) repositoryTestContent() string {
	ctt := `
package __package_name__

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"fmt"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"udious.com/mockingbird/pkg/database"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Any struct{}

func (a Any) Match(v driver.Value) bool {
	return true
}

type AnyNullTime struct{}

func (a AnyNullTime) Match(v driver.Value) bool {
	_, ok := v.(sql.NullTime)
	return ok
}

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func dialector(db *sql.DB) gorm.Dialector {
	return mysql.New(mysql.Config{
		Conn:                      db,
		DriverName:                "mysql",
		SkipInitializeWithVersion: true,
	})
}

func TestGormRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New() // mock sql.DB
	assert.Nil(t, err)
	defer db.Close()
	defer assert.Nil(t, mock.ExpectationsWereMet())
	gdb, err := gorm.Open(dialector(db)) // open gorm db
	assert.Nil(t, err)
	repo := NewRepository(database.NewGormRepository(gdb))
	func() {
		// init your mock here
		tableName := "__table_name__"
		execSql := fmt.Sprintf("^INSERT INTO __sql_quote__%s__sql_quote__ (.*) VALUES (.*)$", tableName)
		mock.ExpectBegin()
		mock.ExpectExec(execSql).
			WithArgs("1", AnyTime{}, AnyTime{}, Any{}).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	}()
	var ctx = context.Background()
	err = repo.Create(ctx, &__model__{
		ID: "1",
	})
	assert.Nil(t, err)
}

func TestGormRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New() // mock sql.DB
	assert.Nil(t, err)
	defer db.Close()
	defer assert.Nil(t, mock.ExpectationsWereMet())
	gdb, err := gorm.Open(dialector(db)) // open gorm db
	assert.Nil(t, err)
	repo := NewRepository(database.NewGormRepository(gdb))
	func() {
		// TODO put your mock here
		tableName := "__table_name__"
		execSql := fmt.Sprintf("^UPDATE __sql_quote__%s__sql_quote__ SET __sql_quote__updated_at__sql_quote__=\\?,__sql_quote__deleted_at__sql_quote__=\\? WHERE __sql_quote__%s__sql_quote__\\.__sql_quote__deleted_at__sql_quote__ IS NULL AND __sql_quote__id__sql_quote__ = \\?$", tableName, tableName)
		mock.ExpectBegin()
		mock.ExpectExec(execSql).
			WithArgs(AnyTime{}, Any{}, "1").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
	}()
	var ctx = context.Background()
	err = repo.Update(ctx, &__model__{
		ID: "1",
	})
	assert.Nil(t, err)
}

func TestGormRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New() // mock sql.DB
	assert.Nil(t, err)
	defer db.Close()
	defer assert.Nil(t, mock.ExpectationsWereMet())
	gdb, err := gorm.Open(dialector(db)) // open gorm db
	assert.Nil(t, err)
	repo := NewRepository(database.NewGormRepository(gdb))
	func() {
		// TODO put your mock here
		tableName := "__table_name__"
		execSql := fmt.Sprintf("^UPDATE __sql_quote__%s__sql_quote__ SET __sql_quote__deleted_at__sql_quote__=\\? WHERE __sql_quote__id__sql_quote__ = \\? AND __sql_quote__%s__sql_quote__\\.__sql_quote__deleted_at__sql_quote__ IS NULL LIMIT 1$", tableName, tableName)
		mock.ExpectBegin()
		mock.ExpectExec(execSql).
			WithArgs(AnyTime{}, "1").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
	}()
	var ctx = context.Background()
	err = repo.Delete(ctx, ID("1"))
	assert.Nil(t, err)
}

func TestGormRepository_Fetch(t *testing.T) {
	db, mock, err := sqlmock.New() // mock sql.DB
	assert.Nil(t, err)
	defer db.Close()
	defer assert.Nil(t, mock.ExpectationsWereMet())
	gdb, err := gorm.Open(dialector(db)) // open gorm db
	assert.Nil(t, err)
	repo := NewRepository(database.NewGormRepository(gdb))
	func() {
		// TODO put your mock here
		tableName := "__table_name__"
		querySql := fmt.Sprintf("^SELECT \\* FROM __sql_quote__%s__sql_quote__ WHERE __sql_quote__id__sql_quote__ = \\? AND __sql_quote__%s__sql_quote__\\.__sql_quote__deleted_at__sql_quote__ IS NULL LIMIT 1$", tableName, tableName)
		mock.ExpectQuery(querySql).
			WithArgs("1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at"}))
	}()
	var ctx = context.Background()
	models := []__model__{}
	err = repo.Fetch(ctx, &models, ID("1"))
	assert.Nil(t, err)
}
`
	ctt = strings.ReplaceAll(ctt, "__package_name__", template.packageName())
	ctt = strings.ReplaceAll(ctt, "__sql_quote__", "`")
	ctt = strings.ReplaceAll(ctt, "__table_name__", strings.ToLower(template.Model)+"s")
	ctt = strings.ReplaceAll(ctt, "__model__", template.Model)

	return ctt
}

func (template RepositoryTemplate) packageName() string {
	ss := strings.Split(template.Package, "/")
	return ss[len(ss)-1]
}
