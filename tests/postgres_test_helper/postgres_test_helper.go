package postgres_test_helper

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/config"
	"Avito-Backend-trainee-assignment-winter-2025/internal/storage/postgres"
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var postgresConfig = &config.PostgresConfig{
	Driver:   "postgres",
	Host:     "localhost",
	Port:     5432,
	User:     "user_test",
	Password: "pass_test",
	DBName:   "shop_test",
}

const (
	Image = "postgres:13-alpine"
)

type TestDatabase struct {
	DbInstance *pgxpool.Pool
	DbAddress  string
	container  testcontainers.Container
}

func SetupTestDatabase() *TestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	container, dbInstance, dbAddr, err := createContainer(ctx)
	if err != nil {
		log.Fatal("failed to setup test: ", err)
	}

	err = migrateDb(dbAddr)
	if err != nil {
		log.Fatal("failed to perform db migration: ", err)
	}
	cancel()

	return &TestDatabase{
		container:  container,
		DbInstance: dbInstance,
		DbAddress:  dbAddr,
	}
}

func (tdb *TestDatabase) TearDown() {
	tdb.DbInstance.Close()
	err := tdb.container.Terminate(context.Background())
	if err != nil {
		log.Fatal("failed to tear down test database: ", err)
	}
}

func createContainer(ctx context.Context) (testcontainers.Container, *pgxpool.Pool, string, error) {
	env := map[string]string{
		"POSTGRES_PASSWORD": postgresConfig.Password,
		"POSTGRES_USER":     postgresConfig.User,
		"POSTGRES_DB":       postgresConfig.DBName,
	}

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        Image,
			ExposedPorts: []string{fmt.Sprintf("%d/tcp", postgresConfig.Port)},
			Env:          env,
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to start container: %v", err)
	}

	p, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to get container external port: %v", err)
	}

	log.Println("postgres container ready and running at port: ", p.Port())
	time.Sleep(time.Second)

	postgresConfig.Port, err = strconv.Atoi(p.Port())
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to parse port string: %v", err)
	}
	db, err := postgres.NewConn(context.Background(), postgresConfig)
	if err != nil {
		return container, db,
			fmt.Sprintf("%s:%d", postgresConfig.Host, postgresConfig.Port),
			fmt.Errorf("failed to establish database connection: %v", err)
	}

	return container, db,
		fmt.Sprintf("%s:%d", postgresConfig.Host, postgresConfig.Port), nil
}

func migrateDb(dbAddr string) error {
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get path")
	}
	_ = path
	migrationFilesPath, err := filepath.Glob(filepath.Join(filepath.Dir(path), "..", "..", "migrations"))
	if err != nil {
		return err
	}
	if len(migrationFilesPath) == 0 {
		return fmt.Errorf("migration file not found")
	}
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		postgresConfig.User, postgresConfig.Password, dbAddr, postgresConfig.DBName)

	m, err := migrate.New(fmt.Sprintf("file:%s", migrationFilesPath[0]), databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
