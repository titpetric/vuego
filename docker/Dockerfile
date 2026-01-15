FROM alpine:latest

ARG TARGETARCH

RUN apk update && apk add --no-cache coreutils
RUN which uname
RUN id nobody 2>/dev/null || adduser -D -H -u 65534 nobody

COPY ./bin/vuego-linux-${TARGETARCH} /usr/local/bin/vuego

WORKDIR /app

EXPOSE 8080

USER nobody
ENTRYPOINT ["/usr/local/bin/vuego"]
CMD ["tour"]
