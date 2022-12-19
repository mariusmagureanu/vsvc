FROM alpine
RUN apk add varnish
COPY ./bin/vsvc /vsvc
ENTRYPOINT ["/vsvc"]
