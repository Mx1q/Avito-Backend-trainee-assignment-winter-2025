package integration_tests

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
)

var testDbInstance *pgxpool.Pool

func TestMain(m *testing.M) {
	testDB := SetupTestDatabase()
	defer testDB.TearDown()
	testDbInstance = testDB.DbInstance
	os.Exit(m.Run())
}
