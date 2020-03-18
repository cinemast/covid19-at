FROM golang:latest as build
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" ./...

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /go/src/app/metadata.csv .
COPY --from=build /go/src/app/covid19-at .
EXPOSE 8282
CMD ["./covid19-at"]