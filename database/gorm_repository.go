package database

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/yang-zzhong/structs"
	"github.com/yang-zzhong/xl/utils"

	"gorm.io/gorm"
)

func (f Field) DESC() string {
	return "`" + string(f) + "` DESC"
}

func (f Field) ASC() string {
	return "`" + string(f) + "` ASC"
}

var operatorMap = map[Operator]string{
	EQ:  "=",
	NEQ: "!=",
	LT:  "<",
	LTE: "<=",
	GT:  ">",
	GTE: ">=",
	IN:  "IN",
}

type tableNamer interface {
	TableName() string
}

type GormStack interface {
	Push(db *gorm.DB)
	Pop()
}

type gormRepository struct {
	db []*gorm.DB
}

// NewGormRepository
// usage:
//   repo := NewGormRepository(db)
//   // find
//   var users []User
//   ctx := context.Background()
//   err := repo.Find(ctx, &users, Role("member"), Limit(20))
//   // join
//   var books []struct {
//       ID 		string `field:"books.id"`
//       Name 		string `field:"books.name"`
//       AuthorID 	string `field:"users.id"`
//       AuthorName string `field:"users.name"`
// 	 }{}
//   func AuthorID(authorID interface{}) MatchOption {
//   	return func(opts *MatchOptions) {
//   		vo := reflect.ValueOf(authorID)
//   		if vo.Kind() == reflect.Slice {
//   			opts.IN("book.author_id", authorID)
//   			return
//   		}
//   		opts.EQ("book.author_id", authorID)
//   	}
//   }
//   err := repo.Find(ctx, database.M(&books, &Book{}).With(&User{}, AuthorID(Field("users.id"))), Limit(20))
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: []*gorm.DB{db}}
}

func (db *gormRepository) recentDB() *gorm.DB {
	return db.db[len(db.db)-1]
}

func (db *gormRepository) First(ctx context.Context, v interface{}, opts ...MatchOption) error {
	selector, result := db.model(v)
	db.applyOptions(selector, opts...)
	return selector.First(result).Error
}

func (db *gormRepository) Find(ctx context.Context, v interface{}, opts ...MatchOption) error {
	selector, result := db.model(v)
	db.applyOptions(selector, opts...)
	return selector.Find(result).Error
}

// db.Count(ctx, database.M(result, &User{}))
func (db *gormRepository) Count(ctx context.Context, v interface{}, result *int64, opts ...MatchOption) error {
	selector, _ := db.model(v)
	db.applyOptions(selector, opts...)
	return selector.Count(result).Error
}

func (db *gormRepository) Update(ctx context.Context, v interface{}) error {
	return db.recentDB().Save(v).Error
}

func (db *gormRepository) Delete(ctx context.Context, v interface{}, opts ...MatchOption) error {
	deletor := db.recentDB().Model(v)
	db.applyOptions(deletor, opts...)
	return deletor.Delete(v).Error
}

func (db *gormRepository) Create(ctx context.Context, v interface{}) error {
	return db.recentDB().Create(v).Error
}

func (db *gormRepository) UpdateFields(ctx context.Context, v interface{}, fields Fields, opts ...MatchOption) error {
	updator := db.recentDB().Model(v)
	db.applyOptions(updator, opts...)
	return updator.UpdateColumns(fields).Error
}

func (repo *gormRepository) tableName(v interface{}) string {
	if t, ok := v.(tableNamer); ok {
		return t.TableName()
	}
	ptrv := reflect.ValueOf(v)
	if ptrv.Kind() == reflect.Ptr {
		ptrv = ptrv.Elem()
	}
	return repo.recentDB().NamingStrategy.TableName(ptrv.Type().Name())
}

func (repo *gormRepository) Push(db *gorm.DB) {
	repo.db = append(repo.db, db)
}

func (repo *gormRepository) Pop() {
	if len(repo.db) <= 1 {
		return
	}
	repo.db = repo.db[:len(repo.db)-1]
}

func (db *gormRepository) Transaction(do func() error) {
	db.recentDB().Transaction(func(tx *gorm.DB) error {
		db.Push(tx)
		defer db.Pop()
		return do()
	})
}

func (repo *gormRepository) compileMatchOptions(opts MatchOptions) (string, []interface{}) {
	ret := ""
	values := []interface{}{}
	for i, opt := range opts.Matches {
		switch opt.Operator {
		case NULL:
			if i > 0 {
				ret += " AND "
			}
			ret += fmt.Sprintf("`%s` IS NULL", opt.Field)
		case NOTNULL:
			if i > 0 {
				ret += " AND "
			}
			ret += fmt.Sprintf("`%s` IS NULL", opt.Field)
		case OR:
			str, subValues := repo.compileMatchOptions(opt.Value.(MatchOptions))
			ret += fmt.Sprintf(" OR (%s)", str)
			values = append(values, subValues...)
		case AND:
			str, subValues := repo.compileMatchOptions(opt.Value.(MatchOptions))
			ret += fmt.Sprintf(" AND (%s)", str)
			values = append(values, subValues...)
		default:
			if i > 0 {
				ret += " AND "
			}
			if field, ok := opt.Value.(Field); ok {
				ret += fmt.Sprintf("%s %s %s", opt.Field, operatorMap[opt.Operator], field)
				continue
			}
			values = append(values, opt.Value)
			ret += fmt.Sprintf("%s %s ?", opt.Field, operatorMap[opt.Operator])
		}
	}
	return ret, values
}

func (repo *gormRepository) model(v interface{}) (*gorm.DB, interface{}) {
	m, ok := v.(*Model)
	if !ok {
		return repo.recentDB().Model(v), v
	}
	model := repo.recentDB().Model(m.From)
	for _, join := range m.Joins {
		str := ""
		switch join.Type {
		case LeftJoin:
			str += "LEFT JOIN "
		case RightJoin:
			str += "RIGHT JOIN "
		case InnerJoin:
			str += "Inner JOIN "
		}
		str += repo.tableName(join.Model) + " ON "
		condi, values := repo.compileMatchOptions(join.Opts)
		str += condi
		model.Joins(str, values...)
	}
	if m.Grp != nil {
		model.Group(m.Grp.By)
		if m.Grp.Having != nil {
			condi, values := repo.compileMatchOptions(*m.Grp.Having)
			model.Having(condi, values...)
		}
	}
	if m.Result == nil {
		return model, nil
	}
	vm := reflect.ValueOf(m.Result)
	for vm.Kind() == reflect.Slice || vm.Kind() == reflect.Array || vm.Kind() == reflect.Ptr {
		vm = vm.Elem()
	}
	if vm.Kind() != reflect.Struct {
		return model, m.Result
	}
	fields := structs.Fields(m.Result)
	fieldNames := make([]string, len(fields))
	for j, f := range fields {
		fieldNames[j] = fmt.Sprintf("%s AS %s", f.Tag("field"), utils.ToSnakeCase(f.Name()))
	}
	model.Select(fieldNames)

	return model, m.Result
}

func (repo *gormRepository) applyOptions(db *gorm.DB, opts ...MatchOption) {
	opt := &MatchOptions{}
	opt.Apply(opts...)
	for _, match := range opt.Matches {
		switch match.Operator {
		case NULL:
			db.Where(fmt.Sprintf("%s IS NULL", match.Field))
		case NOTNULL:
			db.Where(fmt.Sprintf("%s IS NOT NULL", match.Field))
		case OR:
			str, values := repo.compileMatchOptions(*match.Value.(*MatchOptions))
			db.Where(fmt.Sprintf("OR (%s)", str), values...)
		case AND:
			str, values := repo.compileMatchOptions(*match.Value.(*MatchOptions))
			db.Where(fmt.Sprintf("AND (%s)", str), values...)
		default:
			db.Where(fmt.Sprintf("%s %s ?", match.Field, operatorMap[match.Operator]), match.Value)
		}
	}
	if len(opt.Sort) > 0 {
		db.Order(strings.Join(opt.Sort, ","))
	}
	if opt.Limit != nil {
		db.Limit(*opt.Limit)
	}
	if opt.Offset != nil {
		db.Offset(*opt.Offset)
	}
}
