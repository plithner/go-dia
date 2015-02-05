// Copyright 2013-2014 go-diameter authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package diam

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/plithner/go-dia/diam/avp"
	"github.com/plithner/go-dia/diam/avp/format"
	"github.com/plithner/go-dia/diam/dict"
)

// testGroupedAVP is a Vendor-Specific-Application-Id Grouped AVP.
var testGroupedAVP = []byte{
	0x00, 0x00, 0x01, 0x04,
	0x40, 0x00, 0x00, 0x20,
	0x00, 0x00, 0x01, 0x02, // Auth-Application-Id
	0x40, 0x00, 0x00, 0x0c,
	0x00, 0x00, 0x00, 0x04,
	0x00, 0x00, 0x01, 0x0a, // Vendor-Id
	0x40, 0x00, 0x00, 0x0c,
	0x00, 0x00, 0x28, 0xaf,
}

func TestGroupedAVP(t *testing.T) {
	a, err := DecodeAVP(testGroupedAVP, 0, dict.Default)
	if err != nil {
		t.Fatal(err)
	}
	if a.Data.Format() != GroupedAVPFormat {
		t.Fatal("AVP is not grouped")
	}
	b, err := a.Serialize()
	if !bytes.Equal(b, testGroupedAVP) {
		t.Fatalf("Unexpected value.\nWant:\n%s\nHave:\n%s",
			hex.Dump(testGroupedAVP), hex.Dump(b))
	}
	t.Log(a)
}

func TestDecodeMessageWithGroupedAVP(t *testing.T) {
	m := NewRequest(257, 0, dict.Default)
	m.NewAVP(264, 0x40, 0, format.DiameterIdentity("client"))
	a, _ := DecodeAVP(testGroupedAVP, 0, dict.Default)
	m.AddAVP(a)
	t.Logf("Message:\n%s", m)
}

func TestMakeGroupedAVP(t *testing.T) {
	g := &GroupedAVP{
		AVP: []*AVP{
			NewAVP(avp.AuthApplicationId, avp.Mbit, 0, format.Unsigned32(4)),
			NewAVP(avp.VendorId, avp.Mbit, 0, format.Unsigned32(10415)),
		},
	}
	a := NewAVP(avp.VendorSpecificApplicationId, avp.Mbit, 0, g)
	b, err := a.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, testGroupedAVP) {
		t.Fatalf("Unexpected value.\nWant:\n%s\nHave:\n%s",
			hex.Dump(testGroupedAVP), hex.Dump(b))
	}
	t.Logf("Message:\n%s", a)
}
