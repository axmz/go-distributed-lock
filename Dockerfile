FROM golang:1.24.2-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o go-distrilock ./cmd/distrilock

CMD ["./go-distrilock"]


# FROM golang:1.24.2-alpine

# WORKDIR /app

# # Install Air
# RUN go install github.com/air-verse/air@latest

# # Copy go.mod and go.sum first for caching
# COPY go.mod go.sum ./
# RUN go mod download

# # Copy the rest of the source code
# COPY . .

# # Expose Air config if you want to customize (optional)
# # COPY .air.toml ./

# CMD ["air"]
