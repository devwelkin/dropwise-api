# --- 1. Aşama: Build Aşaması ---
# Go'nun kurulu olduğu bir imajı temel alıyoruz. Alpine versiyonu daha küçük olduğu için tercih sebebi.
FROM golang:1.24-alpine AS builder

# Çalışma dizinini ayarlıyoruz.
WORKDIR /app

# Önce sadece bağımlılık dosyalarını kopyalıyoruz. Bu satırlar değişmediği sürece
# Docker, `go mod download` adımını cache'den kullanarak build hızını artırır.
COPY go.mod go.sum ./
RUN go mod download

# Şimdi projenin geri kalan tüm kodunu kopyalıyoruz.
COPY . .

# Kodumuzu build ediyoruz. CGO_ENABLED=0, statik bir binary oluşturmak için önemli.
# Bu sayede son imajımızda C kütüphanelerine ihtiyaç duymayız.
# Çıktı olarak `/app/server` adında bir dosya oluşacak.
# EĞER main.go dosyan BİR ALT KLASÖRDEYSE (örn: ./cmd/api/) SONDAN İKİNCİ ARGÜMANI GÜNCELLE.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/server ./cmd/api/

# --- 2. Aşama: Final Aşaması ---
# "Distroless" imajları, sadece uygulamanın çalışması için gereken minimum şeyleri içerir.
# İçinde shell bile yoktur. Bu, imaj boyutunu küçültür ve güvenliği artırır.
FROM gcr.io/distroless/static-debian11

# Builder aşamasında oluşturduğumuz binary'yi final imajına kopyalıyoruz.
COPY --from=builder /app/server /server

# Uygulamanın hangi portu dinlediğini belirtiyoruz. KENDİ PORTUNU YAZ.
EXPOSE 8080

# Konteyner başladığında çalıştırılacak komut.
ENTRYPOINT ["/server"]