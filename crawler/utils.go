package main

import (
	"crypto/ecdsa"
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func EnodeExtracter(enodeStr string) (net.IP, int, *ecdsa.PublicKey, error) {
	// 1. Utiliser le parser officiel de Geth pour valider le format enode://
	node, err := enode.Parse(enode.ValidSchemes, enodeStr)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("invalid enode: %v", err)
	}

	// 2. Extraire l'IP et le port UDP (utilisé par discv5)
	ip := node.IP()
	port := node.UDP()

	// 3. Extraire la clé publique ECDSA du nœud
	// La clé est stockée dans le nœud sous forme de structure crypto.PublicKey
	pubKey := node.Pubkey()

	return ip, port, pubKey, nil
}

func PubKeyToNodeID(pub *ecdsa.PublicKey) [32]byte {
	// 1. Convertir la clé en bytes (format non-compressé)
	// Cette fonction retourne 65 octets : [0x04, X (32 bytes), Y (32 bytes)]
	pubBytes := crypto.FromECDSAPub(pub)

	// 2. Extraire uniquement les 64 octets bruts (on retire le préfixe 0x04)
	rawPubBytes := pubBytes[1:]

	// 3. Calculer le hash Keccak-256 des 64 octets
	hash := crypto.Keccak256(rawPubBytes)

	// 4. Convertir le slice []byte en tableau fixe [32]byte
	var nodeID [32]byte
	copy(nodeID[:], hash)

	return nodeID
}
