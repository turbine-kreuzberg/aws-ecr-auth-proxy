# Use a minimal base image
FROM alpine:3.20@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d

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
