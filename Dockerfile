FROM golang:1.23-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ashen-crown ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/ashen-crown /app/ashen-crown
COPY web /app/web
RUN mkdir -p /app/data/runs /app/data/saves
EXPOSE 8080
CMD ["/app/ashen-crown", "-web", "/app/web", "-data", "/app/data"]
