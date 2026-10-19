package serverlist

import (
	"errors"
	"testing"

	"github.com/0xhandle/ko4fun-clientless/internal/packet"
)

func encode(servers ...Server) []byte {
	w := packet.NewWriter().Byte(Opcode).Byte(0x01).Byte(0x00).Byte(byte(len(servers)))
	for _, s := range servers {
		w.String(s.Host).
			String(s.LanHost).
			String(s.Name + "|" + s.Channel).
			Uint16(s.UserCount).
			Byte(s.Slot).
			Byte(0).
			Uint16(s.ID).
			Uint16(s.GroupID).
			Uint16(s.MaxUsers).
			Byte(s.Status).
			Byte(1).
			String(s.KarusKing).
			String("").
			String(s.ElmoradKing).
			String("")
	}
	return w.Bytes()
}

func TestParse(t *testing.T) {
	want := []Server{
		{ID: 1, Name: "IMMORTAL1", Channel: "IMMORTAL1(DE1)", Host: "192.0.2.10", LanHost: "192.0.2.10", GroupID: 1, Slot: 1, MaxUsers: 3000, UserCount: 120, KarusKing: "karus", ElmoradKing: "elmo"},
		{ID: 2, Name: "NEMESIS1", Channel: "NEMESIS1(FR1)", Host: "192.0.2.11", LanHost: "192.0.2.11", GroupID: 1, Slot: 2, MaxUsers: 3000, Status: 1},
	}

	got, err := Parse(encode(want...))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("%d sunucu geldi, beklenen %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("sunucu %d = %+v, beklenen %+v", i, got[i], want[i])
		}
	}
}

func TestParseFallsBackToLanHost(t *testing.T) {
	got, err := Parse(encode(Server{Name: "A", Channel: "B", LanHost: "192.0.2.5"}))
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Host != "192.0.2.5" {
		t.Errorf("adres = %q, beklenen 192.0.2.5", got[0].Host)
	}
}

func TestParseRejectsOtherOpcode(t *testing.T) {
	data := encode(Server{Name: "A", Channel: "B"})
	data[0] = 0xF6
	if _, err := Parse(data); !errors.Is(err, ErrUnexpectedOpcode) {
		t.Fatalf("hata = %v, beklenen %v", err, ErrUnexpectedOpcode)
	}
}

func TestParseTruncated(t *testing.T) {
	data := encode(Server{Name: "A", Channel: "B", Host: "192.0.2.1"})
	if _, err := Parse(data[:20]); !errors.Is(err, packet.ErrShortPacket) {
		t.Fatalf("hata = %v, beklenen %v", err, packet.ErrShortPacket)
	}
}
