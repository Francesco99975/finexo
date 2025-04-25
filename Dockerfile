FROM golang:1.23.4-alpine3.19 AS build

RUN apk --no-cache add gcc g++ make git

WORKDIR /go/src/app

COPY . .

RUN go mod tidy

RUN mv .prod.env .env

RUN GOOS=linux go build -ldflags="-s -w" -o ./bin/finexo ./cmd/server/*.go

FROM frolvlad/alpine-glibc:alpine-3.20 AS release

RUN apk update && apk upgrade && apk --no-cache add ca-certificates \
    harfbuzz \
    ttf-freefont \
    chromium \
    nss \
    freetype \
    libx11 \
    libxcomposite \
    libxrandr \
    libxdamage \
    libxi \
    libxcursor \
    libxinerama \
    # libc6-compat \
    alsa-lib \
    dbus \
    fontconfig \
    libjpeg-turbo \
    libpng \
    libstdc++ \
    libxshmfence \
    mesa-gl \
    pango \
    udev \
    libatk-1.0 \
    libatk-bridge-2.0 \
    at-spi2-core \
    cups-libs \
    mesa-gbm \
    libxkbcommon \
    ffmpeg \
    python3 \
    yt-dlp

WORKDIR /go/bin

COPY --from=build /go/src/app/bin /go/bin
COPY --from=build /go/src/app/.env /go/bin/
COPY --from=build /go/src/app/static /go/bin/static
COPY --from=build /go/src/app/sql /go/bin/sql
COPY --from=build /go/src/app/seeds /go/bin/seeds

EXPOSE 5869

ENTRYPOINT /go/bin/finexo --port 5869
