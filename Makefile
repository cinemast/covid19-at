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
	rm -f covid19-at coverage.txt data/report*

report:
	mkdir -p data/nginx
	rsync -avz covid19.spiessknafl.at:/var/log/nginx/ data/nginx/
	gzcat -f data/nginx/* | grep "GET /covid19/metrics" | LANG="en_US.UTF-8" goaccess --log-format=COMBINED -q -a -o data/report-exporter.html --ignore-crawlers
	gzcat -f data/nginx/* | grep "/prometheus/api/v1/query\|/api/datasources/proxy/" | LANG="en_US.UTF-8" goaccess --log-format=COMBINED -q -a -o data/report-prometheus.html --ignore-crawlers
	gzcat -f data/nginx/* | grep "GET /d/" | grep -v "/public/\|favicon.ico\|/images/" | LANG="en_US.UTF-8" goaccess --log-format=COMBINED -q -a -o data/report-grafana.html --ignore-crawlers
	open data/report*
deploy:
	ssh covid19.spiessknafl.at "cd covid19-at && git pull && docker-compose build && docker-compose up --force-recreate -d"