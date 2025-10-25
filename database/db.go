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
		log.Fatalf("error connecting to db: %v", err.Error())
	}
	sqldb, _ := db.DB()
	if err := sqldb.Ping(); err != nil {
		log.Fatalf("error pinging the db: %v", err.Error())
	}
	if err := db.AutoMigrate(Value{}); err != nil {
		log.Fatalf("error pinging the db: %v", err.Error())
	}
	return db, nil
}
