package wg

import (
	"crypto/ecdh"
	"crypto/hmac"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"hash"
	"time"

	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	CONSTRUCTION = []byte("Noise_IKpsk2_25519_ChaChaPoly_BLAKE2s")
	IDENTIFIER   = []byte("WireGuard v1 zx2c4 Jason@zx2c4.com")
	LABEL_MAC1   = []byte("mac1----")
)

func HASH(data []byte) []byte {
	h := blake2s.Sum256(data)
	return h[:]
}

func HMAC(key, data []byte) []byte {
	mac := hmac.New(func() hash.Hash {
		h, _ := blake2s.New256(nil)
		return h
	}, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func AEAD(key []byte, ctr uint64, msg, authtxt []byte) []byte {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		panic(err)
	}
	nonce := make([]byte, 12)
	binary.LittleEndian.PutUint64(nonce[4:], ctr)
	return aead.Seal(nil, nonce, msg, authtxt)
}

func TAI64N() []byte {
	now := time.Now()
	seconds := uint64(now.Unix()) + (1 << 62) + 10
	nanoseconds := uint32(now.Nanosecond() / 1000)

	buf := make([]byte, 12)
	binary.BigEndian.PutUint64(buf[0:8], seconds)
	binary.BigEndian.PutUint32(buf[8:12], nanoseconds)
	return buf
}

func MAC(key, data []byte) []byte {
	h, err := blake2s.New128(key)
	if err != nil {
		panic(err)
	}
	h.Write(data)
	return h.Sum(nil)
}

func GenerateHandshakePacket(clientPrivateBytes, responderStaticPublic, reserved []byte) []byte {
	staticPrivate, err := ecdh.X25519().NewPrivateKey(clientPrivateBytes)
	if err != nil {
		panic(fmt.Sprintf("Invalid private key: %v", err))
	}
	staticPublic := staticPrivate.PublicKey().Bytes()

	chainingKey := HASH(CONSTRUCTION)
	hashVal := HASH(append(HASH(append(chainingKey, IDENTIFIER...)), responderStaticPublic...))

	ephemeralPrivate, err := ecdh.X25519().GenerateKey(crand.Reader)
	if err != nil {
		panic(err)
	}
	unencryptedEphemeral := ephemeralPrivate.PublicKey().Bytes()

	messageType := []byte{0x01}
	senderIndex := make([]byte, 4)
	binary.LittleEndian.PutUint32(senderIndex, 1)

	hashVal = HASH(append(hashVal, unencryptedEphemeral...))

	temp := HMAC(chainingKey, unencryptedEphemeral)
	chainingKey = HMAC(temp, []byte{0x01})

	remoteStaticPub, err := ecdh.X25519().NewPublicKey(responderStaticPublic)
	if err != nil {
		panic(fmt.Sprintf("Invalid responder public key: %v", err))
	}
	dh1, err := ephemeralPrivate.ECDH(remoteStaticPub)
	if err != nil {
		panic(err)
	}

	temp = HMAC(chainingKey, dh1)
	chainingKey = HMAC(temp, []byte{0x01})
	key := HMAC(temp, append(chainingKey, 0x02))

	encryptedStatic := AEAD(key, 0, staticPublic, hashVal)
	hashVal = HASH(append(hashVal, encryptedStatic...))

	dh2, err := staticPrivate.ECDH(remoteStaticPub)
	if err != nil {
		panic(err)
	}

	temp = HMAC(chainingKey, dh2)
	chainingKey = HMAC(temp, []byte{0x01})
	key = HMAC(temp, append(chainingKey, 0x02))

	encryptedTimestamp := AEAD(key, 0, TAI64N(), hashVal)
	hashVal = HASH(append(hashVal, encryptedTimestamp...))

	var packetBody []byte
	packetBody = append(packetBody, messageType...)
	packetBody = append(packetBody, reserved...)
	packetBody = append(packetBody, senderIndex...)
	packetBody = append(packetBody, unencryptedEphemeral...)
	packetBody = append(packetBody, encryptedStatic...)
	packetBody = append(packetBody, encryptedTimestamp...)

	mac1Key := HASH(append(LABEL_MAC1, responderStaticPublic...))
	mac1 := MAC(mac1Key, packetBody)
	mac2 := make([]byte, 16)

	var packet []byte
	packet = append(packet, packetBody...)
	packet = append(packet, mac1...)
	packet = append(packet, mac2...)

	return packet
}
