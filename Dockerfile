ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .

FROM debian:bookworm

# Set the working directory where your binary expects to find static assets.
WORKDIR /app

# Copy the binary into the final image.
COPY --from=builder /run-app /app/run-app

# Copy the static and templates directories.
COPY --from=builder /usr/src/app/static /app/static
COPY --from=builder /usr/src/app/templates /app/templates

# Run the binary.
CMD ["/app/run-app"]

