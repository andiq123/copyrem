FROM node:24-alpine AS frontend
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS build
ENV GOTOOLCHAIN=auto
RUN apk add --no-cache ca-certificates
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist ./frontend/dist
RUN go build -tags netgo -ldflags '-s -w' -o /copyrem .

FROM alpine:3.23
RUN apk add --no-cache ffmpeg ca-certificates
WORKDIR /app
COPY --from=build /copyrem /app/copyrem
COPY --from=build /src/frontend/dist /app/frontend/dist
COPY --from=build /src/settings.json /app/settings.json
EXPOSE 8080
CMD ["/app/copyrem"]
