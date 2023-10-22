package crypto_test

import (
	"testing"

	"github.com/aftermath2/BTRY/crypto"

	"github.com/stretchr/testify/assert"
)

func TestValidatePublicKey(t *testing.T) {
	testCases := []struct {
		desc      string
		publicKey string
		fail      bool
	}{
		{
			desc:      "Valid",
			publicKey: "345fe256754b1b472e58aede6c2f138ce67d05d431c776bcb4e384edbbdca9cd",
		},
		{
			desc:      "Invalid length",
			publicKey: "public key",
			fail:      true,
		},
		{
			desc:      "Invalid encoding",
			publicKey: "pubfe256754b1b472e58aede6c2f138ce67d05d431c776bcb4e384edbbdca9cd",
			fail:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := crypto.ValidatePublicKey(tc.publicKey)
			if tc.fail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifySignature(t *testing.T) {
	cases := []struct {
		desc      string
		publicKey string
		signature string
		fail      bool
	}{
		{
			desc:      "Valid pair",
			publicKey: "e68b99fc5f60c971926fdc3a3af38ccf67e6f4306ab1c388735533e7c5dcc749",
			signature: "b347d0accfdc0e9446921c9d61699432826c4d552201d35f5e81482ea0d7d8ecbc99e10c0c590858d44389a49bccd20c7c2f414f88230d418a2be43157336005",
		},
		{
			desc:      "Invalid public key encoding",
			publicKey: "publicKey",
			fail:      true,
		},
		{
			desc:      "Invalid signature encoding",
			publicKey: "f73a68a323bca465ad2dd2cab8df07d5bb854984277d6354e3c16764e9ddd079",
			signature: "signature",
			fail:      true,
		},
		{
			desc:      "Invalid pair",
			publicKey: "470541ae525b58f98160e5d85a20697e16096020b833b84fb394e0099c874736",
			signature: "df26b9f6e916a57a2493dd18230a7d3071ade982eae2d40e41ae328a0ea9c1654395a2d8cceef9927c60a9bc8dfdf319e30f7945660fb84740c6155b8855f60c",
			fail:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := crypto.VerifySignature(tc.publicKey, tc.signature)
			if tc.fail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
