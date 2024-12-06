# Use a minimal base image
FROM alpine:3.21@sha256:21dc6063fd678b478f57c0e13f47560d0ea4eeba26dfc947b2a4f81f686b9f45

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
