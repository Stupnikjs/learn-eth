package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common/mclock"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover/v5wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

var MainnetBootnodes = []string{
	// Ethereum Foundation Go Bootnodes
	"enode://d860a01f9722d78051619d1e2351aba3f43f943f6f00718d1b9baa4101932a1f5011f16bb2b1bb35db20d6fe28fa0bf09636d26a87d31de9ec6203eeedb1f666@18.138.108.67:30303", // bootnode-aws-ap-southeast-1-001
	"enode://22a8232c3abc76a16ae9d6c3b164f98775fe226f0917b0ca871128a74a8e9630b458460865bab457221f1d448dd9791d24c4e5d88786180ac185df813a68d4de@3.209.45.79:30303",   // bootnode-aws-us-east-1-001
	"enode://2b252ab6a1d0f971d9722cb839a42cb81db019ba44c08754628ab4a823487071b5695317c8ccd085219c3a03af063495b2f1da8d18218da2d6a82981b45e6ffc@65.108.70.101:30303", // bootnode-hetzner-hel
	"enode://4aeb4ab6c14b23e2c4cfdce879c04b0748a20d8e9b59e25ded2a08143e265c6c25936e74cbc8e641e3312ca288673d91f2f93f8e277de3cfa444ecdaaf982052@157.90.35.166:30303", // bootnode-hetzner-fsn
}

func HighV5() {
	// 1. Ton identité locale (Indispensable pour le codec)
	privKey, _ := crypto.GenerateKey()
	db, _ := enode.OpenDB("")
	localNode := enode.NewLocalNode(db, privKey)

	node := enode.MustParse(MainnetBootnodes[0])

	// 3. Initialisation du Codec (Gère le masquage, le RLP et les IV)
	// On passe 'nil' pour le protocole ID car il utilisera par défaut "discv5"
	protcolID := [6]byte{'d', 'i', 's', 'c', 'v', '5'}
	codec := v5wire.NewCodec(localNode, privKey, mclock.System{}, &protcolID)

	// 4. Ouvrir la connexion UDP
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		log.Fatalf("Erreur UDP: %v", err)
	}
	defer conn.Close()

	// 5. Créer un paquet 'Unknown' (Le moyen le plus sûr de forcer un WHOAREYOU)
	var nonce v5wire.Nonce
	if _, err := rand.Read(nonce[:]); err != nil {
		log.Fatal("Erreur génération nonce:", err)
	}
	packet := &v5wire.Unknown{Nonce: nonce}
	fmt.Println(packet.Kind())
	// 6. Encodage avec le Codec officiel
	// Encode s'occupe du Tag (XOR), du Masquage et de l'IV automatiquement
	encoded, nonce, err := codec.Encode(node.ID(), node.IP().String(), packet, nil)
	if err != nil {
		log.Fatalf("Erreur Encodage: %v", err)
	}

	// 7. Envoi au bootnode
	targetAddr := &net.UDPAddr{IP: node.IP(), Port: node.UDP()}
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
