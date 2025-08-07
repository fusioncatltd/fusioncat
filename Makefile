export TESTSERVER_URL ?= http://localhost:8080/

run:
	go run main.go

test:
	go clean -testcache && go test ./tests  -v