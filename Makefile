build:
	go build  -o bin/server main.go

build_linux:
	GOOS=linux GOARCH=amd64 go build -o bin/server_linux  main.go

run:
	go run main.go -m $(MODE)

test:
	go test ./...

build-push: build_linux
	gcloud compute scp --zone=us-east1-b --compress ./bin/server_linux didil@protohackers-1:~