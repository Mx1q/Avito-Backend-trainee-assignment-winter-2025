package e2e_tests

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/models"
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type E2ESuite struct {
	suite.Suite
	e       httpexpect.Expect
	builder squirrel.StatementBuilderType
}

const (
	userCoinsOnRegister = 1000
	item1ToBuy          = "hoody"
	item1ToBuyCost      = 300
	item2ToBuy          = "cup"
	item2ToBuyCost      = 10
)

func (s *E2ESuite) SetupSuite() {
	s.e = *httpexpect.WithConfig(httpexpect.Config{
		Client:   &http.Client{},
		BaseURL:  fmt.Sprintf("http://localhost:%d", TestingPort),
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
	require.NoError(s.T(), err)
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

	reqWithAuth.GET(fmt.Sprintf("/api/buy/%s", item2ToBuy)).
		Expect().
		Status(http.StatusOK)
}

func (s *E2ESuite) TestE2E_BuyItem_NotEnoughCoins() {
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

	for userCoins := userCoinsOnRegister; userCoins > item1ToBuyCost; userCoins -= item1ToBuyCost {
		reqWithAuth.GET(fmt.Sprintf("/api/buy/%s", item1ToBuy)).
			Expect().
			Status(http.StatusOK)
	}

	reqWithAuth.GET(fmt.Sprintf("/api/buy/%s", item1ToBuy)).
		Expect().
		Status(http.StatusBadRequest)
}

func (s *E2ESuite) TestE2E_InvalidToken() {
	token := "invalidToken"
	reqWithAuth := s.e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", "Bearer "+token)
	})

	reqWithAuth.GET(fmt.Sprintf("/api/buy/%s", item1ToBuy)).
		Expect().
		Status(http.StatusUnauthorized)

	reqWithAuth.GET("/api/info").
		Expect().
		Status(http.StatusUnauthorized)

	sendCoinsReq := models.CoinsTransfer{
		ToUser: "first",
		Amount: 100,
	}
	reqWithAuth.POST("/api/sendCoin").
		WithJSON(sendCoinsReq).
		Expect().
		Status(http.StatusUnauthorized)
}

func (s *E2ESuite) TestE2E_InvalidRequests() {
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

	sendCoinsReq := models.CoinsTransfer{ // empty toUser
		ToUser: "",
		Amount: 100,
	}
	reqWithAuth.POST("/api/sendCoin").
		WithJSON(sendCoinsReq).
		Expect().
		Status(http.StatusBadRequest)

	reqWithAuth.GET("/api/buy/undefinedItem").
		Expect().
		Status(http.StatusBadRequest)

	authReq.Password = "invalidPass"
	s.e.POST("/api/auth").
		WithJSON(authReq).
		Expect().
		Status(http.StatusUnauthorized)

	authReq.Username = ""
	s.e.POST("/api/auth").
		WithJSON(authReq).
		Expect().
		Status(http.StatusBadRequest)
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

	userCoins := userCoinsOnRegister
	const item2ToBuyCount = 2
	const item1ToBuyCount = 1

	for i := 0; i < item2ToBuyCount && userCoins > item2ToBuyCost; i++ {
		reqWithAuth.GET(fmt.Sprintf("/api/buy/%s", item2ToBuy)).
			Expect().
			Status(http.StatusOK)
		userCoins -= item2ToBuyCost
	}

	for i := 0; i < item1ToBuyCount && userCoins > item1ToBuyCost; i++ {
		reqWithAuth.GET(fmt.Sprintf("/api/buy/%s", item1ToBuy)).
			Expect().
			Status(http.StatusOK)
		userCoins -= item1ToBuyCost
	}

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
