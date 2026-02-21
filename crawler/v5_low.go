package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

type StaticHeader struct {
	ProtocolID [6]byte
	Version    uint16
	Flag       uint8
	Nonce      [12]byte
	AuthSize   uint16
}

func LowLevelHandshake(enode string) {
	// 1. Initialisation de ton identité
	_, localPubKey, localMaskKey := BuildLocalSetting()

	// 2. Calcul de la clé de masquage de la CIBLE (pour l'envoi)
	targetMaskKey := GetMaskingKeyFromEnode(enode)
	target, targetPort, target_pub, _ := EnodeExtracter(enode)

	conn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	targetAddr := &net.UDPAddr{IP: target, Port: targetPort}

	// generation aleatoire de l'iv et du paquet
	// Il doit être unique pour chaque paquet afin que le masquage soit imprévisible.
	iv := make([]byte, 16)
	rand.Read(iv)
	targetID := PubKeyToNodeID(target_pub)
	sourceID := PubKeyToNodeID(localPubKey) // Ton ID à toi

	// 1. Appel de ta fonction
	toMask, _, findNodeMsg := BuildFindNodePacket(targetID, sourceID)

	// 2. Calcul du Tag (En clair, non masqué)
	tag := make([]byte, 32)
	targetHash := crypto.Keccak256(targetID[:])
	for i := 0; i < 32; i++ {
		tag[i] = targetHash[i] ^ sourceID[i]
	}

	// 3. Masquage AES-CTR du bloc (Header + AuthData)
	block, _ := aes.NewCipher(targetMaskKey)
	// On utilise l'IV (16 octets) généré au début de LowLevelHandshake
	stream := cipher.NewCTR(block, iv)
	maskedHeader := make([]byte, len(toMask))
	stream.XORKeyStream(maskedHeader, toMask)

	// 4. Assemblage final du paquet UDP
	// [IV 16] + [TAG 32] + [MASKED_HEADER 55] + [MESSAGE_EN_CLAIR]
	packet := make([]byte, 0, 16+32+len(maskedHeader)+len(findNodeMsg))
	packet = append(packet, iv...)
	packet = append(packet, tag...)
	packet = append(packet, maskedHeader...)
	packet = append(packet, findNodeMsg...)

	// 5. Envoi
	conn.WriteToUDP(packet, targetAddr)
	fmt.Println("Paquet len: ", len(packet))

	// 4. Lecture de la réponse (WHOAREYOU)
	buf := make([]byte, 1280)

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, _, err := conn.ReadFromUDP(buf)

		// 5. DÉMASQUAGE de la réponse
		// Le bootnode a utilisé NOTRE localMaskKey pour nous répondre
		if n <= 16 {
			continue
		}
		respIV := buf[:16]
		respMaskedHeader := buf[16:n] // Le reste du paquet

		localBlock, _ := aes.NewCipher(localMaskKey[:16])
		localStream := cipher.NewCTR(localBlock, respIV)

		decryptedHeader := make([]byte, len(respMaskedHeader))
		localStream.XORKeyStream(decryptedHeader, respMaskedHeader)

		fmt.Printf("Header WHOAREYOU démasqué (hex): %x\n", decryptedHeader)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func BuildLocalSetting() (*ecdsa.PrivateKey, *ecdsa.PublicKey, []byte) {

	// Tu génères une paire de clés (privée/publique).
	privKey, _ := crypto.GenerateKey()
	// Tu calcules ton Node ID (Hash Keccak-256 de ta clé publique).
	pubKey := privKey.Public().(*ecdsa.PublicKey)
	// 1. Convertir la clé ECDSA en bytes bruts (65 octets avec le préfixe 0x04)
	pubBytes := crypto.FromECDSAPub(pubKey)
	// 2. Retirer le premier octet (0x04) pour avoir les 64 octets bruts
	rawPubBytes := pubBytes[1:]
	// Tu calcules ta propre clé de masquage (les 16 premiers octets de ton Node ID). Elle te servira à lire la réponse.
	nodeID := crypto.Keccak256(rawPubBytes)
	return privKey, pubKey, nodeID
}

func GetMaskingKeyFromEnode(enodeURL string) []byte {
	// 1. Extraire la partie hexadécimale de la pubkey
	// Format: enode://PUBKEY@IP:PORT
	parts := strings.Split(strings.TrimPrefix(enodeURL, "enode://"), "@")
	pubkeyHex := parts[0]
	pubBytes, _ := hex.DecodeString(pubkeyHex)
	nodeID := crypto.Keccak256(pubBytes)
	return nodeID[:16]
}

func BuildFindNodePacket(targetID, sourceID [32]byte) ([]byte, [12]byte, []byte) {
	// --- 1. Le Message FINDNODE (RLP) ---
	// Format : [0x02] + RLP([request-id, [0]])
	// Pour faire simple et low-level :
	requestId := make([]byte, 8)
	rand.Read(requestId)

	// RLP manuel pour [requestId, [0]]
	// 0xcb (liste de 11 octets) + 0x88 (string 8 octets) + requestId + 0xc1 (liste 1 octet) + 0x00
	msgPayload := append([]byte{0x02, 0xcb, 0x88}, requestId...)
	msgPayload = append(msgPayload, 0xc1, 0x00)

	// --- 2. Le Header Statique ---
	h := StaticHeader{
		Version:  1,
		Flag:     0,  // flagMessage
		AuthSize: 32, // Taille fixe pour le messageAuthData (SrcID)
	}
	copy(h.ProtocolID[:], "discv5")
	rand.Read(h.Nonce[:])

	// --- 3. MessageAuthData (indispensable pour les messages) ---
	// Pour un message de type flagMessage, AuthData doit contenir le SourceID (32 octets)
	authData := sourceID[:]

	// --- 4. Sérialisation du Header Statique ---
	hBuf := new(bytes.Buffer)
	binary.Write(hBuf, binary.BigEndian, h)

	// --- 5. Masquage ---
	// On masque [StaticHeader + AuthData]
	toMask := append(hBuf.Bytes(), authData...)
	// (Note : Le message payload FINDNODE lui-même est crypté en GCM normalement,
	// mais ici on s'en fiche car le but est de faire échouer le déchiffrement
	// pour recevoir le WHOAREYOU)

	return toMask, h.Nonce, msgPayload
}
