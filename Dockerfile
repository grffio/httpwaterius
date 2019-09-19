FROM golang:1.13-alpine AS build
ARG VERSION
WORKDIR /httpwaterius
RUN adduser -D -g '' httpwaterius
RUN apk --update add ca-certificates
COPY go.mod .
COPY go.sum .
COPY cmd cmd
COPY internal internal
COPY template template
RUN env CGO_ENABLED=0 go install -ldflags="-w -s -X main.version=${VERSION}" ./...
FROM scratch
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/bin/httpwaterius /usr/bin/httpwaterius
COPY --from=build --chown=1000:1000 /httpwaterius/template/index.html /opt/index.html
USER httpwaterius
ENTRYPOINT ["/usr/bin/httpwaterius"]