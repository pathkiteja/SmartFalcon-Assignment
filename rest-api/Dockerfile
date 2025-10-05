# Build stage
FROM golang:1.24 as builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o rest-api .

# Final image
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /workspace/rest-api /app/rest-api

# copy connection artifacts (you can mount these at runtime instead)
# For security, it's recommended to mount the test-network org credentials at runtime, not bake them in image.
# COPY organizations /app/organizations

EXPOSE 8080
ENTRYPOINT ["/app/rest-api"]