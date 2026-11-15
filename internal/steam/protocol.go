package steam

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// EMsg values, from SteamDatabase/Protobufs enums_clientserver.proto.
const (
	emsgMulti                                = 1
	emsgServiceMethod                        = 146
	emsgServiceMethodResponse                = 147
	emsgServiceMethodCallFromClient          = 151
	emsgClientHeartBeat                      = 703
	emsgClientLogOff                         = 706
	emsgClientChangeStatus                   = 716
	emsgClientLogOnResponse                  = 751
	emsgClientLoggedOff                      = 757
	emsgClientPersonaState                   = 766
	emsgClientFriendsList                    = 767
	emsgClientRequestFriendData              = 815
	emsgClientLogon                          = 5514
	emsgServiceMethodCallFromClientNonAuthed = 9804
	emsgClientHello                          = 9805
)

const (
	protoMask       = 0x80000000
	protocolVersion = 65580
	jobNone         = ^uint64(0)
)

// EResult values used by this package.
const (
	EResultOK                 = 1
	EResultFail               = 2
	EResultInvalidPassword    = 5
	EResultLoggedInElsewhere  = 6
	EResultFileNotFound       = 9
	EResultAccessDenied       = 15
	EResultServiceUnavailable = 20
	EResultRevoked            = 26
	EResultExpired            = 27
	EResultTryAnotherCM       = 48
	EResultRateLimitExceeded  = 84
)

// Error is a Steam EResult returned by the server.
type Error struct {
	EResult int32
	Message string
}

func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("steam: eresult %d (%s)", e.EResult, e.Message)
	}
	return fmt.Sprintf("steam: eresult %d", e.EResult)
}

// header is CMsgProtoBufHeader, restricted to the fields this app uses.
type header struct {
	steamID       uint64
	sessionID     int32
	jobSource     uint64
	jobTarget     uint64
	targetJobName string
	eresult       int32
	errorMessage  string
	realm         uint32
}

func newHeader() header {
	return header{jobSource: jobNone, jobTarget: jobNone, eresult: EResultFail}
}

func (h header) encode() []byte {
	var e enc
	e.fixed64(1, h.steamID)
	e.int32(2, h.sessionID)
	if h.jobSource != jobNone {
		e.fixed64(10, h.jobSource)
	}
	if h.jobTarget != jobNone {
		e.fixed64(11, h.jobTarget)
	}
	if h.targetJobName != "" {
		e.string(12, h.targetJobName)
	}
	if h.realm != 0 {
		e.varint(32, uint64(h.realm))
	}
	return e.b
}

func decodeHeader(b []byte) (header, error) {
	h := newHeader()
	err := walk(b, func(f field) {
		switch f.Num {
		case 1:
			h.steamID = f.Value
		case 2:
			h.sessionID = f.i32()
		case 10:
			h.jobSource = f.Value
		case 11:
			h.jobTarget = f.Value
		case 12:
			h.targetJobName = f.str()
		case 13:
			h.eresult = f.i32()
		case 14:
			h.errorMessage = f.str()
		}
	})
	return h, err
}

type packet struct {
	emsg  uint32
	proto bool
	hdr   header
	body  []byte
}

var errShortPacket = errors.New("steam: short packet")

func parsePacket(b []byte) (packet, error) {
	if len(b) < 8 {
		return packet{}, errShortPacket
	}
	raw := binary.LittleEndian.Uint32(b)
	p := packet{emsg: raw &^ protoMask, proto: raw&protoMask != 0}
	if !p.proto {
		// Extended (non-protobuf) header is 36 bytes. None of the messages this
		// app handles use it, so only the body offset matters.
		if len(b) < 36 {
			return packet{}, errShortPacket
		}
		p.hdr = newHeader()
		p.body = b[36:]
		return p, nil
	}
	hl := int(binary.LittleEndian.Uint32(b[4:]))
	if hl < 0 || 8+hl > len(b) {
		return packet{}, errShortPacket
	}
	h, err := decodeHeader(b[8 : 8+hl])
	if err != nil {
		return packet{}, err
	}
	p.hdr = h
	p.body = b[8+hl:]
	return p, nil
}

func buildPacket(emsg uint32, h header, body []byte) []byte {
	hb := h.encode()
	out := make([]byte, 8, 8+len(hb)+len(body))
	binary.LittleEndian.PutUint32(out, emsg|protoMask)
	binary.LittleEndian.PutUint32(out[4:], uint32(len(hb)))
	out = append(out, hb...)
	return append(out, body...)
}
