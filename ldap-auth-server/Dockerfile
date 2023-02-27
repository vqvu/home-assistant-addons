# # https://developers.home-assistant.io/docs/add-ons/configuration#add-on-dockerfile
ARG BUILD_FROM
FROM $BUILD_FROM AS build

COPY . /build

WORKDIR /build

# Execute during the build of the image
RUN apk add --no-cache go && go build

FROM $BUILD_FROM

COPY --from=build /build/ldap-auth-server /usr/bin

CMD ["/usr/bin/ldap-auth-server", "/data/options.json"]
