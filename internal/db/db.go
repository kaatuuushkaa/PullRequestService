package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var db *gorm.DB

func InitDB() (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, name, port)
	var err error

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {

		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	return db, nil
}
