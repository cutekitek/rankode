FROM golang:1.25-alpine AS backend-build


# Set Go env for module mode
ENV CGO_ENABLED=0

# Create working directory for backend
WORKDIR /app/backend

# Copy go.mod and go.sum first (to take advantage of cache if modules haven’t changed)
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the rest of the backend source
COPY ./internal ./internal
COPY ./main.go ./main.go

RUN go build -ldflags="-s -w" -o /app/app main.go

######### Stage 3: Minimal runtime image #########
FROM alpine:latest

# Create a non-root user (optional but recommended)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy only the Go binary from the previous stage
COPY --from=backend-build /app/app .

# Change to non-root user
USER appuser


# Command to run
ENTRYPOINT ["./app"]
