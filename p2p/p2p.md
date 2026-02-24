# P2P 

## v5wire protocole 
### ancienement v4 pour le PoW 
maintenant surtout pour retrocompatibilité et execution layers 

Le v5 echange des clés valeurs pour connaitre les information sur les nodes 
Tout les communication sont chiffré 
Possibilité de chercher des noeuds spécifiques (qui stockent des blobs)

## Premiere étape de comunication Handshake v5
=> envoie d'un paquet random 
=> reception d'un paquet special WHOAREU avec un nonce 
=> le crawler (miniprojet) signe le nonce avec la clé privé et renvoie un paquet Handshake avec l'ENR signé 

```v5wire.Encode``` et ```v5wire.Decode```

```// packet Header
type Header struct {
	IV [sizeofMaskingIV]byte  
	StaticHeader
	AuthData []byte

	src enode.ID // used by decoder
}

// StaticHeader contains the static fields of a packet header.
type StaticHeader struct {
	ProtocolID [6]byte
	Version    uint16
	Flag       byte
	Nonce      Nonce
	AuthSize   uint16
}

// Authdata layouts.
type (
	whoareyouAuthData struct {
		IDNonce   [16]byte // ID proof data
		RecordSeq uint64   // highest known ENR sequence of requester
	}

	handshakeAuthData struct {
		h struct {
			SrcID      enode.ID
			SigSize    byte // signature data
			PubkeySize byte // offset of
		}
		// Trailing variable-size data.
		signature, pubkey, record []byte
	}

	messageAuthData struct {
		SrcID enode.ID
	}
)
```

Le Header du paquet n'est pas encodé mais masqué.


### ID des paquets

Structure des paquets 
- 32 bytes,  M.A.C., Le hash Keccak256 de tout ce qui suit (Signature + Type + Data).
- 65 bytes,  Signature , "Signature ECDSA (r, s, v) du contenu, générée avec ta privKey."
- 1 byte,    Packet Type, "Un seul octet qui définit si c'est un Ping (0x01), Pong (0x02), etc."
- Variable,Data (RLP),"Les données réelles du message, encodées au format RLP."


### Handshake du crawler 
envoie d'un message aleatoire de 44 octets minimum
=> reponse du noeud avec le nonce et WHOAREU paquet avec le nonce et ton ID 
=> pour t'authentifier tu renvoie un authHeader
	- nonce encrypté par la clé privé 
	- ENR (carte d'identité de ton noeud)
	- Ephemeral Public Key : Une clé publique temporaire générée juste pour cette session


### Le calcul de la Clé de Session (Diffie-Hellman)
Une fois que le nœud distant reçoit ton AuthHeader, vous faites tous les deux un calcul mathématique (le protocole Diffie-Hellman) :

Tu utilises ta clé privée éphémère + sa clé publique statique.
Il utilise sa clé privée statique + ta clé publique éphémère.
Résultat : Vous arrivez au même nombre secret sans jamais l'avoir envoyé sur le réseau.
Ce secret devient votre Clé de Session AES. À partir de cet instant, tous les paquets (Ping, FindNode) seront chiffrés avec cette clé.




### v4 ou v5 ? 

les nodes utilisent souvent les deux protocoles 
car v4 est utilisé pour l'Execution Layer et v5 pour le Consensus Layer
implementation dans ```p2p/server.go```
```
type Server struct {
    // ... other fields
    discv4 *discover.UDPv4
    discv5 *discover.UDPv5
    // ...
}
```