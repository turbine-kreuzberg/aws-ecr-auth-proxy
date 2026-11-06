# Use a minimal base image
FROM alpine:3.21@sha256:87a4361f6c819618a6bafa2ed5228714cb0bb22195d2b49c95eb26799990128a

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Set the working directory
WORKDIR /app

# Copy the pre-built binary into the container
ARG ARCH=amd64
COPY aws-ecr-auth-proxy-${ARCH} /app/aws-ecr-auth-proxy

# Make the binary executable
RUN chmod +x /app/aws-ecr-auth-proxy

# Expose the port the app runs on
EXPOSE 8080

# Run the binary
CMD ["/app/aws-ecr-auth-proxy"]
