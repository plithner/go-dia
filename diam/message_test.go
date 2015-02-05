// Copyright 2013-2014 go-diameter authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package diam

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"net"
	"testing"

	"github.com/plithner/go-dia/diam/avp"
	"github.com/plithner/go-dia/diam/avp/format"
	"github.com/plithner/go-dia/diam/dict"
)

// testMessage is used by the test cases below and also in reflect_test.go.
// The same testMessage is re-created programatically in TestNewMessage.
//
// Capabilities-Exchange-Request (CER)
// {Code:257,Flags:0x80,Version:0x1,Length:204,ApplicationId:0,HopByHopId:0xa8cc407d,EndToEndId:0xa8c1b2b4}
//   Origin-Host {Code:264,Flags:0x40,Length:12,VendorId:0,Value:DiameterIdentity{test},Padding:0}
//   Origin-Realm {Code:296,Flags:0x40,Length:20,VendorId:0,Value:DiameterIdentity{localhost},Padding:3}
//   Host-IP-Address {Code:257,Flags:0x40,Length:16,VendorId:0,Value:Address{10.1.0.1},Padding:2}
//   Vendor-Id {Code:266,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{13}}
//   Product-Name {Code:269,Flags:0x0,Length:20,VendorId:0,Value:UTF8String{go-diameter},Padding:1}
//   Origin-State-Id {Code:278,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{1397760650}}
//   Supported-Vendor-Id {Code:265,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{10415}}
//   Supported-Vendor-Id {Code:265,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{13}}
//   Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{4}}
//   Inband-Security-Id {Code:299,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{0}}
//   Vendor-Specific-Application-Id {Code:260,Flags:0x40,Length:32,VendorId:0,Value:Grouped{
//     Auth-Application-Id {Code:258,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{4}},
//     Vendor-Id {Code:266,Flags:0x40,Length:12,VendorId:0,Value:Unsigned32{10415}},
//   }}
//   Firmware-Revision {Code:267,Flags:0x0,Length:12,VendorId:0,Value:Unsigned32{1}}
var testMessage = []byte{
	0x01, 0x00, 0x00, 0xcc,
	0x80, 0x00, 0x01, 0x01,
	0x00, 0x00, 0x00, 0x00,
	0xa8, 0xcc, 0x40, 0x7d,
	0xa8, 0xc1, 0xb2, 0xb4,
	0x00, 0x00, 0x01, 0x08,
	0x40, 0x00, 0x00, 0x0c,
	0x74, 0x65, 0x73, 0x74,
	0x00, 0x00, 0x01, 0x28,
	0x40, 0x00, 0x00, 0x11,
	0x6c, 0x6f, 0x63, 0x61,
	0x6c, 0x68, 0x6f, 0x73,
	0x74, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x01,
	0x40, 0x00, 0x00, 0x0e,
	0x00, 0x01, 0x0a, 0x01,
	0x00, 0x01, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x0a,
	0x40, 0x00, 0x00, 0x0c,
	0x00, 0x00, 0x00, 0x0d,
	0x00, 0x00, 0x01, 0x0d,
	0x00, 0x00, 0x00, 0x13,
	0x67, 0x6f, 0x2d, 0x64,
	0x69, 0x61, 0x6d, 0x65,
	0x74, 0x65, 0x72, 0x00,
	0x00, 0x00, 0x01, 0x16,
	0x40, 0x00, 0x00, 0x0c,
	0x53, 0x50, 0x22, 0x8a,
	0x00, 0x00, 0x01, 0x09,
	0x40, 0x00, 0x00, 0x0c,
	0x00, 0x00, 0x28, 0xaf,
	0x00, 0x00, 0x01, 0x09,
	0x40, 0x00, 0x00, 0x0c,
	0x00, 0x00, 0x00, 0x0d,
	0x00, 0x00, 0x01, 0x02,
	0x40, 0x00, 0x00, 0x0c,
	0x00, 0x00, 0x00, 0x04,
	0x00, 0x00, 0x01, 0x2b,
	0x40, 0x00, 0x00, 0x0c,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x04,
	0x40, 0x00, 0x00, 0x20,
	0x00, 0x00, 0x01, 0x02,
	0x40, 0x00, 0x00, 0x0c,
	0x00, 0x00, 0x00, 0x04,
	0x00, 0x00, 0x01, 0x0a,
	0x40, 0x00, 0x00, 0x0c,
	0x00, 0x00, 0x28, 0xaf,
	0x00, 0x00, 0x01, 0x0b,
	0x00, 0x00, 0x00, 0x0c,
	0x00, 0x00, 0x00, 0x01,
}

func TestReadMessage(t *testing.T) {
	msg, err := ReadMessage(bytes.NewReader(testMessage), dict.Default)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Message:\n%s", msg)
}

func TestNewMessage(t *testing.T) {
	want, _ := ReadMessage(bytes.NewReader(testMessage), dict.Default)
	m := NewMessage(CapabilitiesExchange, RequestFlag, 0, 0xa8cc407d, 0xa8c1b2b4, dict.Default)
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, format.DiameterIdentity("test"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, format.DiameterIdentity("localhost"))
	m.NewAVP(avp.HostIPAddress, avp.Mbit, 0, format.Address(net.ParseIP("10.1.0.1")))
	m.NewAVP(avp.VendorId, avp.Mbit, 0, format.Unsigned32(13))
	m.NewAVP(avp.ProductName, 0, 0, format.UTF8String("go-diameter"))
	m.NewAVP(avp.OriginStateId, avp.Mbit, 0, format.Unsigned32(1397760650))
	m.NewAVP(avp.SupportedVendorId, avp.Mbit, 0, format.Unsigned32(10415))
	m.NewAVP(avp.SupportedVendorId, avp.Mbit, 0, format.Unsigned32(13))
	m.NewAVP(avp.AuthApplicationId, avp.Mbit, 0, format.Unsigned32(4))
	m.NewAVP(avp.InbandSecurityId, avp.Mbit, 0, format.Unsigned32(0))
	m.NewAVP(avp.VendorSpecificApplicationId, avp.Mbit, 0, &GroupedAVP{
		AVP: []*AVP{
			NewAVP(avp.AuthApplicationId, avp.Mbit, 0, format.Unsigned32(4)),
			NewAVP(avp.VendorId, avp.Mbit, 0, format.Unsigned32(10415)),
		},
	})
	m.NewAVP(avp.FirmwareRevision, 0, 0, format.Unsigned32(1))
	if m.Len() != want.Len() {
		t.Fatalf("Unexpected message length.\nWant: %d\n%s\nHave: %d\n%s",
			want.Len(), want, m.Len(), m)
	}
	a, err := m.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := want.Serialize()
	if !bytes.Equal(a, b) {
		t.Fatalf("Unexpected message.\nWant:\n%s\n%s\nHave:\n%s\n%s",
			want, hex.Dump(b), m, hex.Dump(a))
	}
	t.Logf("%d bytes\n%s", len(a), m)
	t.Logf("Message:\n%s", hex.Dump(a))
}

func TestMessageFindAVP(t *testing.T) {
	m, _ := ReadMessage(bytes.NewReader(testMessage), dict.Default)
	a, err := m.FindAVP(avp.OriginStateId)
	if err != nil {
		t.Fatal(err)
	}
	a, err = m.FindAVP("Origin-State-Id")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(a)
}

func BenchmarkReadMessage(b *testing.B) {
	reader := bytes.NewReader(testMessage)
	for n := 0; n < b.N; n++ {
		ReadMessage(reader, dict.Default)
		reader.Seek(0, 0)
	}
}

func BenchmarkWriteMessage(b *testing.B) {
	m, _ := ReadMessage(bytes.NewReader(testMessage), dict.Default)
	for n := 0; n < b.N; n++ {
		m.WriteTo(ioutil.Discard)
	}
}
