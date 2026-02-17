FROM golang:1.25 as build

WORKDIR /build

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go vet ./...

RUN go build -o ./industry-tool ./cmd/industry-tool

#testing
FROM build as test

RUN go get github.com/golang/mock/mockgen@latest

CMD go test -race -coverprofile=/artifacts/coverage.txt -covermode=atomic -p 1 ./...

# final image
FROM ubuntu:26.04 as final-backend

RUN apt update && apt install -y ca-certificates curl
WORKDIR /app

COPY --from=0 /build/industry-tool /app/

CMD ["/app/industry-tool"]
