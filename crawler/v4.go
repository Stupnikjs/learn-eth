package main

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
)

var TARGETIP net.IP = net.ParseIP("18.138.108.67") // Bootnode Fondation

func V4() {

	privKey, _ := crypto.GenerateKey()
	ping := PingSetup(privKey)
	packet, hash, err := v4wire.Encode(privKey, ping)
	if err != nil {
		fmt.Println("Erreur d'encodage:", err)
		return
	}

	fmt.Printf("Hash du Ping Initial (ReplyTok): %x\n", hash)

	// 5. Envoi UDP
	conn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	targetAddr := &net.UDPAddr{IP: TARGETIP, Port: 30303}

	// envoie le paquet
	_, err = conn.WriteToUDP(packet, targetAddr)
	if err != nil {
		fmt.Println("Erreur réseau:", err)
		return
	}

	buffer := make([]byte, 1280)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		ProcessPacket(conn, buffer[:n], privKey, remoteAddr)

		if err != nil {
			fmt.Printf("Paquet malformé reçu de %s: %v\n", remoteAddr, err)
			continue
		}

		// 3. Identification du type de message via un Type Switch

	}
}

func sendBackPong() {}

func buildPong(hash []byte, to v4wire.Endpoint, privkey *ecdsa.PrivateKey) *v4wire.Pong {
	return &v4wire.Pong{
		To: to,
		// 2. ReplyTok: C'est le HASH du Ping que tu viens de recevoir
		// C'est ce qui lie ton Pong à son Ping.
		ReplyTok: hash,
		// 3. Expiration: Standard (20 minutes dans le futur)
		Expiration: uint64(time.Now().Add(20 * time.Minute).Unix()),
		// 4. ENRSeq: Optionnel (0 si tu n'as pas de record ENR)
		ENRSeq: CreateMyENR(privkey).Seq(),
	}
}

func PingSetup(privkey *ecdsa.PrivateKey) *v4wire.Ping {
	toEndpoint := v4wire.Endpoint{
		IP:  TARGETIP,
		UDP: 30303,
		TCP: 30303,
	}
	fromEndpoint := v4wire.Endpoint{
		IP:  net.ParseIP("0.0.0.0"), // Ton IP (0.0.0.0 laisse la node détecter)
		UDP: 30303,
		TCP: 30303,
	}
	return &v4wire.Ping{
		Version:    4, // Version du protocole Discovery (v4)
		From:       fromEndpoint,
		To:         toEndpoint,
		ENRSeq:     CreateMyENR(privkey).Seq(),
		Expiration: uint64(time.Now().Add(20 * time.Minute).Unix()),
	}

}

func ProcessPacket(conn *net.UDPConn, buffer []byte, privkey *ecdsa.PrivateKey, remoteAddr *net.UDPAddr) {
	sender := v4wire.Endpoint{
		IP:  remoteAddr.IP, // Ton IP (0.0.0.0 laisse la node détecter)
		UDP: 30303,
		TCP: 30303,
	}
	if !remoteAddr.IP.Equal(TARGETIP) {
		fmt.Println("process only response from target node")
		return
	}
	packet, pubkey, hash, err := v4wire.Decode(buffer)
	if err != nil {
		fmt.Println(err)
	}
	switch p := packet.(type) {
	case *v4wire.Pong:
		fmt.Printf("✅ PONG reçu de la node ID: %x\n", pubkey[:8])
		fmt.Printf("   Correspond au Ping Hash: %x\n", p.ReplyTok)

	case *v4wire.Neighbors:
		fmt.Printf("🌐 REÇU %d NOUVEAUX VOISINS de %s\n", len(p.Nodes), remoteAddr)
		for _, node := range p.Nodes {
			fmt.Printf("   -> Node ID: %x | IP: %s | Port TCP: %d\n",
				node.ID[:8], node.IP, node.TCP)
			// C'est ici que tu alimentes ta base de données de crawler !
		}

	case *v4wire.Ping:
		fmt.Printf("📥 Reçu PING de %s. On devrait lui répondre Pong !\n", remoteAddr)

		pong := buildPong(hash, sender, privkey)
		packet, _, err := v4wire.Encode(privkey, pong)
		if err != nil {
			fmt.Printf("Paquet malformé reçu de %s: %v\n", remoteAddr, err)

		}
		_, err = conn.WriteToUDP(packet, remoteAddr)
		if err != nil {
			fmt.Println(err)
		}
		findNode := &v4wire.Findnode{
			Target:     pubkey,
			Expiration: uint64(time.Now().Add(20 * time.Minute).Unix()),
			Rest:       []rlp.RawValue{},
		}
		packet, _, err = v4wire.Encode(privkey, findNode)
		_, err = conn.WriteToUDP(packet, remoteAddr)
		if err != nil {
			fmt.Println(err)
		}

	}
}

func CreateMyENR(privKey *ecdsa.PrivateKey) *enr.Record {
	// 1. On crée une structure de clé locale compatible avec enode
	// (C'est elle qui possède la méthode de signature)
	db, _ := enode.OpenDB("") // Une chaîne vide avec OpenDB crée une DB en mémoire
	localNode := enode.NewLocalNode(db, privKey)

	// 2. On configure les informations de notre record
	localNode.Set(enr.IP(net.ParseIP("0.0.0.0")))
	localNode.Set(enr.UDP(30303))
	localNode.Set(enr.TCP(30303))

	// 3. On récupère le record signé
	// La signature se fait automatiquement via la clé privée fournie au début
	return localNode.Node().Record()
}
