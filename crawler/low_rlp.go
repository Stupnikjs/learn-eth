package main

import "encoding/binary"

func encodeStaticTag(tag [32]byte) []byte {
	// Une string de 32 octets est "courte" (< 55)
	// Préfixe = 0x80 + 32 = 0xA0
	rlp := make([]byte, 0, 33)
	rlp = append(rlp, 0xa0)
	rlp = append(rlp, tag[:]...)
	return rlp
}

func encodeList(elements ...[]byte) []byte {
	var body []byte
	for _, e := range elements {
		body = append(body, e...)
	}

	n := len(body)
	if n <= 55 {
		// Cas court : 0xc0 + longueur
		header := []byte{0xc0 + byte(n)}
		return append(header, body...)
	}

	// Cas long (> 55 octets) :
	// 1. On calcule combien d'octets il faut pour écrire le nombre 'n'
	lengthInBytes := encodeLength(n)
	// 2. Le préfixe est 0xf7 + nombre d'octets de la longueur
	// donc premier byte au moins 0xF8 car longeur de la liste minimum sur 1 byte
	firstByte := 0xf7 + byte(len(lengthInBytes))

	header := append([]byte{firstByte}, lengthInBytes...)
	return append(header, body...)
}

func encodeLength(n int) []byte {
	// 1. On crée un buffer de 8 octets (taille max d'un int64)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n))

	// 2. Le RLP interdit les zéros de tête (ex: 500 doit être 0x01F4, pas 0x000001F4)
	// On cherche donc le premier octet non nul
	firstNonZero := 0
	for firstNonZero < len(buf) && buf[firstNonZero] == 0 {
		firstNonZero++
	}

	// 3. On retourne uniquement la partie significative
	return buf[firstNonZero:]
}
