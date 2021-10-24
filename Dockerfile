# The base go-image
FROM golang:alpine AS builder
# RUN apk add --update --no-cache \
#     --repository http://dl-3.alpinelinux.org/alpine/edge/community \
#     --repository http://dl-3.alpinelinux.org/alpine/edge/main \
#     build-base vips-dev
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=1 GOOS=linux go build -o /server .

# production image
FROM alpine
RUN apk add --update vips
RUN mkdir /app
EXPOSE 8000
WORKDIR /app

COPY --from=builder /server /app/

CMD [ "/app/server", "web" ]
