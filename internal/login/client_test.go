package login

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/0xhandle/ko4fun-clientless/internal/frame"
	"github.com/0xhandle/ko4fun-clientless/internal/packet"
	"github.com/0xhandle/ko4fun-clientless/internal/password"
	"github.com/0xhandle/ko4fun-clientless/internal/serverlist"
)

const (
	testUser     = "tester"
	testPassword = "Deneme123*"
)

type fakeServer struct {
	listener net.Listener
	failure  chan error
	result   byte
	reason   string
}

func startFakeServer(t *testing.T, result byte, reason string) *fakeServer {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &fakeServer{listener: l, failure: make(chan error, 1), result: result, reason: reason}
	go func() {
		conn, err := l.Accept()
		if err != nil {
			s.failure <- err
			return
		}
		defer conn.Close()
		if err := s.serve(frame.New(conn, 5*time.Second)); err != nil {
			s.failure <- err
		}
	}()
	t.Cleanup(func() { l.Close() })
	return s
}

func (s *fakeServer) serve(conn *frame.Conn) error {
	request, err := conn.Read()
	if err != nil {
		return err
	}
	r := packet.NewReader(request)
	if op := r.Byte(); op != opLogin {
		return fmt.Errorf("beklenmeyen opcode %#x", op)
	}
	if user, pass := r.String(), r.String(); user != testUser || pass != password.Encode(testPassword) {
		return fmt.Errorf("beklenmeyen kimlik %q %q", user, pass)
	}

	reply := packet.NewWriter().Byte(opLogin).Byte(0).Byte(0).Byte(s.result)
	if s.result == resultBanned {
		reply.Uint16(0).Uint16(0).String(s.reason)
	} else {
		reply.Uint16(0xFFFF).String(testUser)
	}
	if err := conn.Write([]byte{0xFD}); err != nil {
		return err
	}
	if err := conn.Write(reply.Bytes()); err != nil {
		return err
	}
	if s.result != resultSuccess {
		return nil
	}

	request, err = conn.Read()
	if err != nil {
		return err
	}
	if string(request) != string([]byte{serverlist.Opcode, 0x01, 0x00}) {
		return fmt.Errorf("beklenmeyen liste isteği %x", request)
	}

	list := packet.NewWriter().Byte(serverlist.Opcode).Byte(1).Byte(0).Byte(1).
		String("192.0.2.10").String("192.0.2.10").String("IMMORTAL1|IMMORTAL1(DE1)").
		Uint16(5).Byte(1).Byte(0).Uint16(1).Uint16(1).Uint16(3000).Byte(0).Byte(1).
		String("").String("").String("").String("")
	return conn.Write(list.Bytes())
}

func fetch(t *testing.T, server *fakeServer) ([]serverlist.Server, error) {
	t.Helper()
	servers, err := FetchServerList(context.Background(), Config{
		Address:  server.listener.Addr().String(),
		Username: testUser,
		Password: testPassword,
		Timeout:  5 * time.Second,
	})
	select {
	case serverErr := <-server.failure:
		t.Fatalf("sahte sunucu: %v", serverErr)
	default:
	}
	return servers, err
}

func TestFetchServerList(t *testing.T) {
	servers, err := fetch(t, startFakeServer(t, resultSuccess, ""))
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 1 || servers[0].Channel != "IMMORTAL1(DE1)" || servers[0].UserCount != 5 {
		t.Fatalf("beklenmeyen liste %+v", servers)
	}
}

func TestFetchServerListRejected(t *testing.T) {
	for _, tc := range []struct {
		result byte
		reason string
		text   string
	}{
		{resultWrongLogin, "", "şifre yanlış"},
		{resultBanned, "", "hesap engellenmiş"},
		{resultBanned, "hile", "hesap engellenmiş: hile"},
		{resultInGame, "", "oyunda"},
		{42, "", "bilinmeyen sonuç kodu 42"},
	} {
		_, err := fetch(t, startFakeServer(t, tc.result, tc.reason))
		if !errors.Is(err, ErrLoginRejected) || !strings.Contains(err.Error(), tc.text) {
			t.Errorf("sonuç %d: hata = %v, %q içermeli", tc.result, err, tc.text)
		}
	}
}
