FROM golang:1.23-bookworm AS build
ENV GOTOOLCHAIN=auto
RUN apt-get update && apt-get install -y --no-install-recommends nodejs npm && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN cd frontend && npm ci && npm run build && cd ..
RUN go build -tags netgo -ldflags '-s -w' -o /copyrem .

FROM python:3.11-slim AS full
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg ca-certificates && rm -rf /var/lib/apt/lists/*
RUN pip install --no-cache-dir demucs soundfile
WORKDIR /app
COPY --from=build /copyrem /app/copyrem
COPY --from=build /src/frontend/dist /app/frontend/dist
COPY --from=build /src/settings.json /app/settings.json
EXPOSE 8080
CMD ["/app/copyrem"]

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /copyrem /app/copyrem
COPY --from=build /src/frontend/dist /app/frontend/dist
COPY --from=build /src/settings.json /app/settings.json
EXPOSE 8080
CMD ["/app/copyrem"]
