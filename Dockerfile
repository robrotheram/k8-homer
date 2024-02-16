FROM b4bz/homer as UI_BUILDER

FROM golang:1.21.6 as GO_BUILDER
ARG VER
WORKDIR /
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine
LABEL org.opencontainers.image.source="https://github.com/robrotheram/k8-homer"
WORKDIR /app
COPY --from=GO_BUILDER /k8-homer /app/k8-homer
COPY --from=UI_BUILDER /www /app/www
RUN yes n | cp -i www/default-assets/config.yml.dist www/assets/config.yml &> /dev/null
EXPOSE 8080
ENTRYPOINT ["/app/k8-homer"]