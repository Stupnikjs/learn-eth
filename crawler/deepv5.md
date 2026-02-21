# DEEP DIVE IN v5 

## CODEC 

Init codec 
```
    // NewCodec creates a wire codec.
    func NewCodec(ln *enode.LocalNode, key *ecdsa.PrivateKey, clock mclock.Clock, protocolID *[6]byte) *Codec {
        c := &Codec{
            sha256:     sha256.New(),
            localnode:  ln,
            privkey:    key,
            sc:         NewSessionCache(1024, clock),
            protocolID: DefaultProtocolID,              // discv5
            decbuf:     make([]byte, maxPacketSize),
        }
        if protocolID != nil {
            c.protocolID = *protocolID
        }
        return c
    }
```



Encoding func 
```
func (c *Codec) Encode(id enode.ID, addr string, packet Packet, challenge *Whoareyou) ([]byte, Nonce, error) {
	// Create the packet header.
	var (
		head    Header
		session *session
		msgData []byte
		err     error
	)
	switch {
	case packet.Kind() == WhoareyouPacket:
		// just send the WHOAREYOU packet raw again, rather than the re-encoded challenge data
		w := packet.(*Whoareyou)
		if len(w.Encoded) > 0 {
			return w.Encoded, w.Nonce, nil
		}
		head, err = c.encodeWhoareyou(id, packet.(*Whoareyou))
	case challenge != nil:
		// We have an unanswered challenge, send handshake.
		head, session, err = c.encodeHandshakeHeader(id, addr, challenge)
	default:
		session = c.sc.session(id, addr)
		if session != nil {
			// There is a session, use it.
			head, err = c.encodeMessageHeader(id, session)
		} else {
			// No keys, send random data to kick off the handshake.
			head, msgData, err = c.encodeRandom(id)
		}
	}
	if err != nil {
		return nil, Nonce{}, err
	}

	// Generate masking IV.
	if err := c.sc.maskingIVGen(head.IV[:]); err != nil {
		return nil, Nonce{}, fmt.Errorf("can't generate masking IV: %v", err)
	}

	// Encode header data.
	c.writeHeaders(&head)

	// Store sent WHOAREYOU challenges.
	if challenge, ok := packet.(*Whoareyou); ok {
		challenge.ChallengeData = slices.Clone(c.buf.Bytes())
		enc, err := c.EncodeRaw(id, head, msgData)
		if err != nil {
			return nil, Nonce{}, err
		}
		challenge.Encoded = bytes.Clone(enc)
		c.sc.storeSentHandshake(id, addr, challenge)
		return enc, head.Nonce, err
	}

	if msgData == nil {
		headerData := c.buf.Bytes()
		msgData, err = c.encryptMessage(session, packet, &head, headerData)
		if err != nil {
			return nil, Nonce{}, err
		}
	}
	enc, err := c.EncodeRaw(id, head, msgData)
	return enc, head.Nonce, err
}


```
dans le premier contact on a ces call 

```
head, msgData, err = c.encodeRandom(id)
err := c.sc.maskingIVGen(head.IV[:])
c.writeHeaders(&head)
headerData := c.buf.Bytes()
msgData, err = c.encryptMessage(session, packet, &head, headerData)
enc, err := c.EncodeRaw(id, head, msgData)
return enc, head.Nonce, err
```

enc c'est le paquet encodé 
