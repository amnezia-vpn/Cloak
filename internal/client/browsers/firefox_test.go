package browsers

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestComposeExtensions(t *testing.T) {
	target, _ := hex.DecodeString("000000170015000012636f6e73656e742e676f6f676c652e636f6d00170000ff01000100000a000e000c001d00170018001901000101000b00020100002300000010000e000c02683208687474702f312e310005000501000000000022000a000804030503060302030033006b0069001d00208d8ea1b80430b7710b65f0d89b0144a5eeb218709ce6613d4fc8bfb117657c1500170041947458330e3553dcde0a8741eb1dde26ebaee8262029c5edb3cbacc9ee1d7c866085b9cf483d943248997a65c5fa1d35725213895d0e5569d4e291863061b7d075002b00050403040303000d0018001604030503060308040805080604010501060102030201002d00020101001c0002400100150084000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")

	serverName := "consent.google.com"
	keyShare, _ := hex.DecodeString("8d8ea1b80430b7710b65f0d89b0144a5eeb218709ce6613d4fc8bfb117657c15")

	result := (&Firefox{}).composeExtensions(serverName, keyShare)
	// skip random secp256r1
	if !bytes.Equal(result[:151], target[:151]) || !bytes.Equal(result[216:], target[216:]) {
		t.Errorf("got %x", result)
	}
}