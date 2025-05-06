# syntax=docker/dockerfile:1

FROM golang:1.24.2 AS build-stage

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY src /src
WORKDIR /src

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o ./flight-price ./cmd

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
EXPOSE 8080

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /src/flight-price flight-price
COPY --from=build-stage /src/cert.* /
COPY src/assets assets

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/flight-price"]