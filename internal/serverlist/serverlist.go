package serverlist

import (
	"errors"
	"strings"

	"github.com/0xhandle/ko4fun-clientless/internal/packet"
)

const Opcode = 0xF5

var ErrUnexpectedOpcode = errors.New("serverlist: beklenmeyen opcode")

type Server struct {
	ID          uint16 `json:"id"`
	Name        string `json:"name"`
	Channel     string `json:"channel"`
	Host        string `json:"host"`
	LanHost     string `json:"lan_host"`
	GroupID     uint16 `json:"group_id"`
	Slot        uint8  `json:"slot"`
	Status      uint8  `json:"status"`
	UserCount   uint16 `json:"user_count"`
	MaxUsers    uint16 `json:"max_users"`
	KarusKing   string `json:"karus_king,omitempty"`
	ElmoradKing string `json:"elmorad_king,omitempty"`
}

func Parse(payload []byte) ([]Server, error) {
	r := packet.NewReader(payload)
	if op := r.Byte(); op != Opcode {
		if err := r.Err(); err != nil {
			return nil, err
		}
		return nil, ErrUnexpectedOpcode
	}
	r.Byte()
	r.Byte()
	count := int(r.Byte())

	servers := make([]Server, 0, count)
	for range count {
		s := readServer(r)
		if err := r.Err(); err != nil {
			return servers, err
		}
		servers = append(servers, s)
	}
	return servers, nil
}

func readServer(r *packet.Reader) Server {
	var s Server
	s.Host = r.String()
	s.LanHost = r.String()
	name, channel, _ := strings.Cut(r.String(), "|")
	s.Name = strings.TrimSpace(name)
	s.Channel = strings.TrimSpace(channel)

	s.UserCount = r.Uint16()
	s.Slot = r.Byte()
	r.Byte()
	s.ID = r.Uint16()
	s.GroupID = r.Uint16()
	s.MaxUsers = r.Uint16()
	s.Status = r.Byte()
	r.Byte()

	s.KarusKing = r.String()
	r.SkipString()
	s.ElmoradKing = r.String()
	r.SkipString()

	if s.Host == "" {
		s.Host = s.LanHost
	}
	return s
}
