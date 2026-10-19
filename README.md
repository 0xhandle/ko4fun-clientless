# ko4fun-clientless

KO4FUN login sunucusuna oyun olmadan giriş yapıp sunucu listesini çeken komut satırı aracı.

```
$ ko4fun-clientless -user kullanici -password sifre
ID  SUNUCU     KANAL           ADRES           OYUNCU  KAPASİTE  DURUM
1   IMMORTAL1  IMMORTAL1(DE5)  51.xxx.xx.xx    0       10000     0
1   IMMORTAL1  IMMORTAL1(FR1)  79.xxx.xx.xxx   0       10000     0
1   NEMESIS1   NEMESIS1(DE5)   51.xxx.xx.xx    0       10000     0
2   NEMESIS2   NEMESIS2(US1)   45.xx.xxx.xxx   0       10000     0
...
```

## Gereksinimler

Go 1.24 veya üstü ve bir KO4FUN hesabı. Liste sadece giriş yapan hesaba gönderiliyor.

## Derleme

```
make build
./bin/ko4fun-clientless -user kullanici -password sifre
```

ya da direkt:

```
go run ./cmd/ko4fun-clientless -user kullanici -password sifre
```

## Kullanım

| Parametre | Varsayılan | Açıklama |
|-----------|------------|----------|
| `-user` | | Kullanıcı adı (zorunlu) |
| `-password` | | Şifre (zorunlu) |
| `-host` | `login.ko4.fun` | Login sunucusu adresi |
| `-port` | `15100` | Login sunucusu portu |
| `-json` | `false` | Listeyi json olarak yazdırır |
| `-timeout` | `10s` | Bağlantı zaman aşımı |

Giriş başarısız olursa sunucunun döndürdüğü sebep yazdırılıyor (yanlış şifre, engellenmiş hesap, hesap oyunda vb.).

## Protokol

Paketler `AA55 | uzunluk:u16le | veri | 55AA` şeklinde gidip geliyor. Resmi sunucunun aksine login tarafında şifreleme yok, her şey düz gidiyor. Yazılar `uzunluk:u16le` ve ardından karakterler şeklinde.

Giriş paketi `F3 | kullanıcı_adı | şifre`. Şifre düz gönderilmiyor, resmi istemcideki kodlamayla gönderiliyor:

1. Şifre 4 baytlık bloklara bölünüyor, son blok sıfırla dolduruluyor.
2. Her blok little endian `uint32` olarak okunup `0x3E8` ekleniyor.
3. Sonucun bitleri sabit bir tabloya göre yer değiştiriyor.
4. Çıkan sayı 7 haneli base36 (`0-9A-Z`) olarak, düşük hane önce yazılıyor.

Sunucu `F3 | ? | ? | sonuç` ile cevap veriyor. Sonuç 1 ise giriş başarılı, diğer kodlar:

| Kod | Anlamı |
|-----|--------|
| 2 | Kullanıcı adı veya şifre yanlış |
| 3 | Kullanıcı adı veya şifre geçersiz |
| 4 | Hesap engellenmiş |
| 5 | Hesap şu an oyunda |
| 6 | Sunucu hatası |
| 16 | OTP gerekiyor |
| 255 | Giriş denemesi başarısız |

Ardından `F5 01 00` ile sunucu listesi isteniyor. Gelen liste standart Knight Online formatında: her sunucu için adres, iç adres, `sunucu|kanal` adı, oyuncu sayısı, kapasite, durum ve iki ırkın kral bilgileri.

## Test

```
make test
```
