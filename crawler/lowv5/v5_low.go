package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common/mclock"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover/v5wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type messageAuthData struct {
	SrcID [32]byte
}

type StaticHeader struct {
	ProtocolID [6]byte
	Version    uint16
	Flag       uint8
	Nonce      []byte
	AuthSize   uint16
}

type Header struct {
	IV []byte
	StaticHeader
	AuthData []byte
	src      [32]byte
}

func V5low() {

	privKey, _ := crypto.GenerateKey()
	db, _ := enode.OpenDB("")
	localNode := enode.NewLocalNode(db, privKey)

	_ = enode.MustParse(MainnetBootnodes[0])

	// 3. Initialisation du Codec (Gère le masquage, le RLP et les IV)
	// On passe 'nil' pour le protocole ID car il utilisera par défaut "discv5"
	protcolID := [6]byte{'d', 'i', 's', 'c', 'v', '5'}
	codec := v5wire.NewCodec(localNode, privKey, mclock.System{}, &protcolID)

	localID := GetNodeID(privKey)
	// 4. Ouvrir la connexion UDP
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		log.Fatalf("Erreur UDP: %v", err)
	}
	defer conn.Close()
	ip, _, pubkey, err := EnodeExtracter(MainnetBootnodes[0])
	nodeID := PubKeyToNodeID(pubkey)
	// 7. Envoi au bootnode
	targetAddr := &net.UDPAddr{IP: ip, Port: 30303}
	encoded := BuildRandomPacket(localID, nodeID)
	_, err = conn.WriteToUDP(encoded, targetAddr)
	if err != nil {
		log.Fatalf("Erreur Envoi: %v", err)
	}
	fmt.Printf(">> Paquet Unknown envoyé à %s\n", targetAddr)

	// 8. Boucle de lecture pour capturer le WHOAREYOU
	buf := make([]byte, 1280)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		log.Fatalf("Aucune réponse (Timeout): %v", err)
	}

	// 9. Décodage du paquet reçu
	fromID, _, respPacket, err := codec.Decode(buf[:n], addr.String())
	if err != nil {
		log.Fatalf("Erreur Décodage réponse: %v", err)
	}

	fmt.Println(fromID, respPacket.Name())
}

func BuildRandomPacket(localNodeID [32]byte, targetNodeID [32]byte) []byte {
	// privkey, _ := crypto.GenerateKey()
	auth := messageAuthData{SrcID: localNodeID}
	header, _ := BaseHeader(auth, localNodeID)

	var buf bytes.Buffer
	var dest bytes.Buffer

	binary.Write(&buf, binary.BigEndian, header.ProtocolID)
	binary.Write(&buf, binary.BigEndian, header.Version)
	buf.WriteByte(header.Flag)
	buf.Write(header.Nonce) // Écrit les 12 octets
	binary.Write(&buf, binary.BigEndian, header.AuthSize)
	block, err := aes.NewCipher(targetNodeID[:16])
	if err != nil {
		fmt.Println(err)
	}
	_, err = buf.Write(header.AuthData)

	if err != nil {
		fmt.Println(err)
	}
	maskedbuf := make([]byte, len(buf.Bytes()))
	randomData := make([]byte, 30)
	cleanbuf := buf.Bytes()
	stream := cipher.NewCTR(block, header.IV)
	dest.Write(header.IV)
	stream.XORKeyStream(maskedbuf, cleanbuf)
	dest.Write(maskedbuf)
	rand.Read(randomData)
	dest.Write(randomData)
	return dest.Bytes()

}

func BaseHeader(auth messageAuthData, nodeid [32]byte) (Header, error) {
	nonce := make([]byte, 12)
	iv := make([]byte, 16)
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, auth)

	// You must check for errors when reading randomness!
	if _, err := rand.Read(iv); err != nil {
		return Header{}, err
	}
	if _, err := rand.Read(nonce); err != nil {
		return Header{}, err
	}

	return Header{
		IV: iv,
		StaticHeader: StaticHeader{
			ProtocolID: [6]byte{'d', 'i', 's', 'c', 'v', '5'},
			Version:    1,
			Flag:       0, // Using 0 for a standard message flag
			Nonce:      nonce,
			AuthSize:   uint16(binary.Size(auth)),
		},
		AuthData: buf.Bytes(),
		src:      nodeid,
		// src is usually an internal field or part of AuthData
	}, nil
}

func encryptGCM(dest, key, nonce, plaintext, authData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(fmt.Errorf("can't create block cipher: %v", err))
	}
	aesgcm, err := cipher.NewGCMWithNonceSize(block, 12)
	if err != nil {
		panic(fmt.Errorf("can't create GCM: %v", err))
	}
	return aesgcm.Seal(dest, nonce, plaintext, authData), nil
}
