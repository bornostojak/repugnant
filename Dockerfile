FROM node:22-alpine AS web-build
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS server-build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /out/rpg-server ./cmd/rpg-server

FROM alpine:3.22
RUN addgroup -S rpg && adduser -S rpg -G rpg
WORKDIR /app
COPY --from=server-build /out/rpg-server ./rpg-server
COPY --from=web-build /src/web/dist ./web
USER rpg
EXPOSE 8080
ENV RPG_HTTP_ADDR=:8080
ENTRYPOINT ["./rpg-server"]
