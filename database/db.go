package database

import (
	"log"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {

	viper.AutomaticEnv()

	dbUrl := viper.GetString("DATABASE_URL")
	if dbUrl == "" {
		log.Println("missing database url in .env")
	}
	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqldb, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqldb.Ping(); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(Value{}); err != nil {
		return nil, err
	}
	return db, nil
}
