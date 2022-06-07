#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM golang:1.17-alpine as builder

ARG BUILD_VERSION=0.1

WORKDIR /go/src/
ENV CGO_ENABLED=0

COPY . /go/src
RUN CGO_ENABLED=0 go build -o /go/bin/tagger -ldflags "-X main.version=$BUILD_VERSION" cmd/tagger/main.go
#########################################################################################
FROM alpine:3.15 as runner

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="tagger" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to host the LT POS tagger service" \
      description="This image is used to host the LT POS tagger service" 

WORKDIR /app
EXPOSE 8000
ENV UID=999

RUN addgroup app && adduser -S -G app app --uid $UID app
RUN mkdir -p /app && chown -R app:app /app   

COPY --from=builder /go/bin/tagger /app/
COPY build/lt-pos-tagger/config.yaml /app/

RUN chown app:app /app/* /app
 
USER app

ENTRYPOINT ["./tagger"]
