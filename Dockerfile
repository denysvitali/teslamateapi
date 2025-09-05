# get golang container
FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

# get args
ARG apiVersion=unknown
ARG TARGETOS
ARG TARGETARCH

# create and set workingfolder
WORKDIR /app

# copy go mod files and sourcecode
COPY go.mod go.sum ./
COPY src/ ./src/

# download go mods and compile the program
RUN go mod download && \
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
  -a -installsuffix cgo -ldflags="-w -s \
  -X 'main.apiVersion=${apiVersion}' \
  " -o app ./src/...


# get alpine container
FROM alpine:3.22.1 AS app

# create workdir
WORKDIR /opt/app

# add packages, create nonroot user and group
RUN apk --no-cache add ca-certificates tzdata && \
  addgroup -S nonroot && \
  adduser -S nonroot -G nonroot && \
  chown -R nonroot:nonroot .

# set user to nonroot
USER nonroot:nonroot

# copy binary from builder
COPY --from=builder --chown=nonroot:nonroot --chmod=755 /app/app .

# expose port 8080
EXPOSE 8080

# run application
CMD ["./app"]
