# # https://developers.home-assistant.io/docs/add-ons/configuration#add-on-dockerfile
ARG BUILD_FROM
FROM golang:1.24.0-alpine AS builder

COPY . /build

WORKDIR /build

# Execute during the build of the image
RUN go build

FROM $BUILD_FROM

COPY --from=builder /build/ldap-auth-server /usr/bin

CMD ["/usr/bin/ldap-auth-server", "/data/options.json"]
