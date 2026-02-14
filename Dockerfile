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

RUN apt update && apt install -y ca-certificates
WORKDIR /app

COPY --from=0 /build/industry-tool /app/

CMD ["/app/industry-tool"]

FROM node:24.9.0-slim as ui-deps

WORKDIR /app

COPY frontend/package.json frontend/yarn.lock ./
COPY frontend/packages/ ./packages

RUN yarn install --frozen-lockfile

FROM node:24.9.0-slim as ui-build

WORKDIR /app

COPY frontend .
COPY --from=ui-deps /app/node_modules ./node_modules

ENV API_SERVICE=http://localhost:8080/api/

RUN yarn build && yarn install --production --ignore-scripts --prefer-offline

FROM node:24.9.0-slim as publish-ui
WORKDIR /app
ENV NODE_ENV production

COPY --from=ui-build /app/next.config.ts ./
COPY --from=ui-build /app/public ./public
COPY --from=ui-build /app/.next ./.next
COPY --from=ui-build /app/node_modules ./node_modules
COPY --from=ui-build /app/package.json ./package.json

EXPOSE 3000

CMD ["yarn", "start"]
