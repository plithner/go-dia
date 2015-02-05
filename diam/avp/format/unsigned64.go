// Copyright 2013-2014 go-diameter authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package format

import (
	"encoding/binary"
	"fmt"
)

// Unsigned64 Diameter Format.
type Unsigned64 uint64

func DecodeUnsigned64(b []byte) (Format, error) {
	return Unsigned64(binary.BigEndian.Uint64(b)), nil
}

func (n Unsigned64) Serialize() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	return b
}

func (n Unsigned64) Len() int {
	return 8
}

func (n Unsigned64) Padding() int {
	return 0
}

func (n Unsigned64) Format() FormatId {
	return Unsigned64Format
}

func (n Unsigned64) String() string {
	return fmt.Sprintf("Unsigned64{%d}", n)
}
