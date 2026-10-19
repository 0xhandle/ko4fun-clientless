package login

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/0xhandle/ko4fun-clientless/internal/frame"
	"github.com/0xhandle/ko4fun-clientless/internal/packet"
	"github.com/0xhandle/ko4fun-clientless/internal/password"
	"github.com/0xhandle/ko4fun-clientless/internal/serverlist"
)

const opLogin = 0xF3

const (
	resultSuccess     = 1
	resultWrongLogin  = 2
	resultInvalid     = 3
	resultBanned      = 4
	resultInGame      = 5
	resultServerError = 6
	resultOTPRequired = 16
	resultFailed      = 255
)

var ErrLoginRejected = errors.New("login: giriş reddedildi")

type Config struct {
	Address  string
	Username string
	Password string
	Timeout  time.Duration
}

func FetchServerList(ctx context.Context, cfg Config) ([]serverlist.Server, error) {
	dialer := net.Dialer{Timeout: cfg.Timeout}
	raw, err := dialer.DialContext(ctx, "tcp", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("login: bağlantı kurulamadı: %w", err)
	}
	defer raw.Close()
	stop := context.AfterFunc(ctx, func() { raw.Close() })
	defer stop()

	conn := frame.New(raw, cfg.Timeout)

	request := packet.NewWriter().
		Byte(opLogin).
		String(cfg.Username).
		String(password.Encode(cfg.Password)).
		Bytes()
	if err := conn.Write(request); err != nil {
		return nil, fmt.Errorf("login: giriş paketi gönderilemedi: %w", err)
	}

	reply, err := waitFor(conn, opLogin)
	if err != nil {
		return nil, fmt.Errorf("login: giriş cevabı alınamadı: %w", err)
	}
	if err := checkLogin(reply); err != nil {
		return nil, err
	}

	if err := conn.Write([]byte{serverlist.Opcode, 0x01, 0x00}); err != nil {
		return nil, fmt.Errorf("login: sunucu listesi isteği gönderilemedi: %w", err)
	}
	list, err := waitFor(conn, serverlist.Opcode)
	if err != nil {
		return nil, fmt.Errorf("login: sunucu listesi alınamadı: %w", err)
	}
	return serverlist.Parse(list)
}

func checkLogin(reply []byte) error {
	r := packet.NewReader(reply[1:])
	r.Byte()
	r.Byte()
	result := r.Byte()
	if err := r.Err(); err != nil {
		return fmt.Errorf("login: giriş cevabı okunamadı: %w", err)
	}

	switch result {
	case resultSuccess:
		return nil
	case resultWrongLogin:
		return fmt.Errorf("%w: kullanıcı adı veya şifre yanlış ya da böyle bir kullanıcı yok", ErrLoginRejected)
	case resultInvalid:
		return fmt.Errorf("%w: kullanıcı adı veya şifre geçersiz", ErrLoginRejected)
	case resultBanned:
		r.Uint16()
		r.Uint16()
		if reason := r.String(); reason != "" {
			return fmt.Errorf("%w: hesap engellenmiş: %s", ErrLoginRejected, reason)
		}
		return fmt.Errorf("%w: hesap engellenmiş", ErrLoginRejected)
	case resultInGame:
		return fmt.Errorf("%w: hesap şu an oyunda", ErrLoginRejected)
	case resultServerError:
		return fmt.Errorf("%w: sunucu hatası", ErrLoginRejected)
	case resultOTPRequired:
		return fmt.Errorf("%w: hesap için OTP gerekiyor", ErrLoginRejected)
	case resultFailed:
		return fmt.Errorf("%w: giriş denemesi başarısız", ErrLoginRejected)
	default:
		return fmt.Errorf("%w: bilinmeyen sonuç kodu %d", ErrLoginRejected, result)
	}
}

func waitFor(conn *frame.Conn, opcode byte) ([]byte, error) {
	for {
		payload, err := conn.Read()
		if err != nil {
			return nil, err
		}
		if len(payload) > 0 && payload[0] == opcode {
			return payload, nil
		}
	}
}
