
Log dosyalarını kurallara göre analiz eden ve gerçek zamanlı uyarı üreten, Go ile yazılmış ve Docker ile paketlenmiş bir web uygulaması.

## Gereksinimler
- Docker
- Docker Compose

## Kurulum ve Çalıştırma
1. Depoyu klonlayın ve klasöre girin:
   - `git clone https://github.com/07enesavci/CLI-Tabanli-Log-Analiz-Araci`
   - `cd log-analyzer`
1. Uygulamayı başlatın:
   - `docker-compose up -d`
2. Tarayıcıdan açın:
   - `http://localhost:8080`

## Giriş Bilgileri
- Kullanıcı adı: `admin`
- Şifre: `admin`

## Özellikler
- Dashboard: özet istatistikler ve uyarılar
- Gerçek zamanlı izleme: log akışını takip
- Analiz: seçili log dosyalarını analiz
- Kurallar ve log dosyaları: yapılandırmayı görüntüleme

## Yapılandırma
Kurallar ve izlenecek log dosyaları `config/rules.yaml` içinde tanımlıdır. Değişikliklerden sonra uygulamayı yeniden başlatın.

## Docker Notları
- Uygulama konteyneri `8080` portunu kullanır.
- `docker-compose.yml` içinde `./config` klasörü konteynere bağlanır.
- Host makinedeki `/var/log` ve `/tmp` dizinleri konteynere bağlanmıştır.

## Docker Komutları
- Logları izlemek için:
  - `docker-compose logs -f`
- Durdurmak için:
  - `docker-compose down`
- Yeniden derleyip başlatmak için:
  - `docker-compose up -d --build`

## Durdurma
- `docker-compose down`
