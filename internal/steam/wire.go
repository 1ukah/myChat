package steam

import (
	"errors"
	"math"

	"google.golang.org/protobuf/encoding/protowire"
)

// enc is a tiny protobuf encoder. Steam messages are encoded by hand so that
// only the handful of fields this app needs are compiled in.
type enc struct{ b []byte }

func (e *enc) varint(f int, v uint64) {
	e.b = protowire.AppendTag(e.b, protowire.Number(f), protowire.VarintType)
	e.b = protowire.AppendVarint(e.b, v)
}

func (e *enc) int32(f int, v int32) { e.varint(f, uint64(int64(v))) }

func (e *enc) bool(f int, v bool) {
	if v {
		e.varint(f, 1)
	} else {
		e.varint(f, 0)
	}
}

func (e *enc) fixed64(f int, v uint64) {
	e.b = protowire.AppendTag(e.b, protowire.Number(f), protowire.Fixed64Type)
	e.b = protowire.AppendFixed64(e.b, v)
}

func (e *enc) fixed32(f int, v uint32) {
	e.b = protowire.AppendTag(e.b, protowire.Number(f), protowire.Fixed32Type)
	e.b = protowire.AppendFixed32(e.b, v)
}

func (e *enc) bytes(f int, v []byte) {
	e.b = protowire.AppendTag(e.b, protowire.Number(f), protowire.BytesType)
	e.b = protowire.AppendBytes(e.b, v)
}

func (e *enc) string(f int, v string) { e.bytes(f, []byte(v)) }

// field is one decoded protobuf field. For length-delimited fields Bytes is set;
// for varint/fixed fields Num is set.
type field struct {
	Num   protowire.Number
	Type  protowire.Type
	Value uint64
	Bytes []byte
}

func (f field) i32() int32      { return int32(f.Value) }
func (f field) u32() uint32     { return uint32(f.Value) }
func (f field) boolean() bool   { return f.Value != 0 }
func (f field) str() string     { return string(f.Bytes) }
func (f field) f32() float32    { return math.Float32frombits(uint32(f.Value)) }

var errBadProto = errors.New("steam: malformed protobuf")

// walk calls fn for every field in b, in order.
func walk(b []byte, fn func(f field)) error {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return errBadProto
		}
		b = b[n:]
		fl := field{Num: num, Type: typ}
		switch typ {
		case protowire.VarintType:
			v, m := protowire.ConsumeVarint(b)
			if m < 0 {
				return errBadProto
			}
			fl.Value, b = v, b[m:]
		case protowire.Fixed64Type:
			v, m := protowire.ConsumeFixed64(b)
			if m < 0 {
				return errBadProto
			}
			fl.Value, b = v, b[m:]
		case protowire.Fixed32Type:
			v, m := protowire.ConsumeFixed32(b)
			if m < 0 {
				return errBadProto
			}
			fl.Value, b = uint64(v), b[m:]
		case protowire.BytesType:
			v, m := protowire.ConsumeBytes(b)
			if m < 0 {
				return errBadProto
			}
			fl.Bytes, b = v, b[m:]
		default:
			m := protowire.ConsumeFieldValue(num, typ, b)
			if m < 0 {
				return errBadProto
			}
			b = b[m:]
			continue
		}
		fn(fl)
	}
	return nil
}
