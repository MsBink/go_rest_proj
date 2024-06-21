FROM golang:1.21-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash gcc musl-dev

COPY ["./go.mod", "./go.sum", "./"]
RUN go mod download

COPY package_docker ./
RUN go build -o ./bin/app ./main.go

FROM alpine AS runner

COPY --from=builder usr/local/src/bin/app /
COPY --from=builder usr/local/src/static /static
COPY config.yml /config.yml
CMD ["/app"]