# ----------------------
# Build Stage
# ----------------------
    FROM --platform=$BUILDPLATFORM golang:1.23-bookworm AS builder

    WORKDIR /app
    
    # Build arguments for cross-compilation
    ARG TARGETOS
    ARG TARGETARCH
    
    # Install necessary build tools
    RUN apt-get update && apt-get install -y make
    
    # Copy go.mod and go.sum files and download dependencies
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the rest of the source code
    COPY . .
    
    # Build the application
    RUN \
      CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /app/target/asvec .
    
# ----------------------
# Final Stage
# ----------------------
    FROM --platform=$TARGETPLATFORM gcr.io/distroless/static-debian12
    
    # Copy the binary from the builder stage
    COPY --from=builder /app/target/asvec /usr/local/bin/asvec
    
    # Set the entrypoint to the asvec binary
    ENTRYPOINT ["/usr/local/bin/asvec"]
    