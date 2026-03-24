package controllers_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/database"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupDatabase(t *testing.T) (*sql.DB, error) {
	t.Helper()

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

	// Set connection pool limits to prevent exhausting PostgreSQL connections
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	t.Cleanup(func() {
		db.Close()
		// Drop the test database to free tmpfs space
		adminDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
			getEnvOrDefault("DATABASE_HOST", "localhost"),
			getEnvOrDefault("DATABASE_PORT", "5432"),
			getEnvOrDefault("DATABASE_USER", "postgres"),
			getEnvOrDefault("DATABASE_PASSWORD", "postgres"),
		)
		adminDB, err := sql.Open("postgres", adminDSN)
		if err == nil {
			adminDB.Exec("DROP DATABASE IF EXISTS \"" + databaseName + "\"")
			adminDB.Close()
		}
	})

	return db, nil
}

// MockRouter is a mock implementation of the router for testing
type MockRouter struct{}

func (m *MockRouter) RegisterRestAPIRoute(path string, auth web.AuthAccess, handler func(*web.HandlerArgs) (any, *web.HttpError), methods ...string) {
	// Mock implementation - does nothing
}

// MockContactPermissionsRepository is a mock implementation of ContactPermissionsRepository
type MockContactPermissionsRepository struct {
	mock.Mock
}

func (m *MockContactPermissionsRepository) GetByContact(ctx context.Context, contactID int64, userID int64) ([]*models.ContactPermission, error) {
	args := m.Called(ctx, contactID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ContactPermission), args.Error(1)
}

func (m *MockContactPermissionsRepository) Upsert(ctx context.Context, perm *models.ContactPermission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockContactPermissionsRepository) GetUserPermissionsForService(ctx context.Context, viewerUserID int64, serviceType string) ([]int64, error) {
	args := m.Called(ctx, viewerUserID, serviceType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockContactPermissionsRepository) CheckPermission(ctx context.Context, grantingUserID, receivingUserID int64, serviceType string) (bool, error) {
	args := m.Called(ctx, grantingUserID, receivingUserID, serviceType)
	return args.Bool(0), args.Error(1)
}
