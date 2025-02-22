# Gunakan image base Golang
FROM golang:alpine

# Install git untuk dependency management
RUN apk update && apk add --no-cache git

# Set direktori kerja dalam container
WORKDIR /app

# Salin semua file dari direktori saat ini ke dalam /app di container
COPY . /app

# Download semua dependency Go
RUN go mod tidy 

# Build binary aplikasi
RUN go build -o auth_service

# Jalankan aplikasi saat container dimulai
ENTRYPOINT ["/app/auth_service"]

