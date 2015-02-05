// Copyright 2013-2014 go-diameter authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package format

import "fmt"

// OctetString Diameter Format.
type OctetString string

func DecodeOctetString(b []byte) (Format, error) {
	return OctetString(b), nil
}

func (s OctetString) Serialize() []byte {
	return []byte(s)
}

func (s OctetString) Len() int {
	return len(s)
}

func (s OctetString) Padding() int {
	l := len(s)
	return pad4(l) - l
}

func (s OctetString) Format() FormatId {
	return OctetStringFormat
}

func (s OctetString) String() string {
	return fmt.Sprintf("OctetString{%s},Padding:%d", string(s), s.Padding())
}
