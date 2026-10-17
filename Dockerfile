FROM golang:1.27.2 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go test -c -o /out/e2e.test ./e2e \
    && CGO_ENABLED=0 go build -o /out/sampleapp ./cmd/sampleapp

FROM scratch AS sampleapp
COPY --from=build /out/sampleapp /sampleapp
USER 65532:65532
ENV SAMPLE_ADDR=0.0.0.0:3000 GIN_MODE=release
EXPOSE 3000
HEALTHCHECK --interval=2s --timeout=2s --start-period=5s --retries=15 CMD ["/sampleapp", "-healthcheck"]
ENTRYPOINT ["/sampleapp"]

FROM scratch AS e2e
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/e2e.test /e2e.test
USER 65532:65532
ENV GIN_MODE=release
ENTRYPOINT ["/e2e.test"]
CMD ["-test.v"]
