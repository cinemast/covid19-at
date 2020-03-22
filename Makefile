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

report:
	 zcat -f /var/log/nginx/* | grep "GET /d/\|GET /api/datasources/proxy/50/api/v1/query?query=cov19_dead&time=\|/prometheus/api/v1/query\|/covid19/metrics" | goaccess --log-format=COMBINED -q -a -o /home/cinemast/report/index.html --ignore-crawlers

deploy:
	ssh covid19.spiessknafl.at "cd covid19-at && git pull && docker-compose build && docker-compose up --force-recreate -d"