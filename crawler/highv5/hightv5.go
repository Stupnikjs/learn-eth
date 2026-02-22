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

	send(conn, codec, node, &v5wire.Unknown{}, nil)

	// 8. Boucle de lecture pour capturer le WHOAREYOU
	buf := make([]byte, 1280)

	for {

		conn.SetReadDeadline(time.Now().Add(20 * time.Second))

		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalf("Aucune réponse (Timeout): %v", err)
		}

		// 9. Décodage du paquet reçu
		_, _, respPacket, err := codec.Decode(buf[:n], addr.String())
		if err != nil {
			log.Fatalf("Erreur Décodage réponse: %v", err)
		}

		switch p := respPacket.(type) {
		case *v5wire.Whoareyou:
			p.Node = node
			ping := v5wire.Ping{
				ReqID:  []byte("h"),
				ENRSeq: localNode.Seq(),
			}
			send(conn, codec, node, &ping, p)
		case *v5wire.Pong:
			findnode := v5wire.Findnode{
				ReqID:     []byte("michel"),
				Distances: []uint{0, 0},
			}
			send(conn, codec, node, &findnode, nil)
		case *v5wire.Nodes:
			fmt.Printf("🌐 Reçu %d nœuds voisins !\n", len(p.Nodes))
			for i, enr := range p.Nodes {
				// On décode l'ENR pour voir l'IP/Port des voisins
				fmt.Printf("  [%d]| ENR Seq: %d\n", i, enr.Seq())
			}
		default:
			fmt.Println(p.Name())
		}

	}

}

func send(conn *net.UDPConn, codec *v5wire.Codec, node *enode.Node, packet v5wire.Packet, challenge *v5wire.Whoareyou) {
	addr := &net.UDPAddr{IP: node.IP(), Port: node.UDP()}
	enc, _, err := codec.Encode(node.ID(), addr.String(), packet, challenge)
	if err != nil {
		log.Printf("Erreur encodage (%v): %v", packet.Kind(), err)
		return
	}
	conn.WriteToUDP(enc, addr)
	fmt.Printf(">> Paquet %v envoyé à %s\n", packet.Kind(), addr)
}
