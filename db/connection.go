package db

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var db *gorm.DB

func GetDbConnection() *gorm.DB {
	DbHost := os.Getenv("PG_HOST")
	DbUser := os.Getenv("PG_USER")
	DbPassword := os.Getenv("PG_PASSWORD")
	DbName := os.Getenv("PG_DB_NAME")
	DbPort := os.Getenv("PG_PORT")
	DbSslmode := os.Getenv("PG_SSLMODE")

	// Fusioncat leverages GORM migration engine
	// but some changes require manual migration (e.g. adding plugins or sequences).
	// In order to address this, we use golang-migrate library which allows to
	// also run manual migrations along with GORM migrations.
	// This approach may be not optimal because having two migration engines
	// is not the best practice, but it works and allows to have things working properly.
	// Manual migrations have to be applied before GORM migrations,
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		DbUser, DbPassword, DbHost, DbPort, DbName, DbSslmode)
	m, err := migrate.New(os.Getenv("DB_MANUAL_MIGRATIONS_FOLDER"), databaseUrl)
	if err != nil {
		panic("Manual DB migrations error: " + err.Error())
	}

	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			panic("Error while running manual migrations: " + err.Error())
		}
		log.Info("No manual SQL migrations to run")
	}

	// Specify database connection strings
	databaseDsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		DbHost, DbUser, DbPassword, DbName, DbPort, DbSslmode)

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseDsn), &gorm.Config{})
	if err != nil {
		panic("DB connection error: " + err.Error())
	} else {
		log.Info("DB connection established successfully")
	}

	// Migrate ORM related schemas
	err = db.AutoMigrate(
		&UsersDBModel{},
		&ProjectsDBModel{},
		&SchemaVersionsDBModel{},
		&SchemasDBModel{},
		&MessagesDBModel{},
		&AppsDBModel{},
		&ServersDBModel{},
		&ResourcesDBModel{},
		&ResourceBindingsDBModel{},
	)
	if err != nil {
		panic("DB GORM migration error" + err.Error())
	}

	return db
}

func GetDB() *gorm.DB {
	if db == nil {
		db = GetDbConnection()
	}
	return db
}
