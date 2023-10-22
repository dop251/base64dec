//go:generate go run generate_decode_map.go

// Package base64dec contains a universal base64 decoder that works on both the standard and url-safe variants, padded and raw.
// The code is based on the standard encoding/base64 package.
package base64dec

import (
	"encoding/base64"
	"encoding/binary"
	"strconv"
)

const padChar = '='

type ByteSeq interface {
	[]byte | string
}

// DecodeBase64 decodes src and writes at most base64.RawStdEncoding.DecodedLen(len(src))
// bytes to dst and returns the number of bytes written. If src contains invalid base64 data, it will return the
// number of bytes successfully written and base64.CorruptInputError.
// New line characters (\r and \n) are ignored.
// The input can be in the standard or the alternate (aka url-safe) encoding. It can be padded or un-padded.
// If there is a correct padding, it is consumed and no error is returned. If there is no padding where it's required,
// no error is returned. If there is an incorrect padding (i.e. too many or too few characters) it is treated
// as garbage at the end (i.e. the error will point to the first padding character).
func DecodeBase64[T ByteSeq](dst []byte, src T) (n int, err error) {
	if len(src) == 0 {
		return 0, nil
	}

	si := 0
	for strconv.IntSize >= 64 && len(src)-si >= 8 && len(dst)-n >= 8 {
		src2 := src[si : si+8]
		if dn, ok := assemble64(
			decodeMap[src2[0]],
			decodeMap[src2[1]],
			decodeMap[src2[2]],
			decodeMap[src2[3]],
			decodeMap[src2[4]],
			decodeMap[src2[5]],
			decodeMap[src2[6]],
			decodeMap[src2[7]],
		); ok {
			binary.BigEndian.PutUint64(dst[n:], dn)
			n += 6
			si += 8
		} else {
			var ninc int
			si, ninc, err = decodeQuantum(dst[n:], src, si)
			n += ninc
			if err != nil {
				return n, err
			}
		}
	}

	for len(src)-si >= 4 && len(dst)-n >= 4 {
		src2 := src[si : si+4]
		if dn, ok := assemble32(
			decodeMap[src2[0]],
			decodeMap[src2[1]],
			decodeMap[src2[2]],
			decodeMap[src2[3]],
		); ok {
			binary.BigEndian.PutUint32(dst[n:], dn)
			n += 3
			si += 4
		} else {
			var ninc int
			si, ninc, err = decodeQuantum(dst[n:], src, si)
			n += ninc
			if err != nil {
				return n, err
			}
		}
	}

	for si < len(src) {
		var ninc int
		si, ninc, err = decodeQuantum(dst[n:], src, si)
		n += ninc
		if err != nil {
			return n, err
		}
	}
	return n, err
}

// assemble32 assembles 4 base64 digits into 3 bytes.
// Each digit comes from the decode map, and will be 0xff
// if it came from an invalid character.
func assemble32(n1, n2, n3, n4 byte) (dn uint32, ok bool) {
	// Check that all the digits are valid. If any of them was 0xff, their
	// bitwise OR will be 0xff.
	if n1|n2|n3|n4 == 0xff {
		return 0, false
	}
	return uint32(n1)<<26 |
			uint32(n2)<<20 |
			uint32(n3)<<14 |
			uint32(n4)<<8,
		true
}

// assemble64 assembles 8 base64 digits into 6 bytes.
// Each digit comes from the decode map, and will be 0xff
// if it came from an invalid character.
func assemble64(n1, n2, n3, n4, n5, n6, n7, n8 byte) (dn uint64, ok bool) {
	// Check that all the digits are valid. If any of them was 0xff, their
	// bitwise OR will be 0xff.
	if n1|n2|n3|n4|n5|n6|n7|n8 == 0xff {
		return 0, false
	}
	return uint64(n1)<<58 |
			uint64(n2)<<52 |
			uint64(n3)<<46 |
			uint64(n4)<<40 |
			uint64(n5)<<34 |
			uint64(n6)<<28 |
			uint64(n7)<<22 |
			uint64(n8)<<16,
		true
}

// decodeQuantum decodes up to 4 base64 bytes. The received parameters are
// the destination buffer dst, the source buffer src and an index in the
// source buffer si.
// It returns the number of bytes read from src, the number of bytes written
// to dst, and an error, if any.
func decodeQuantum[T ByteSeq](dst []byte, src T, si int) (nsi, n int, err error) {
	// Decode quantum using the base64 alphabet
	var dbuf [4]byte
	dlen := 4

	for j := 0; j < len(dbuf); j++ {
		if len(src) == si {
			if j == 0 {
				return si, 0, nil
			}
			dlen = j
			break
		}
		in := src[si]
		si++

		out := decodeMap[in]
		if out != 0xff {
			dbuf[j] = out
			continue
		}

		if in == '\n' || in == '\r' {
			j--
			continue
		}

		dlen = j

		if rune(in) != padChar {
			err = base64.CorruptInputError(si - 1)
			break
		}

		// We've reached the end and there's padding
		switch j {
		case 0, 1:
			// incorrect padding
			err = base64.CorruptInputError(si - 1)
		case 2:
			// "==" is expected, the first "=" is already consumed.
			// skip over newlines
			for si < len(src) && (src[si] == '\n' || src[si] == '\r') {
				si++
			}
			if si == len(src) {
				// not enough padding
				err = base64.CorruptInputError(si - 1)
				break
			} else if rune(src[si]) != padChar {
				// incorrect padding
				err = base64.CorruptInputError(si - 1)
				break
			}

			si++
		}

		if err == nil {
			// skip over newlines
			for si < len(src) && (src[si] == '\n' || src[si] == '\r') {
				si++
			}
			if si < len(src) {
				// trailing garbage
				err = base64.CorruptInputError(si)
			}
		}
		break
	}

	if dlen == 0 {
		return si, 0, err
	}

	// Convert 4x 6bit source bytes into 3 bytes
	val := uint(dbuf[0])<<18 | uint(dbuf[1])<<12 | uint(dbuf[2])<<6 | uint(dbuf[3])
	dbuf[2], dbuf[1], dbuf[0] = byte(val>>0), byte(val>>8), byte(val>>16)
	switch dlen {
	case 4:
		dst[2] = dbuf[2]
		dbuf[2] = 0
		fallthrough
	case 3:
		dst[1] = dbuf[1]
		dbuf[1] = 0
		fallthrough
	case 2:
		dst[0] = dbuf[0]
	}

	return si, dlen - 1, err
}
