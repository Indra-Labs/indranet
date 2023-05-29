package packet

import (
	"crypto/cipher"
	"fmt"
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"time"
)

const (
	// Overhead is the base overhead on a packet, use GetOverhead to add any extra
	// as found in a Packet.
	Overhead    = 4 + crypto.PubKeyLen + crypto.CloakLen + nonce.IVLen
	PacketMagic = "rpkt"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

// GetOverhead returns the amount of the message that will not be part of the
// payload.
func (ep PacketParams) GetOverhead() int {
	return Overhead + nonce.IDLen + 7
}

// sort.Interface implementation.

func (p Packets) Len() int           { return len(p) }
func (p Packets) Less(i, j int) bool { return p[i].Seq < p[j].Seq }

// GetOverhead returns the packet frame overhead given the settings found in the
// packet.
func (p *Packet) GetOverhead() int {
	return Overhead
}

type (
	// PacketParams defines the parameters for creating a ( split) packet given a
	// set of keys, cipher, and data. To, From, Data are required, Parity is
	// optional, set it to define a level of Reed Solomon redundancy on the split
	// packets.
	PacketParams struct {
		ID     nonce.ID
		To     *crypto.Pub
		From   *crypto.Prv
		Parity int
		Seq    int
		Length int
		Data   []byte
	}
	// Packet is the standard format for an encrypted, possibly segmented message
	// container with parameters for Reed Solomon Forward Error Correction.
	Packet struct {
		ID nonce.ID
		// Seq specifies the segment number of the message, 4 bytes long.
		Seq uint16
		// Length is the number of segments in the batch
		Length uint32
		// Parity is the ratio of redundancy. The remainder from 256 is the
		// proportion from 256 of data shards in a packet batch.
		Parity byte
		// Data is the message.
		Data      []byte
		TimeStamp time.Time
	}
	// Packets is a slice of pointers to packets.
	Packets []*Packet
)

// DecodePacket decodes a packet and return the Packet with encrypted payload
// and signer's public key. This assumes GetKeysFromPacket succeeded and the
// matching private key was found.
func DecodePacket(d []byte, from *crypto.Pub, to *crypto.Prv,
	iv nonce.IV) (f *Packet, e error) {
	pktLen := len(d)
	if pktLen < Overhead {
		// If this isn't checked the slice operations later can hit bounds
		// errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			Overhead, pktLen)
		log.E.Ln(e)
		return
	}
	f = &Packet{TimeStamp: time.Now()}
	// copy the nonce
	c := new(slice.Cursor)
	copy(iv[:], d[c.Inc(4+crypto.PubKeyLen+crypto.CloakLen):c.Inc(nonce.IVLen)])
	var blk cipher.Block
	if blk = ciph.GetBlock(to, from, "packet decode"); fails(e) {
		return
	}
	// This decrypts the rest of the packet.
	data := d[*c:]
	ciph.Encipher(blk, iv, data)
	var id slice.Bytes
	id, data = slice.Cut(data, nonce.IDLen)
	copy(f.ID[:], id)
	seq := slice.NewUint16()
	length := slice.NewUint32()
	seq, data = slice.Cut(data, slice.Uint16Len)
	f.Seq = uint16(slice.DecodeUint16(seq))
	length, data = slice.Cut(data, slice.Uint32Len)
	f.Length = uint32(slice.DecodeUint32(length))
	f.Parity, data = data[0], data[1:]
	f.Data = data
	return
}

// EncodePacket creates a Packet, encrypts the payload using the given private
// from key and the public to key, serializes the form and signs the bytes. the
// signature to the end.
func EncodePacket(ep *PacketParams) (pkt []byte, e error) {
	var blk cipher.Block
	if blk = ciph.GetBlock(ep.From, ep.To, "packet encode"); fails(e) {
		return
	}
	nonc := nonce.New()
	Seq := slice.NewUint16()
	slice.EncodeUint16(Seq, ep.Seq)
	Length := slice.NewUint32()
	slice.EncodeUint32(Length, ep.Length)
	pkt = make([]byte, slice.SumLen(Seq, Length,
		ep.Data)+1+Overhead+nonce.IDLen)
	// Append pubkey used for encryption key derivation.
	k := crypto.DerivePub(ep.From).ToBytes()
	cloaked := crypto.GetCloak(ep.To)
	// Copy nonce, address and key over top of the header.
	c := new(slice.Cursor)
	c.Inc(4)
	copy(pkt[*c:c.Inc(crypto.PubKeyLen)], k[:])
	copy(pkt[*c:c.Inc(crypto.CloakLen)], cloaked[:])
	copy(pkt[*c:c.Inc(nonce.IVLen)], nonc[:])
	copy(pkt[*c:c.Inc(nonce.IDLen)], ep.ID[:])
	copy(pkt[*c:c.Inc(slice.Uint16Len)], Seq)
	copy(pkt[*c:c.Inc(slice.Uint32Len)], Length)
	pkt[*c] = byte(ep.Parity)
	copy(pkt[c.Inc(1):], ep.Data)
	// Encrypt the encrypted part of the data.
	ciph.Encipher(blk, nonc, pkt[Overhead:])
	// last but not least, the packet check header, which protects the entire
	// packet.
	checkBytes := sha256.Single(pkt[4:])
	copy(pkt[:4], checkBytes[:4])
	return
}

// GetKeysFromPacket returns the ToHeaderPub field of the message in order,
// checks the packet checksum and recovers the public key.
//
// After this, if the matching private key to the cloaked address returned is
// found, it is combined with the public key to generate the cipher and the
// entire packet should then be decrypted, and the DecodePacket function will
// then decode a OnionSkin.
func GetKeysFromPacket(d []byte) (from *crypto.Pub, to crypto.PubKey, iv nonce.IV,
	e error) {
	pktLen := len(d)
	if pktLen < Overhead {
		// If this isn't checked the slice operations later can hit
		// bounds errors.
		e = fmt.Errorf("packet too small, min %d, got %d",
			Overhead, pktLen)
		log.E.Ln(e)
		return
	}
	// split off the signature and recover the public key
	var k crypto.PubBytes
	var chek []byte
	c := new(slice.Cursor)
	chek = d[:c.Inc(4)]
	checkHash := sha256.Single(d[*c:])
	if string(chek) != string(checkHash[:4]) {
		e = fmt.Errorf("check failed: got '%v', expected '%v'",
			chek, checkHash[:4])
		return
	}
	copy(k[:], d[*c:c.Inc(crypto.PubKeyLen)])
	copy(to[:], d[*c:c.Inc(crypto.CloakLen)])
	if from, e = crypto.PubFromBytes(k[:]); fails(e) {
		return
	}
	return
}

func (p Packets) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
