FROM --platform=${BUILDPLATFORM} golang:alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY . .
ARG TARGETOS
ARG TARGETARCH

# Set default version, this would be overwrite from GitHub Action
ARG CLOUD_CONNECTOR_VERSION=v0.0

# Build flags
ARG LDFLAGS="-ldflags=-w -s"
ARG OTHERFLAGS="-trimpath -mod=readonly"
ARG VERSION="-X 'main.version=${CLOUD_CONNECTOR_VERSION}'"
RUN echo ${VERSION}

# Build
WORKDIR /src/cmd
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/cloudconnector "${LDFLAGS} ${VERSION}" ${OTHERFLAGS} .

# Set executable flag
RUN chmod +x /src/resources/*.sh

# Final container
FROM busybox AS bin

# Copy binary
COPY --from=build /out/cloudconnector /app/

# Copy default protobuf messages and CA cert
COPY --from=build /src/resources/. /app/

# Create /config dir to be used for custom configuration
RUN mkdir /config

# Execute start script to support ENV variables
WORKDIR /app
CMD ["./cloudconnector_container_start.sh"]
