code-coverage:
	go test ./... -coverprofile cover.out && go tool cover -html=cover.out -o coverage.html

test-race:
	go test -race
