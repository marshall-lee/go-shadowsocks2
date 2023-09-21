package shadowaead

import (
	"crypto/rand"
	"errors"
	"io"
	"net/netip"
	"sync"

	"github.com/shadowsocks/go-shadowsocks2/internal"
	"github.com/shadowsocks/go-shadowsocks2/udp"
)

// ErrShortPacket means that the packet is too short for a valid encrypted packet.
var ErrShortPacket = errors.New("short packet")

var _zerononce [128]byte // read-only. 128 bytes is more than enough.

// Pack encrypts plaintext using Cipher with a randomly generated salt and
// returns a slice of dst containing the encrypted packet and any error occurred.
// Ensure len(dst) >= ciph.SaltSize() + len(plaintext) + aead.Overhead().
func Pack(dst, plaintext []byte, ciph Cipher) ([]byte, error) {
	saltSize := ciph.SaltSize()
	salt := dst[:saltSize]
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	aead, err := ciph.Encrypter(salt)
	if err != nil {
		return nil, err
	}
	internal.AddSalt(salt)

	if len(dst) < saltSize+len(plaintext)+aead.Overhead() {
		return nil, io.ErrShortBuffer
	}
	b := aead.Seal(dst[saltSize:saltSize], _zerononce[:aead.NonceSize()], plaintext, nil)
	return dst[:saltSize+len(b)], nil
}

// Unpack decrypts pkt using Cipher and returns a slice of dst containing the decrypted payload and any error occurred.
// Ensure len(dst) >= len(pkt) - aead.SaltSize() - aead.Overhead().
func Unpack(dst, pkt []byte, ciph Cipher) ([]byte, error) {
	saltSize := ciph.SaltSize()
	if len(pkt) < saltSize {
		return nil, ErrShortPacket
	}
	salt := pkt[:saltSize]
	aead, err := ciph.Decrypter(salt)
	if err != nil {
		return nil, err
	}
	if internal.CheckSalt(salt) {
		return nil, ErrRepeatedSalt
	}
	if len(pkt) < saltSize+aead.Overhead() {
		return nil, ErrShortPacket
	}
	if saltSize+len(dst)+aead.Overhead() < len(pkt) {
		return nil, io.ErrShortBuffer
	}
	b, err := aead.Open(dst[:0], _zerononce[:aead.NonceSize()], pkt[saltSize:], nil)
	return b, err
}

type udpConn struct {
	udp.Conn
	Cipher
	sync.Mutex
	buf []byte // write lock
}

// NewUDPConn wraps a udp.Conn with cipher
func NewUDPConn(c udp.Conn, ciph Cipher) udp.Conn {
	const maxPacketSize = 64 * 1024
	return &udpConn{Conn: c, Cipher: ciph, buf: make([]byte, maxPacketSize)}
}

// WriteTo encrypts b and write to addr using the embedded PacketConn.
func (c *udpConn) WriteTo(b []byte, addr netip.AddrPort) (int, error) {
	c.Lock()
	defer c.Unlock()
	buf, err := Pack(c.buf, b, c)
	if err != nil {
		return 0, err
	}
	_, err = c.Conn.WriteTo(buf, addr)
	return len(b), err
}

// ReadFrom reads from the embedded PacketConn and decrypts into b.
func (c *udpConn) ReadFrom(b []byte) (int, netip.AddrPort, error) {
	n, addr, err := c.Conn.ReadFrom(b)
	if err != nil {
		return n, addr, err
	}
	bb, err := Unpack(b[c.Cipher.SaltSize():], b[:n], c)
	if err != nil {
		return n, addr, err
	}
	copy(b, bb)
	return len(bb), addr, err
}
