FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /pr-reviewer ./cmd/server

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /pr-reviewer /pr-reviewer

EXPOSE 8080
CMD ["/pr-reviewer"]