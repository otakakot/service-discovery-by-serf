// Code generated by ogen, DO NOT EDIT.

package api

import (
	"math/bits"
	"strconv"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/validate"
)

// Encode implements json.Marshaler.
func (s *Cluster) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *Cluster) encodeFields(e *jx.Encoder) {
	{
		e.FieldStart("id")
		e.Str(s.ID)
	}
	{
		e.FieldStart("rpcAddr")
		e.Str(s.RpcAddr)
	}
	{
		e.FieldStart("isLeader")
		e.Bool(s.IsLeader)
	}
}

var jsonFieldsNameOfCluster = [3]string{
	0: "id",
	1: "rpcAddr",
	2: "isLeader",
}

// Decode decodes Cluster from json.
func (s *Cluster) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode Cluster to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "id":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				v, err := d.Str()
				s.ID = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"id\"")
			}
		case "rpcAddr":
			requiredBitSet[0] |= 1 << 1
			if err := func() error {
				v, err := d.Str()
				s.RpcAddr = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"rpcAddr\"")
			}
		case "isLeader":
			requiredBitSet[0] |= 1 << 2
			if err := func() error {
				v, err := d.Bool()
				s.IsLeader = bool(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"isLeader\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode Cluster")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000111,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfCluster) {
					name = jsonFieldsNameOfCluster[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *Cluster) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *Cluster) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes Clusters as json.
func (s Clusters) Encode(e *jx.Encoder) {
	unwrapped := []Cluster(s)
	if unwrapped == nil {
		e.ArrEmpty()
		return
	}
	if unwrapped != nil {
		e.ArrStart()
		for _, elem := range unwrapped {
			elem.Encode(e)
		}
		e.ArrEnd()
	}
}

// Decode decodes Clusters from json.
func (s *Clusters) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode Clusters to nil")
	}
	var unwrapped []Cluster
	if err := func() error {
		unwrapped = make([]Cluster, 0)
		if err := d.Arr(func(d *jx.Decoder) error {
			var elem Cluster
			if err := elem.Decode(d); err != nil {
				return err
			}
			unwrapped = append(unwrapped, elem)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return errors.Wrap(err, "alias")
	}
	*s = Clusters(unwrapped)
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s Clusters) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *Clusters) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *ListClusterResponse) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *ListClusterResponse) encodeFields(e *jx.Encoder) {
	{
		if s.Clusters != nil {
			e.FieldStart("clusters")
			s.Clusters.Encode(e)
		}
	}
}

var jsonFieldsNameOfListClusterResponse = [1]string{
	0: "clusters",
}

// Decode decodes ListClusterResponse from json.
func (s *ListClusterResponse) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode ListClusterResponse to nil")
	}

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "clusters":
			if err := func() error {
				if err := s.Clusters.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"clusters\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode ListClusterResponse")
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *ListClusterResponse) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *ListClusterResponse) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}
