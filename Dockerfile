FROM golang:1.17.1-buster AS builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build

FROM chromedp/headless-shell:latest
COPY --from=builder /build/parser /app/
COPY templates/ /app/templates
COPY static/ /app/static
WORKDIR /app
ENTRYPOINT ["/app/parser"]