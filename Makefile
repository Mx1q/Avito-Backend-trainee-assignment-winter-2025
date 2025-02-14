integration_tests:
	go test -v -cover -coverpkg "./internal/storage/postgres" "./tests/integration_tests"

unit_tests:
	go test -v -cover -coverpkg "./internal/service" "./tests/unit_tests"