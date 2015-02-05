// Copyright 2013-2014 go-diameter authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package format

import (
	"encoding/binary"
	"fmt"
)

// Integer64 Diameter Format.
type Integer64 int64

func DecodeInteger64(b []byte) (Format, error) {
	return Integer64(binary.BigEndian.Uint64(b)), nil
}

func (n Integer64) Serialize() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	return b
}

func (n Integer64) Len() int {
	return 4
}

func (n Integer64) Padding() int {
	return 0
}

func (n Integer64) Format() FormatId {
	return Integer64Format
}

func (n Integer64) String() string {
	return fmt.Sprintf("Integer64{%d}", n)
}
