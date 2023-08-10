package packet

import (
	"errors"
	"fmt"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"github.com/templexxx/reedsolomon"
	"sort"
)

// Errors that can be thrown by methods in this package:
const (
	ErrEmptyBytes      = "cannot encode empty bytes"
	ErrDupe            = "found duplicate packet, no redundancy, decoding failed"
	ErrLostNoRedundant = "no redundancy with %d lost of %d"
	ErrMismatch        = "found disagreement about common data in segment %d of %d" +
		" in field %s value: got %v expected %v"
	ErrNotEnough = "too many lost to recover in section %d, have %d, need " +
		"%d minimum"
)

// JoinPackets a collection of Packets together.
func JoinPackets(packets Packets) (pkts Packets, msg []byte, e error) {
	if len(packets) == 0 {
		e = errors.New("empty packets")
		return
	}
	// By sorting the packets we know we can iterate through them and detect
	// missing and duplicated items by simple rules.
	var tmp Packets
	for i := range packets {
		if packets[i] != nil {
			tmp = append(tmp, packets[i])
		}
	}
	packets = tmp
	sort.Sort(packets)
	lp := len(packets)
	p := packets[0]
	// Construct the segments map.
	overhead := p.GetOverhead()
	// log.D.Ln(
	// 	int(p.Length), len(p.Data)+overhead, overhead, int(p.Parity))
	segMap := NewSegments(
		int(p.Length), len(p.Data)+overhead, overhead, int(p.Parity))
	segCount := segMap[len(segMap)-1].PEnd
	// log.D.S("segMap", segMap)
	length, red := p.Length, p.Parity
	prevSeq := p.Seq
	var discard []int
	// Check that the data that should be common to all packets is common,
	// and no sequence number is repeated.
	for i, ps := range packets {
		// Skip the first because we are comparing the rest to it. It is
		// arbitrary which item is reference because all should be the
		// same.
		if i == 0 {
			continue
		}
		// fails that the sequence number isn't repeated.
		if ps.Seq == prevSeq {
			if red == 0 {
				e = fmt.Errorf(ErrDupe)
				return
			}
			// Check the data is the same, then discard the second
			// if they match.
			if sha256.Single(ps.Data) ==
				sha256.Single(packets[prevSeq].Data) {
				discard = append(discard, int(ps.Seq))
				// Node need to go on, we will discard this one.
				continue
			}
		}
		prevSeq = ps.Seq
		// All messages must have the same parity settings.
		if ps.Parity != red {
			e = fmt.Errorf(ErrMismatch, i, lp, "parity",
				ps.Parity, red)
			return
		}
		// All segments must specify the same total message length.
		if ps.Length != length {
			e = fmt.Errorf(ErrMismatch, i, lp, "length",
				ps.Length, length)
			return
		}
	}
	// Duplicates somehow found. Remove them.
	for i := range discard {
		// Subtracting the iterator accounts for the backwards shift of
		// the shortened slice.
		packets = RemovePacket(packets, discard[i]-i)
		lp--
	}
	// fails there is all pieces if there is no redundancy.
	log.T.Ln("red", red, "lp", lp, "segCount", segCount)
	if red == 0 && lp < segCount {
		e = fmt.Errorf(ErrLostNoRedundant, segCount-lp, segCount)
		return
	}
	msg = make([]byte, 0, length)
	// If all segments were received we can just concatenate the data shards
	if segCount == lp {
		for _, sm := range segMap {
			segments := make([][]byte, 0, sm.DEnd-sm.DStart)
			for i := sm.DStart; i < sm.DEnd; i++ {
				segments = append(segments, packets[i].Data)
			}
			msg = join(msg, segments, sm.SLen, sm.Last)
		}
		return
	}
	pkts = make(Packets, segCount)
	// Collate to correctly ordered, so gaps are easy to find
	for i := range packets {
		pkts[packets[i].Seq] = packets[i]
	}
	// Count and collate found and lost segments, adding empty segments if
	// there is lost.
	for si, sm := range segMap {
		var lD, lP, hD, hP []int
		var segments [][]byte
		for i := sm.DStart; i < sm.DEnd; i++ {
			idx := i - sm.DStart
			if pkts[i] == nil {
				lD = append(lD, idx)
			} else {
				hD = append(hD, idx)
			}
		}
		for i := sm.DEnd; i < sm.PEnd; i++ {
			idx := i - sm.DStart
			if pkts[i] == nil {
				lP = append(lP, idx)
			} else {
				hP = append(hP, idx)
			}
		}
		dLen := sm.DEnd - sm.DStart
		lhD, lhP := len(hD), len(hP)
		if lhD+lhP < dLen {
			// segment cannot be corrected
			e = fmt.Errorf(ErrNotEnough, si, lhD+lhP, dLen)
			return
		}
		// if we have all the data segments we can just assemble and
		// return.
		if lhD == dLen {
			for i := sm.DStart; i < sm.DEnd; i++ {
				segments = append(segments, pkts[i].Data)
			}
			msg = join(msg, segments, sm.SLen, sm.Last)
			continue
		}
		// We have enough to do correction
		for i := sm.DStart; i < sm.PEnd; i++ {
			if pkts[i] == nil {
				segments = append(segments,
					make([]byte, sm.SLen))
			} else {
				segments = append(segments,
					pkts[i].Data)
			}
		}
		var rs *reedsolomon.RS
		if rs, e = reedsolomon.New(dLen, sm.PEnd-sm.DEnd); fails(e) {
			return
		}
		have := append(hD, hP...)
		if e = rs.Reconst(segments, have, lD); fails(e) {
			return
		}
		msg = join(msg, segments[:dLen], sm.SLen, sm.Last)
	}
	return
}

func RemovePacket(slice Packets, s int) Packets {
	return append(slice[:s], slice[s+1:]...)
}

// SplitToPackets creates a series of packets including the defined Reed Solomon
// parameters for extra parity shards and the return encryption public key for a
// reply.
//
// The last packet that falls short of the segmentSize is padded random bytes.
func SplitToPackets(pp *PacketParams, segSize int,
	ks *crypto.KeySet) (dataShards int, packets [][]byte, e error) {
	if pp.Data == nil || len(pp.Data) == 0 {
		e = fmt.Errorf(ErrEmptyBytes)
		return
	}
	pp.Length = len(pp.Data)
	overhead := pp.GetOverhead()
	ss := segSize - overhead
	segments := slice.Segment(pp.Data, ss)
	segMap := NewSegments(pp.Length, segSize, pp.GetOverhead(), pp.Parity)
	dataShards = segMap[len(segMap)-1].DEnd
	var p [][]byte
	p, e = segMap.AddParity(segments)
	for i := range p {
		pp.Data, pp.Seq = p[i], i
		pp.From = ks.Next()
		var s []byte
		if s, e = EncodePacket(pp); fails(e) {
			return
		}
		packets = append(packets, s)
	}
	return
}

func join(msg []byte, segments [][]byte, sLen, last int) (b []byte) {
	b = slice.Cat(segments...)
	if sLen != last {
		b = b[:len(b)-sLen+last]
	}
	b = append(msg, b...)
	return
}
