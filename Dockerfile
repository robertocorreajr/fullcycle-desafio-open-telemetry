FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
RUN go mod download
COPY . .
RUN go build -o /app/server ./cmd/server
RUN go build -o /app/server-a ./cmd/service-a

FROM alpine:3.18
WORKDIR /root/
COPY --from=build /app/server ./server
COPY --from=build /app/server-a ./server-a
EXPOSE 8080 8081
CMD ["./server"]
