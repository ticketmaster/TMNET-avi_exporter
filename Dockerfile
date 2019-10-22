##################################################################
# Build binary.
##################################################################
FROM golang:1.12.9 as build
WORKDIR /go/src/github.com/ticketmaster/TMNET-avi_exporter
ADD . .
RUN make build-linux
##################################################################\
# Build non-root container.
##################################################################
FROM alpine:latest
EXPOSE 8080
RUN addgroup -g 1000 aviexporter &&\
    adduser aviexporter -D aviexporter -u 1000 -G aviexporter
USER 1000:1000
COPY --from=build /go/src/github.com/ticketmaster/TMNET-avi_exporter/target/linux/aviexporter /home/aviexporter/aviexporter
COPY --from=build /go/src/github.com/ticketmaster/TMNET-avi_exporter/lib /home/aviexporter/lib
WORKDIR /home/aviexporter
ENTRYPOINT ["/home/aviexporter/aviexporter"]
