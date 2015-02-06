// TODO: Create struct to use for channel between requestors of diameter messages
// and the receiver (the send message function)

// Copyright 2013-2015 go-diameter authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Diameter client example. This is by no means a complete client.
// The commands in here are not fully implemented. For that you have
// to read the RFCs (base and credit control) and follow the spec.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/plithner/go-dia/diam"
	"github.com/plithner/go-dia/diam/avp"
	"github.com/plithner/go-dia/diam/avp/format"
	"github.com/plithner/go-dia/diam/dict"
)

const (
	Identity    = format.DiameterIdentity("client")
	Realm       = format.DiameterIdentity("localhost")
	VendorId    = format.Unsigned32(13)
	ProductName = format.UTF8String("go-diameter")
)

// signaling data structure
type Sig struct {
	Code     string // Code of this message (i.e CER, DWR or CCR)
	msisdn   int
	Identity int // unique identifier (needed here... hmmm?)
}

func main() {
	ssl := flag.Bool("ssl", false, "connect using SSL/TLS")
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println("Use: client [-ssl] host:port")
		return
	}
	// Load the credit control dictionary on top of the base dictionary.
	dict.Default.Load(bytes.NewReader(dict.CreditControlXML))

	// ALL incoming messages are handled here.
	diam.HandleFunc("CEA", OnCEA)
	diam.HandleFunc("CCA", OnCCA)
	diam.HandleFunc("ALL", OnMSG) // Catch-all.

	// Connect using the default handler and base.Dict.
	addr := os.Args[len(os.Args)-1]
	log.Println("Connecting to", addr)
	var (
		c   diam.Conn
		err error
	)
	if *ssl {
		c, err = diam.DialTLS(addr, "", "", nil, nil)
	} else {
		c, err = diam.Dial(addr, nil, nil)
	}
	if err != nil {
		log.Fatal(err)
	}

	// This channel is a receive channel for the Composer and sender of the
	// DIAMETER message. CapabilityRequestStub, Watchdog and GenerateServiceRequest
	// all send to this "message bus" when they have something to commicate
	SigChannel := FanIn(CapabilityRequestStub(), Watchdog(), GenerateServiceRequest())

	// This is the go-routine responsible for communicating with the DIAMETER server
	// It is listening on the SigChannel for message requets
	// when something arrives, it contructs the appropriate DIAMETER
	// message and sends it to the DIAMETER server.
	go ComposeAndSendDiameterMessage(c, SigChannel)

	// Wait until the server kick us out.
	<-c.(diam.CloseNotifier).CloseNotify()
	log.Println("Server disconnected.")
}

// Simple handler for HTTP requests. Currently not used.
func HttpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
	log.Printf("handling... %s", r.URL.Path[1:])
}

// This one just passes the responsibility to listen for HTTP access
// to HttpHandler
func ClientListen(cnnl chan string) {
	http.HandleFunc("/", HttpHandler)
	http.ListenAndServe(":8080", nil)
}

// fan-in (multiplexing function) for the signaing channels
// takes the input from all the channels, and multiplexes it into one
// to enable unblocked reception of data on that channel
func FanIn(CerChannel, DwrChannel, CcrChannel <-chan Sig) <-chan Sig {
	channel := make(chan Sig)
	go func() {
		for {
			channel <- <-CerChannel
		}
	}()
	go func() {
		for {
			channel <- <-DwrChannel
		}
	}()
	go func() {
		for {
			channel <- <-CcrChannel
		}
	}()
	return channel
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

// This is a test-stub, that attempts to mimic the system that
// receives service requests from subscribers (end-users (people))
// and initiates a debit on their accounts
// It is a simple go routine that lives forever, and sends a service
// request now and then
func GenerateServiceRequest() <-chan Sig { // returns a receive only channel of string
	channel := make(chan Sig)

	go func() {
		time.Sleep(5 * time.Second)
		Identity := 0
		//msisdn := rand.Int()

		// Send CCR  every x*rand seconds
		for {
			// create random fake MSISDN
			msisdn := random(1000000, 4999999)
			msisdn = msisdn + 46702000000

			Identity = Identity + 1

			infoElem := Sig{
				Code:     "CCR",
				msisdn:   msisdn,
				Identity: Identity,
			}

			sleeptime := time.Second * time.Duration(rand.Intn(25))
			log.Printf("Time until next CCR: %s", sleeptime)
			time.Sleep(sleeptime)
			channel <- infoElem
		}
	}()
	return channel
}

// lives forever and sends DWR message at a fixed interval
func Watchdog() <-chan Sig { // returns a receive only channel of string
	channel := make(chan Sig)

	// Send watchdog messages every x seconds
	go func() {
		for {
			time.Sleep(10 * time.Second)
			infoElem := Sig{
				Code: "DWR",
			}

			channel <- infoElem
		}
	}()
	return channel
}

func CapabilityRequestStub() <-chan Sig { // returns a receive only channel of string
	// Send CER
	channel := make(chan Sig)

	go func() {
		infoElem := Sig{
			Code: "CER",
		}
		channel <- infoElem
	}()
	return channel
}

func ComposeAndSendDiameterMessage(c diam.Conn, SigChannel <-chan Sig) {
	// Listen forver on the SigChannel for requests to send messages to server
	for {
		msg := <-SigChannel //get whatever is on the channel

		if msg.Code == "CCR" {
			log.Printf("Internal message received: %s for %d with ID: %d", msg.Code, msg.msisdn, msg.Identity)

			// Craft a CCR message.
			r := diam.NewRequest(diam.CreditControl, 4, nil)
			r.NewAVP(avp.SessionId, avp.Mbit, 0, format.UTF8String("fake-session"))
			r.NewAVP(avp.OriginHost, avp.Mbit, 0, Identity)
			r.NewAVP(avp.OriginRealm, avp.Mbit, 0, Realm)
			r.NewAVP(avp.DestinationRealm, avp.Mbit, 0, Realm)
			//peerRealm, _ := m.FindAVP(avp.OriginRealm) // You should handle errors.
			//r.NewAVP(avp.DestinationRealm, avp.Mbit, 0, peerRealm.Data)
			r.NewAVP(avp.AuthApplicationId, avp.Mbit, 0, format.Unsigned32(4))
			r.NewAVP(avp.CCRequestType, avp.Mbit, 0, format.UTF8String("INITIAL_REQUEST"))
			r.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, format.Unsigned32(1))

			r.NewAVP(avp.SubscriptionId, avp.Mbit, 0, &diam.GroupedAVP{
				AVP: []*diam.AVP{
					diam.NewAVP(avp.SubscriptionIdType, avp.Mbit, 0, format.Enumerated(0)),
					diam.NewAVP(avp.SubscriptionIdData, avp.Mbit, 0, format.UTF8String("12345678")),
				},
			})

			// Add Service-Context-Id and all other AVPs..
			//r.WriteTo(c)
			log.Printf("Sending CCR message to %s", c.RemoteAddr().String())
			// Send message to the connection
			if _, err := r.WriteTo(c); err != nil {
				log.Fatal("Write failed:", err)
			}
		}

		if msg.Code == "DWR" {
			log.Printf("TODO: Send DWR")
			/*
				log.Printf("Internal message received: %s for %d with ID: %d", msg.Code, msg.msisdn, msg.Identity)

				d = diam.NewRequest(diam.DeviceWatchdogRequest, 0, nil)
				d.NewAVP(avp.OriginHost, avp.Mbit, 0, Identity)
				d.NewAVP(avp.OriginRealm, avp.Mbit, 0, Realm)
				d.NewAVP(avp.OriginStateId, avp.Mbit, 0, format.Unsigned32(rand.Uint32()))
				log.Printf("Sending message to %s", c.RemoteAddr().String())
				log.Println(d)
				if _, err := d.WriteTo(c); err != nil {
					log.Fatal("Write failed:", err)
				}
			*/
		}

		if msg.Code == "CER" {
			log.Printf("Internal message received: %s", msg.Code)

			// Create and send CER
			m := diam.NewRequest(diam.CapabilitiesExchange, 0, nil)
			m.NewAVP(avp.OriginHost, avp.Mbit, 0, Identity)
			m.NewAVP(avp.OriginRealm, avp.Mbit, 0, Realm)
			m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, Realm)
			laddr := c.LocalAddr()
			ip, _, _ := net.SplitHostPort(laddr.String())
			m.NewAVP(avp.HostIPAddress, avp.Mbit, 0, format.Address(net.ParseIP(ip)))
			m.NewAVP(avp.VendorId, avp.Mbit, 0, VendorId)
			m.NewAVP(avp.ProductName, avp.Mbit, 0, ProductName)

			m.NewAVP(avp.OriginStateId, avp.Mbit, 0, format.Unsigned32(rand.Uint32()))
			m.NewAVP(avp.VendorSpecificApplicationId, avp.Mbit, 0, &diam.GroupedAVP{
				AVP: []*diam.AVP{
					diam.NewAVP(avp.AuthApplicationId, avp.Mbit, 0, format.Unsigned32(4)),
					diam.NewAVP(avp.VendorId, avp.Mbit, 0, format.Unsigned32(10415)),
				},
			})
			log.Printf("Sending CER message to %s", c.RemoteAddr().String())
			//log.Println(m)
			// Send message to the connection
			if _, err := m.WriteTo(c); err != nil {
				log.Fatal("Write failed:", err)
			}
		}
	}
}

// OnCEA handles Capabilities-Exchange-Answer messages.
func OnCEA(c diam.Conn, m *diam.Message) {
	rc, err := m.FindAVP(avp.ResultCode)
	if err != nil {
		log.Fatal(err)
	}
	if v, _ := rc.Data.(format.Unsigned32); v != diam.Success {
		log.Fatal("Unexpected response:", rc)
	}

	log.Printf("CEA Received from %s", c.RemoteAddr().String())

	// TODO somhow communicated that the connection is up
	// and ready to send/receive messages

}

// OnCCA handles Credit-Control-Answer messages.
func OnCCA(c diam.Conn, m *diam.Message) {
	//log.Println(m)
	log.Println("CCA Received")
	// TODO: Communicate back to the sender, that a CCA has been received
	// and the status.
}

// OnMSG handles all other messages and just print them.
func OnMSG(c diam.Conn, m *diam.Message) {
	log.Printf("Receiving message from %s", c.RemoteAddr().String())
	//log.Println(m)
}
