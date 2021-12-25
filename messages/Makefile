name = messages

preparation:
	go get github.com/google/uuid
	go get github.com/jackc/pgx/v4
	go get github.com/jackc/pgx/v4/pgxpool
	go mod download golang.org/x/net
	@createdb $(name) 2> /dev/null || true

run: 
	go run $(name).go 

build: preparation
	go build -o $(name) $(name).go

clean: 
	rm $(exec)
	@dropdb $(name) 2> /dev/null || true
