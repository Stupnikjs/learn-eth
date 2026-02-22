package main

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"golang.org/x/crypto/hkdf"
)

func EphemeralKeys() (*ecdh.PrivateKey, *ecdh.PublicKey, error) {
	ephemeralPriv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	ephemeralPubBytes := ephemeralPriv.PublicKey()
	return ephemeralPriv, ephemeralPubBytes, nil
}

func DeriveKeys(n1, n2 enode.ID, z []byte, idNonce [16]byte) (encKey, decKey []byte, err error) {
	// 1. Préparation du "Salt" (Le sel est le ID-Nonce)
	salt := idNonce[:]

	// 2. Préparation de "Info" (Concaténation des NodeIDs et du label)
	info := []byte("discovery v5 key derivation")
	info = append(info, n1[:]...)
	info = append(info, n2[:]...)

	// 3. Initialisation de HKDF avec SHA-256
	kdf := hkdf.New(sha256.New, z, salt, info)

	// 4. Extraction de 16 octets pour AES-GCM (ou 32 selon la config)
	// On extrait deux clés : Initiator Key et Recipient Key
	k1 := make([]byte, 16)
	k2 := make([]byte, 16)

	io.ReadFull(kdf, k1)
	io.ReadFull(kdf, k2)

	return k1, k2, nil
}
