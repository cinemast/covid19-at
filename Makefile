
build-arm64:
	GOOS=linux GOARCH=arm64 go build

image:
	docker build -t covid19-at .

deploy: build-arm64
	scp covid19.service root@on2:/etc/systemd/system/covid19.service
	scp covid19-at on2:cov19/covid19-at