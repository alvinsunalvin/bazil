package edtls_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"testing"
	"time"

	"bytes"

	"bazil.org/bazil/util/edtls"
	"github.com/agl/ed25519"
)

type fakeRand struct{}

func (fakeRand) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		p[i] = 0x42
	}
	return len(p), nil
}

func TestVouch(t *testing.T) {
	signPub, signPriv, err := ed25519.GenerateKey(fakeRand{})
	if err != nil {
		t.Fatalf("unexpected error from ed25519.GenerateKey: %v", err)
	}
	tlsPriv, err := ecdsa.GenerateKey(elliptic.P256(), fakeRand{})
	if err != nil {
		t.Fatalf("unexpected error from ecdsa.GenerateKey: %v", err)
	}
	cert := x509.Certificate{
		// most fields don't matter

		NotAfter: time.Date(2014, 12, 13, 14, 15, 16, 17, time.UTC),
		ExtraExtensions: []pkix.Extension{
			{Id: asn1.ObjectIdentifier{42}, Value: []byte("filler")},
		},
	}
	if err := edtls.Vouch(signPub, signPriv, &cert, &tlsPriv.PublicKey); err != nil {
		t.Fatalf("unexpected error from Vouch: %v", err)
	}
	if ext := cert.ExtraExtensions; len(ext) != 2 {
		t.Fatalf("unexpected ExtraExtensions: %#v", ext)
	}
	vouch := cert.ExtraExtensions[1]
	if g, e := vouch.Id.String(), "1.2.840.113556.1.8000.2554.31830.5190.18203.20240.41147.7688498.2373901"; g != e {
		t.Errorf("unexpected oid for vouch: %q != %q", g, e)
	}
	if g, e := vouch.Critical, false; g != e {
		t.Errorf("unexpected critical flag: %v != %v", g, e)
	}
	if g, e := len(vouch.Value), ed25519.PublicKeySize+ed25519.SignatureSize; g != e {
		t.Errorf("unexpected signature length: %v != %v", g, e)
	}
	if g, e := vouch.Value, []byte{
		// pubkey
		0x21, 0x52, 0xf8, 0xd1, 0x9b, 0x79, 0x1d, 0x24,
		0x45, 0x32, 0x42, 0xe1, 0x5f, 0x2e, 0xab, 0x6c,
		0xb7, 0xcf, 0xfa, 0x7b, 0x6a, 0x5e, 0xd3, 0x00,
		0x97, 0x96, 0x0e, 0x06, 0x98, 0x81, 0xdb, 0x12,
		// sig
		0x81, 0x1b, 0xf8, 0xef, 0x30, 0xf3, 0x2b, 0x6d,
		0x5f, 0x41, 0xa2, 0x04, 0xff, 0x5a, 0xb5, 0xd0,
		0xb1, 0xfc, 0x60, 0xf0, 0x79, 0x10, 0x57, 0xa8,
		0x42, 0x0f, 0x16, 0xe4, 0x57, 0xc3, 0xe4, 0x60,
		0x28, 0x29, 0xf2, 0xfd, 0x06, 0x67, 0x7c, 0x3a,
		0xc0, 0xcd, 0x85, 0xda, 0x0e, 0xf4, 0x8c, 0x1b,
		0xc0, 0xed, 0xf8, 0xd3, 0x10, 0xa7, 0xa0, 0x4e,
		0x34, 0xed, 0x83, 0x57, 0x41, 0x49, 0xbe, 0x06,
	}; !bytes.Equal(g, e) {
		t.Errorf("unexpected signature packet: %x != %x", g, e)
	}
}

func TestVerify(t *testing.T) {
	signPub, _, err := ed25519.GenerateKey(fakeRand{})
	if err != nil {
		t.Fatalf("unexpected error from ed25519.GenerateKey: %v", err)
	}
	tlsPriv, err := ecdsa.GenerateKey(elliptic.P256(), fakeRand{})
	if err != nil {
		t.Fatalf("unexpected error from ecdsa.GenerateKey: %v", err)
	}
	cert := x509.Certificate{
		// most fields don't matter

		PublicKey: &tlsPriv.PublicKey,
		NotAfter:  time.Date(2014, 12, 13, 14, 15, 16, 17, time.UTC),
		Extensions: []pkix.Extension{
			{
				Id: asn1.ObjectIdentifier{1, 2, 840, 113556, 1, 8000, 2554, 31830, 5190, 18203, 20240, 41147, 7688498, 2373901},
				Value: []byte{
					// pubkey
					0x21, 0x52, 0xf8, 0xd1, 0x9b, 0x79, 0x1d, 0x24,
					0x45, 0x32, 0x42, 0xe1, 0x5f, 0x2e, 0xab, 0x6c,
					0xb7, 0xcf, 0xfa, 0x7b, 0x6a, 0x5e, 0xd3, 0x00,
					0x97, 0x96, 0x0e, 0x06, 0x98, 0x81, 0xdb, 0x12,
					// sig
					0x81, 0x1b, 0xf8, 0xef, 0x30, 0xf3, 0x2b, 0x6d,
					0x5f, 0x41, 0xa2, 0x04, 0xff, 0x5a, 0xb5, 0xd0,
					0xb1, 0xfc, 0x60, 0xf0, 0x79, 0x10, 0x57, 0xa8,
					0x42, 0x0f, 0x16, 0xe4, 0x57, 0xc3, 0xe4, 0x60,
					0x28, 0x29, 0xf2, 0xfd, 0x06, 0x67, 0x7c, 0x3a,
					0xc0, 0xcd, 0x85, 0xda, 0x0e, 0xf4, 0x8c, 0x1b,
					0xc0, 0xed, 0xf8, 0xd3, 0x10, 0xa7, 0xa0, 0x4e,
					0x34, 0xed, 0x83, 0x57, 0x41, 0x49, 0xbe, 0x06,
				},
			},
		},
	}
	peerPub, ok := edtls.Verify(&cert)
	if !ok {
		t.Fatalf("unexpected negative result from Verify")
	}
	if g, e := peerPub, signPub; *g != *e {
		t.Fatalf("unexpected result from Verify: %v != %v", g, e)
	}
}

func TestVerifyMissing(t *testing.T) {
	tlsPriv, err := ecdsa.GenerateKey(elliptic.P256(), fakeRand{})
	if err != nil {
		t.Fatalf("unexpected error from ecdsa.GenerateKey: %v", err)
	}
	cert := x509.Certificate{
		// most fields don't matter

		PublicKey: &tlsPriv.PublicKey,
		NotAfter:  time.Date(2014, 12, 13, 14, 15, 16, 17, time.UTC),
	}
	_, ok := edtls.Verify(&cert)
	if ok {
		t.Fatalf("unexpected positive from Verify")
	}
}

func TestVerifyBad(t *testing.T) {
	tlsPriv, err := ecdsa.GenerateKey(elliptic.P256(), fakeRand{})
	if err != nil {
		t.Fatalf("unexpected error from ecdsa.GenerateKey: %v", err)
	}
	cert := x509.Certificate{
		// most fields don't matter

		PublicKey: &tlsPriv.PublicKey,
		NotAfter:  time.Date(2014, 12, 13, 14, 15, 16, 17, time.UTC),
		Extensions: []pkix.Extension{
			{
				Id: asn1.ObjectIdentifier{1, 2, 840, 113556, 1, 8000, 2554, 31830, 5190, 18203, 20240, 41147, 7688498, 2373901},
				Value: []byte{
					// pubkey
					0x21, 0x52, 0xf8, 0xd1, 0x9b, 0x79, 0x1d, 0x24,
					0x45, 0x32, 0x42, 0xe1, 0x5f, 0x2e, 0xab, 0x6c,
					0xb7, 0xcf, 0xfa, 0x7b, 0x6a, 0x5e, 0xd3, 0x00,
					0x97, 0x96, 0x0e, 0x06, 0x98, 0x81, 0xdb, 0x12,
					// sig
					0x81, 0x1b, 0xf8, 0xef, 0x30, 0xf3, 0x2b, 0x6d,
					0x5f, 0x41, 0xa2, 0x04, 0xff, 0x5a, 0xb5, 0xd0,
					0xb1, 0xfc, 0x60, 0xf0, 0x79, 0x10, 0x57, 0xa8,
					0x42, 0x0f, 0x16, 0xe4, 0x57, 0xc3, 0xe4, 0x60,
					0x28, 0x29, 0xf2, 0xfd, 0x06, 0x67, 0x7c, 0x3a,
					0xc0, 0xcd, 0x85, 0xda, 0x0e, 0xf4, 0x8c, 0x1b,
					0xc0, 0xed, 0xf8, 0xd3, 0x10, 0xa7, 0xa0, 0x4e,
					0x34, 0xed, 0x83, 0x57, 0x41, 0x49, 0xbe, 0x07, /* evil here */
				},
			},
		},
	}
	_, ok := edtls.Verify(&cert)
	if ok {
		t.Fatalf("unexpected positive from Verify")
	}
}
