/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/yang-zzhong/xl/database"
	"github.com/yang-zzhong/xl/errs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	dsn string
)

// testDatabaseCmd represents the testDatabase command
var testDatabaseCmd = &cobra.Command{
	Use:   "testDatabase",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: run,
}

func init() {
	rootCmd.AddCommand(testDatabaseCmd)
	testDatabaseCmd.Flags().StringVar(&dsn, "dsn", "root@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", "")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testDatabaseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testDatabaseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func run(cmd *cobra.Command, args []string) {
	fmt.Println("testDatabase called")
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't open mysql: %s\n", err.Error())
		return
	}
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Book{})
	ctx := context.Background()
	test("insert/update/delete/select", db, func() error {
		repo := database.NewGormRepository(db)
		u := User{
			ID:    uuid.NewString(),
			Name:  "author1",
			Email: "email@a.com",
		}
		if err := repo.Create(ctx, u); err != nil {
			return errs.Wrap(err, "create user")
		}
		books := []Book{
			{ID: uuid.NewString(), Name: "book1", AuthorID: u.ID},
			{ID: uuid.NewString(), Name: "book2", AuthorID: u.ID},
			{ID: uuid.NewString(), Name: "book3", AuthorID: u.ID},
		}
		for _, book := range books {
			if err := repo.Create(ctx, book); err != nil {
				return errs.Wrap(err, "create book")
			}
		}
		return nil
	})
}

type User struct {
	ID        string `gorm:"type:varchar(36);primarykey"`
	Name      string
	Email     string
	Phone     string
	CreatedAt time.Time      `gorm:"<-:create,autoCreateTime"`
	UpdateAt  time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Book struct {
	ID        string `gorm:"type:varchar(36);primarykey"`
	Name      string
	AuthorID  string
	Phone     string
	CreatedAt time.Time      `gorm:"<-:create,autoCreateTime"`
	UpdateAt  time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func test(name string, db *gorm.DB, test func() error) {
	db.Migrator().AutoMigrate(&User{})
	db.Migrator().AutoMigrate(&Book{})
	defer func() {
		db.Migrator().DropTable(&User{})
		db.Migrator().DropTable(&Book{})
	}()
	err := test()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s [failed]", name, err)
		return
	}
	fmt.Fprintf(os.Stdout, "%s: [ok]", name)
}
