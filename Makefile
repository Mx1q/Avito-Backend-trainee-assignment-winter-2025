.PHONY: unit_tests integration_tests e2e_tests

tests: unit_tests integration_tests e2e_tests

e2e_tests:
	go test -v -cover -coverpkg "./internal/web" "./tests/e2e_tests"

integration_tests:
	go test -v -cover -coverpkg "./internal/storage/postgres" "./tests/integration_tests"

unit_tests:
	go test -v -cover -coverpkg "./internal/service" "./tests/unit_tests"