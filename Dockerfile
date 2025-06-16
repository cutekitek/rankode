######### Stage 1: Build React (Vite) frontend #########
FROM node:22-alpine AS frontend-build

# Create app directory
WORKDIR /app/frontend

# Copy only package.json & yarn.lock so we can leverage caching
COPY frontend/package.json frontend/yarn.lock ./

# Install dependencies with Yarn (use --frozen-lockfile to respect lockfile)
RUN yarn install --frozen-lockfile

# Copy all frontend source files
COPY frontend/ ./

# Build the Vite app (outputs to /app/frontend /dist)
RUN yarn build



######### Stage 2: Build Go backend (and pull in React build) #########
FROM golang:1.24-alpine AS backend-build


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

# Copy the React build artifacts from the first stage into a folder (e.g. ./static)
# (Optional: only if your Go code is going to serve static files)
COPY --from=frontend-build /app/frontend/dist ./frontend/dist

RUN go build -ldflags="-s -w" -o /app/app main.go

######### Stage 3: Minimal runtime image #########
FROM alpine:3.18

# Create a non-root user (optional but recommended)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy only the Go binary from the previous stage
COPY --from=backend-build /app/app .

# If your Go app requires a configuration or ENV defaults, set them here
# ENV SOME_ENV_VAR=default_value

# Change to non-root user
USER appuser


# Command to run
ENTRYPOINT ["./app"]
