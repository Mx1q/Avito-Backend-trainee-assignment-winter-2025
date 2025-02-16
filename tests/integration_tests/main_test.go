package integration_tests

import (
	"Avito-Backend-trainee-assignment-winter-2025/tests/postgres_test_helper"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testDbInstance *pgxpool.Pool

func TestMain(m *testing.M) {
	testDB := postgres_test_helper.SetupTestDatabase()
	defer testDB.TearDown()
	testDbInstance = testDB.DbInstance
	os.Exit(m.Run())
}
