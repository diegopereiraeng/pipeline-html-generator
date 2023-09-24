# FROM golang:1.21-alpine as build
# # Install git for fetching dependencies
# RUN apk add --no-cache --update git

# # Set working directory inside the container
# WORKDIR /app

# # Copy source code files
# COPY *.go ./
# COPY internal ./internal
# # Copy go mod and sum files
# COPY go.mod go.sum ./

# # Fetch dependencies
# RUN go mod download

# # Build the application
# RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o pipeline-html-generator 

# Final stage
FROM alpine:3.14

# Install required packages
RUN apk --no-cache --update add curl unzip git

# Copy the built binary from the build stage
# COPY --from=build /app/pipeline-html-generator /bin/
COPY binary/pipeline-html-generator /bin/

# Set working directory
WORKDIR /bin

# Command to run
ENTRYPOINT ["/bin/pipeline-html-generator"]

