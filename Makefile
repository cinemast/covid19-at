.PHONY: test clean

default: build

build:
	go build -ldflags="-s -w" ./...
	du -h covid19-at

image:
	docker build -t covid19-at .

test:
	GORACE="halt_on_error=1" go test -timeout 5s -race -v -coverprofile="coverage.txt" -covermode=atomic ./...

clean:
	rm -f covid19-at coverage.txt

deploy:
	ssh covid19.spiessknafl.at "cd covid19-at && git pull && docker-compose build && docker-compose up --force-recreate -d"