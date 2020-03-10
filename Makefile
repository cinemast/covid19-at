.PHONY: test
build-arm64:
	GOOS=linux GOARCH=arm64 go build

image:
	docker build -t covid19-at .

test:
	go test ./...

deploy:
	ssh covid19.spiessknafl.at "cd covid19-at && git pull && docker-compose build && docker-compose up --force-recreate -d"