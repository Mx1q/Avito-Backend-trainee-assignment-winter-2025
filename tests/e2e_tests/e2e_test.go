package e2e_tests

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/models"
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type E2ESuite struct {
	suite.Suite
	e       httpexpect.Expect
	builder squirrel.StatementBuilderType
}

func (s *E2ESuite) SetupSuite() {
	s.e = *httpexpect.WithConfig(httpexpect.Config{
		Client:   &http.Client{},
		BaseURL:  "http://localhost:8081",
		Reporter: httpexpect.NewAssertReporter(s.T()),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(s.T(), true),
		},
	})
	s.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	started := make(chan bool, 1)
	go RunTheApp(testDbInstance, started)
	<-started
}

func (s *E2ESuite) SetupTest() {
	clearQuery := `truncate table users cascade`
	_, err := testDbInstance.Exec(
		context.Background(),
		clearQuery,
	)
	require.NoError(s.T(), err)

	query, args, err := s.builder.
		Insert("users").
		Columns("username", "password").
		Values("first", "hashedPass").
		Values("second", "hashedPass").
		ToSql()
	require.NoError(s.T(), err)

	_, err = testDbInstance.Exec(
		context.Background(),
		query,
		args...,
	)
}

func (s *E2ESuite) TestE2E_SendCoins() {
	authReq := models.Auth{
		Username: "user",
		Password: "pass",
	}

	r := s.e.POST("/api/auth").
		WithJSON(authReq).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()
	token := r.Value("token").String().Raw()
	require.NotEmpty(s.T(), token)

	reqWithAuth := s.e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", "Bearer "+token)
	})
	sendCoinsReq := models.CoinsTransfer{
		ToUser: "first",
		Amount: 100,
	}

	reqWithAuth.POST("/api/sendCoin").
		WithJSON(sendCoinsReq).
		Expect().
		Status(http.StatusOK)
}

func (s *E2ESuite) TestE2E_BuyItem() {
	authReq := models.Auth{
		Username: "user",
		Password: "pass",
	}

	r := s.e.POST("/api/auth").
		WithJSON(authReq).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()
	token := r.Value("token").String().Raw()
	require.NotEmpty(s.T(), token)

	reqWithAuth := s.e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", "Bearer "+token)
	})

	reqWithAuth.GET("/api/buy/cup").
		Expect().
		Status(http.StatusOK)
}

func (s *E2ESuite) TestE2E_GetInventory() {
	authReq := models.Auth{
		Username: "user",
		Password: "pass",
	}

	r := s.e.POST("/api/auth").
		WithJSON(authReq).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()
	token := r.Value("token").String().Raw()
	require.NotEmpty(s.T(), token)

	reqWithAuth := s.e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", "Bearer "+token)
	})

	reqWithAuth.GET("/api/buy/cup").
		Expect().
		Status(http.StatusOK)
	reqWithAuth.GET("/api/buy/cup").
		Expect().
		Status(http.StatusOK)

	reqWithAuth.GET("/api/buy/powerbank").
		Expect().
		Status(http.StatusOK)

	reqWithAuth.GET("/api/info").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		NotEmpty()
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}
