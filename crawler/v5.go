package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

/*
La taille minimale (44 octets)
Le protocole impose une taille minimale pour éviter les attaques par amplification UDP.

Si ton paquet fait moins de 44 octets, le nœud distant l'ignorera totalement.

En pratique, on envoie souvent un paquet d'environ 64 à 100 octets pour simuler un vrai message chiffré.

2. Le Header statique (après démasquage)
Le bootnode va essayer de démasquer ton paquet. Pour qu'il accepte de te répondre par un WHOAREYOU, le header démasqué doit contenir :

Protocol ID : Les 6 octets fixes discv5.
Version : Le chiffre 1.
Flag : La valeur 0 (qui signifie "Ordinary Packet").
AuthData : Ton Node ID (32 octets). C'est ainsi que le bootnode sait à qui il doit répondre.

3. Le Masquage correct (La clé du destinataire)
C'est le critère le plus important. Tu dois utiliser la Masking Key du bootnode (dérivée de sa PubKey dans l'enode) pour masquer ton header.
*/
func LowLevelHandshake(target net.IP, targetPort int, targetPubKeyHex string) {
	// 1. Initialisation de ton identité
	_, _, localMaskKey := BuildLocalSetting()

	// 2. Calcul de la clé de masquage de la CIBLE (pour l'envoi)
	targetMaskKey := GetMaskingKeyFromEnode(targetPubKeyHex)

	conn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	targetAddr := &net.UDPAddr{IP: target, Port: targetPort}

	// 3. Forger un "Random Packet" un peu plus crédible
	// On l'appelle random, mais il doit être masqué avec la clé de la cible
	iv := make([]byte, 16)
	rand.Read(iv)
	headerData := make([]byte, 45) // Taille d'un header standard
	rand.Read(headerData)

	// Masquage pour la cible
	block, _ := aes.NewCipher(targetMaskKey)
	stream := cipher.NewCTR(block, iv)
	maskedHeader := make([]byte, len(headerData))
	stream.XORKeyStream(maskedHeader, headerData)

	packet := append(iv, maskedHeader...)
	conn.WriteToUDP(packet, targetAddr)

	// 4. Lecture de la réponse (WHOAREYOU)
	buf := make([]byte, 1280)
	n, _, _ := conn.ReadFromUDP(buf)

	// 5. DÉMASQUAGE de la réponse
	// Le bootnode a utilisé NOTRE localMaskKey pour nous répondre
	respIV := buf[:16]
	respMaskedHeader := buf[16:n] // Le reste du paquet

	localBlock, _ := aes.NewCipher(localMaskKey)
	localStream := cipher.NewCTR(localBlock, respIV)

	decryptedHeader := make([]byte, len(respMaskedHeader))
	localStream.XORKeyStream(decryptedHeader, respMaskedHeader)

	fmt.Printf("Header WHOAREYOU démasqué (hex): %x\n", decryptedHeader)
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
	return privKey, pubKey, nodeID[:16]
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
