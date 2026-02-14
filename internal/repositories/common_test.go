package repositories_test

import (
	"database/sql"
	"math/rand"
	"os"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/database"
	"github.com/pkg/errors"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupDatabase() (*sql.DB, error) {
	databaseName := "testDatabase_" + strconv.Itoa(rand.Int())

	host := getEnvOrDefault("DATABASE_HOST", "localhost")
	port, _ := strconv.Atoi(getEnvOrDefault("DATABASE_PORT", "5432"))
	user := getEnvOrDefault("DATABASE_USER", "postgres")
	password := getEnvOrDefault("DATABASE_PASSWORD", "postgres")

	settings := &database.PostgresDatabaseSettings{
		Host:     host,
		Name:     databaseName,
		Port:     port,
		User:     user,
		Password: password,
	}
	err := settings.WaitForDatabaseToBeOnline(30)
	if err != nil {
		return nil, errors.Wrap(err, "failed waiting for database")
	}

	err = settings.MigrateUp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to migrate database")
	}

	db, err := settings.EnsureDatabaseExistsAndGetConnection()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database")
	}

	return db, nil
}
