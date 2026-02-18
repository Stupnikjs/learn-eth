package main

import (
	"fmt"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover/v4wire"
)

func main() {

	privKey, _ := crypto.GenerateKey()

	targetIP := net.ParseIP("18.138.108.67") // Bootnode Fondation

	toEndpoint := v4wire.Endpoint{
		IP:  targetIP,
		UDP: 30303,
		TCP: 30303,
	}

	fromEndpoint := v4wire.Endpoint{
		IP:  net.ParseIP("0.0.0.0"), // Ton IP (0.0.0.0 laisse la node détecter)
		UDP: 30303,
		TCP: 30303,
	}

	// 3. Création du Ping
	ping := &v4wire.Ping{
		Version:    4, // Version du protocole Discovery (v4)
		From:       fromEndpoint,
		To:         toEndpoint,
		Expiration: uint64(time.Now().Add(20 * time.Minute).Unix()),
	}

	/* Encode
	func Encode(priv *ecdsa.PrivateKey, req Packet) (packet, hash []byte, err error) {
		b := new(bytes.Buffer)
		b.Write(headSpace)        // 97 empty bytes
		b.WriteByte(req.Kind())
		packet = b.Bytes()
		sig, err := crypto.Sign(crypto.Keccak256(packet[headSize:]), priv)

		copy(packet[macSize:], sig)
		hash = crypto.Keccak256(packet[macSize:])
		copy(packet, hash)
		return packet, hash, nil
	}
	*/
	packet, hash, err := v4wire.Encode(privKey, ping)
	if err != nil {
		fmt.Println("Erreur d'encodage:", err)
		return
	}

	fmt.Printf("Paquet forgé ! Hash du Ping (ReplyTok): %x\n", hash)

	// 5. Envoi UDP
	conn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	targetAddr := &net.UDPAddr{IP: targetIP, Port: 30303}

	_, err = conn.WriteToUDP(packet, targetAddr)
	if err != nil {
		fmt.Println("Erreur réseau:", err)
		return
	}

	buffer := make([]byte, 1280)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		// 2. Utiliser v4wire pour décoder TOUTE l'enveloppe (Signature, Hash, Payload)
		// packet: l'interface (Ping, Pong, etc.)
		// pubkey: la clé publique de la node qui t'a répondu (Node ID)
		// hash: le hash keccak256 du paquet reçu
		packet, pubkey, hash, err := v4wire.Decode(buffer[:n])
		if err != nil {
			fmt.Printf("Paquet malformé reçu de %s: %v\n", remoteAddr, err)
			continue
		}

		// 3. Identification du type de message via un Type Switch
		switch p := packet.(type) {
		case *v4wire.Pong:
			fmt.Printf("✅ PONG reçu de la node ID: %x\n", pubkey[:8])
			fmt.Printf("   Correspond au Ping Hash: %x\n", p.ReplyTok)

			// C'est le moment d'envoyer un FindNode !
			// findNode := &v4wire.FindNode{Target: tonNodeID, Expiration: ...}

		case *v4wire.Neighbors:
			fmt.Printf("🌐 REÇU %d NOUVEAUX VOISINS de %s\n", len(p.Nodes), remoteAddr)
			for _, node := range p.Nodes {
				fmt.Printf("   -> Node ID: %x | IP: %s | Port TCP: %d\n",
					node.ID[:8], node.IP, node.TCP)
				// C'est ici que tu alimentes ta base de données de crawler !
			}

		case *v4wire.Ping:
			fmt.Printf("📥 Reçu PING de %s. On devrait lui répondre Pong !\n", remoteAddr)
			pong := buildPong(hash, toEndpoint)
			packet, _, err := v4wire.Encode(privKey, pong)
			if err != nil {
				fmt.Printf("Paquet malformé reçu de %s: %v\n", remoteAddr, err)
				continue
			}
			_, err = conn.WriteToUDP(packet, targetAddr)
			if err != nil {
				fmt.Println(err)
			}
			// Pour être un bon citoyen du réseau, tu devrais forger un Pong et l'envoyer.
		}
	}
}

func sendBackPong() {}

func buildPong(hash []byte, to v4wire.Endpoint) *v4wire.Pong {
	return &v4wire.Pong{
		To: to,
		// 2. ReplyTok: C'est le HASH du Ping que tu viens de recevoir
		// C'est ce qui lie ton Pong à son Ping.
		ReplyTok: hash,
		// 3. Expiration: Standard (20 minutes dans le futur)
		Expiration: uint64(time.Now().Add(20 * time.Minute).Unix()),
		// 4. ENRSeq: Optionnel (0 si tu n'as pas de record ENR)
		ENRSeq: 0,
	}
}
