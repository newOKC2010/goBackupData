FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o backup .

FROM alpine:3.21
RUN apk add --no-cache postgresql16-client cifs-utils tzdata
ENV TZ=Asia/Bangkok
WORKDIR /app
COPY --from=build /app/backup .
CMD ["./backup"]
