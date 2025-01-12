package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/deroproject/derohe/cryptography/bn256"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/globals"
	"golang.org/x/crypto/chacha20poly1305"
)

const INDENTIFIER = "<DERO ENCRYPTED>"
const MSG_MIN_LENGTH = 189
const ZEROHASH = "0000000000000000000000000000000000000000000000000000000000000000"

var privateKey *big.Int

// Return public key and shared key(s)
func GenerateSharedSecrets(receivers []string) (public_key string, shared_keys []string, key [32]byte, err error) {

	k := crypto.RandomScalar()
	x := new(bn256.G1).ScalarMult(crypto.G, k)
	public_key = hex.EncodeToString(x.EncodeCompressed())
	if len(public_key) != 66 {
		return "", nil, key, fmt.Errorf("invalid length of public key")
	}

	s := crypto.RandomScalar()
	sy := new(bn256.G1).ScalarMult(crypto.G, s)

	for _, a := range receivers {
		addr, err := globals.ParseValidateAddress(a)
		if err != nil {
			return "", nil, key, err
		}

		r_pub, err := bn256.Decompress(addr.PublicKey.EncodeCompressed())
		if err != nil {
			return "", nil, key, err
		}

		shared_key := new(bn256.G1).Add(new(bn256.G1).Set(sy), new(bn256.G1).ScalarMult(r_pub, k))
		shared_keys = append(shared_keys, hex.EncodeToString(shared_key.EncodeCompressed()))
	}

	sha_key := sha256.Sum256(sy.EncodeCompressed())

	return public_key, shared_keys, sha_key, nil
}

func AddKeyword(msg string) string {

	len := len(msg)

	var pos []int
	var count int

	for i := 0; i < len; i++ {
		if msg[i] == ' ' {
			count++
			pos = append(pos, i)
		}
	}
	if msg[len-1] == ' ' {
		count--
	}
	if count < 2 {
		return fmt.Sprintf("%s\n%s", INDENTIFIER, msg)
	}

	var p_bytes [4]byte

	rand.Read(p_bytes[:])
	p := binary.BigEndian.Uint32(p_bytes[:])
	p %= (uint32(count) - 1)

	return fmt.Sprintf("%s %s %s", msg[:pos[p]], INDENTIFIER, msg[pos[p]+1:])
}

func EncryptMessage(msg string, key [32]byte) (encrypted string, modified string, err error) {

	if !strings.Contains(msg, INDENTIFIER) {
		msg = AddKeyword(msg)
	}
	data, err := EncryptMessageWithKey(key, []byte(msg))
	if err != nil {
		return "", "", err
	}

	if len(msg) < 28 {
		return "", "", fmt.Errorf("message is too short, we need at least 28 characters")
	}

	return hex.EncodeToString(data), msg, nil
}

func EncryptMessageWithKey(Key [32]byte, Data []byte) (result []byte, err error) {

	nonce := make([]byte, chacha20poly1305.NonceSize, chacha20poly1305.NonceSize)
	cipher, err := chacha20poly1305.New(Key[:])
	if err != nil {
		return
	}

	_, err = rand.Read(nonce)
	if err != nil {
		return
	}
	Data = cipher.Seal(Data[:0], nonce, Data, nil)

	result = append(Data, nonce...)

	return
}

func DecryptMessageWithKey(Key [32]byte, Data []byte) (result []byte, err error) {

	if len(Data) < 28 {
		err = fmt.Errorf("invalid data")
		return
	}

	data_without_nonce := Data[0 : len(Data)-chacha20poly1305.NonceSize]
	nonce := Data[len(Data)-chacha20poly1305.NonceSize:]

	cipher, err := chacha20poly1305.New(Key[:])
	if err != nil {
		return
	}

	return cipher.Open(result[:0], nonce, data_without_nonce, nil)
}

func DecryptMessages(data string) (contents []string) {

	for _, m := range GetMessages(data) {
		if !SanityCheck(m) {
			continue
		}
		pub, commits := GetCommitments(m)
		msg_hex, err := PayloadCheck(GetPayload(m))
		if err != nil {
			continue
		}
		if content, err := Decrypt(msg_hex, pub, commits); err != nil {
			continue
		} else {
			contents = append(contents, content)
		}
	}

	return
}

func Decrypt(msg []byte, pubkey []byte, commits [][]byte) (content string, err error) {

	shared_keys, err := GetSharedKeys(pubkey, commits)
	if err != nil {
		return "", err
	}

	for _, k := range shared_keys {
		decrypted, err := DecryptMessageWithKey(k, msg)
		if err != nil {
			continue
		}
		plain, err := hex.DecodeString(hex.EncodeToString(decrypted))
		if err != nil {
			log.Println(err)
			continue
		}
		if !HasIdentifier(string(plain)) {
			continue
		}
		content = string(plain)
	}

	return content, nil
}

func GetMessages(data string) []string {

	if !strings.Contains(data, "+") {
		return []string{data}
	}
	msgs := strings.Split(data, "+")

	return msgs
}

func SanityCheck(msg string) bool {

	if len(msg) < MSG_MIN_LENGTH {
		return false
	}
	if strings.Count(msg, "x") != 1 {
		return false
	}
	keys_len := strings.Index(msg, "x")
	if keys_len < 0 || keys_len%66 != 0 {
		return false
	}

	keys := keys_len / 66
	if keys < 2 {
		return false
	}
	for i := 0; i < keys; i++ {
		if pub, err := hex.DecodeString(msg[(i * 66):((i + 1) * 66)]); err != nil {
			return false
		} else {
			if _, err = bn256.Decompress(pub); err != nil {
				return false
			}
		}
	}

	return true
}

func GetCommitments(msg string) (p []byte, c [][]byte) {

	keys_len := strings.Index(msg, "x") + 1
	p, _ = hex.DecodeString(msg[:66])

	keys := (keys_len / 66) - 1
	for i := 1; i <= keys; i++ {
		key, _ := hex.DecodeString(msg[i*66 : ((i + 1) * 66)])
		c = append(c, key)
	}

	return
}

func GetPayload(msg string) string {
	return msg[strings.Index(msg, "x")+1:]
}

func PayloadCheck(msg string) ([]byte, error) {
	hex, err := hex.DecodeString(msg)

	return hex, err
}

func GetSharedKeys(pubkey []byte, commits [][]byte) (shared_keys [][32]byte, err error) {

	commit := new(bn256.G1)
	pub, _ := bn256.Decompress(pubkey)

	for _, c := range commits {
		commit, _ = bn256.Decompress(c)
		shared := new(bn256.G1).Add(new(bn256.G1).Set(commit), new(bn256.G1).Neg(new(bn256.G1).ScalarMult(pub, privateKey)))
		shared_keys = append(shared_keys, sha256.Sum256(shared.EncodeCompressed()))
	}

	return
}

func HasIdentifier(msg string) bool {
	return strings.Contains(msg, INDENTIFIER)
}
