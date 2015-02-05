// Copyright 2013-2014 go-diameter authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package format

import "fmt"

// DiameterIdentity Diameter Format.
type DiameterIdentity OctetString

func DecodeDiameterIdentity(b []byte) (Format, error) {
	return DiameterIdentity(OctetString(b)), nil
}

func (s DiameterIdentity) Serialize() []byte {
	return OctetString(s).Serialize()
}

func (s DiameterIdentity) Len() int {
	return len(s)
}

func (s DiameterIdentity) Padding() int {
	l := len(s)
	return pad4(l) - l
}

func (s DiameterIdentity) Format() FormatId {
	return DiameterIdentityFormat
}

func (s DiameterIdentity) String() string {
	return fmt.Sprintf("DiameterIdentity{%s},Padding:%d", string(s), s.Padding())
}
