FROM node:24-alpine AS client-builder

WORKDIR /app/client
COPY client/package*.json ./
RUN npm install
COPY client/ ./
RUN npm run build

FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=client-builder /app/client/dist ./client/dist
COPY --from=builder /server ./

ENV FILE_STORAGE_PATH=/file_store
ENV POSTGRES_HOST=db

EXPOSE 3000

CMD ["./server"]
