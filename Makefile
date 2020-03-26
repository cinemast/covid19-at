.PHONY: test clean

default: build sync-logs

build:
	go build -ldflags="-s -w" ./...
	du -h covid19-at

image:
	docker build -t covid19-at .

test:
	GORACE="halt_on_error=1" go test -timeout 5s -race -v -coverprofile="coverage.txt" -covermode=atomic ./...

clean:
	rm -f covid19-at coverage.txt data/report*

sync-logs:
	mkdir -p data/nginx
	rsync -avz covid19.spiessknafl.at:/var/log/nginx/ data/nginx/

report-exporter: sync-logs
	gzcat -f data/nginx/* | grep "GET /covid19/metrics\|GET /api" | LANG="en_US.UTF-8" goaccess --log-format=COMBINED -q -a -o data/report-exporter.html --ignore-crawlers
	open data/report-exporter.html

report-prometheus: sync-logs
	gzcat -f data/nginx/* | grep "/prometheus/api/v1/query\|/api/datasources/proxy/" | LANG="en_US.UTF-8" goaccess --log-format=COMBINED -q -a -o data/report-prometheus.html --ignore-crawlers
	open data/report-prometheus.html

report-grafana: sync-logs
	gzcat -f data/nginx/* | grep "GET /d/\|/impressum.html" | grep -v "/public/\|favicon.ico\|/images/" | LANG="en_US.UTF-8" goaccess --log-format=COMBINED -q -a -o data/report-grafana.html --ignore-crawlers
	open data/report-grafana.html

reports: report-exporter report-grafana report-prometheus

deploy:
	ssh covid19.spiessknafl.at "cd covid19-at && git pull && docker-compose build && docker-compose up --force-recreate -d"