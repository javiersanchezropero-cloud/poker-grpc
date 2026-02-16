FROM golang:1.24 AS build
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /app/server /server
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/server"]
