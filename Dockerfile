
FROM golang:1.23 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/task-planner ./cmd/task-planner

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /bin/task-planner ./task-planner
COPY migration ./migration

RUN apk add --no-cache bash
COPY scripts/wait-for-postgres.sh /usr/local/bin/wait-for-postgres.sh
RUN chmod +x /usr/local/bin/wait-for-postgres.sh

EXPOSE 8080
ENTRYPOINT ["./task-planner"]

