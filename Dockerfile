FROM golang:alpine as builder
ENV POSTGRES_DB=chat \
    POSTGRES_HOST=postgres \
    POSTGRES_USER=postgres \
    POSTGRES_PASSWORD=5432

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o bin main.go

FROM scratch
WORKDIR /app
COPY --from=builder /app/bin .
EXPOSE 5000
CMD ["/app/bin"]