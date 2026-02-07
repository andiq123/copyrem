FROM golang:1.25-alpine AS build
RUN apk add --no-cache nodejs npm
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN cd frontend && npm ci && npm run build && cd ..
RUN go build -tags netgo -ldflags '-s -w' -o /app .

FROM alpine:3.21
RUN apk add --no-cache ffmpeg ca-certificates
COPY --from=build /app /app
COPY --from=build /src/frontend/dist /frontend/dist
EXPOSE 8080
CMD ["/app"]
