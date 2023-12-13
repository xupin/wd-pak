package lzss

import "bytes"

const (
	// n is the size of ring buffer - must be power of 2
	n = 4096
	// f is the upper limit for match_length
	f = 18
	// threshold encode string into position and length if match_length is greater than this
	threshold = 2
	padding   = 0x16c
)

// Decompress decompresses lzss data
func Decompress(src []byte) []byte {

	var i, j, r, c int
	var flags uint

	srcBuf := bytes.NewBuffer(src)
	dst := bytes.Buffer{}

	// ring buffer of size n, with extra f-1 bytes to aid string comparison
	textBuf := make([]byte, n+f-1)

	r = n - f
	flags = 0

	for {
		flags = flags >> 1
		if ((flags) & 0x100) == 0 {
			bite, err := srcBuf.ReadByte()
			if err != nil {
				break
			}
			c = int(bite)
			flags = uint(c | 0xFF00) /* uses higher byte cleverly to count eight*/
		}
		if flags&1 == 1 {
			bite, err := srcBuf.ReadByte()
			if err != nil {
				break
			}
			c = int(bite)
			dst.WriteByte(byte(c))
			textBuf[r] = byte(c)
			r++
			r &= (n - 1)
		} else {
			bite, err := srcBuf.ReadByte()
			if err != nil {
				break
			}
			i = int(bite)

			bite, err = srcBuf.ReadByte()
			if err != nil {
				break
			}
			j = int(bite)

			i |= ((j & 0xF0) << 4)
			j = (j & 0x0F) + threshold
			for k := 0; k <= j; k++ {
				c = int(textBuf[(i+k)&(n-1)])
				dst.WriteByte(byte(c))
				textBuf[r] = byte(c)
				r++
				r &= (n - 1)
			}
		}
	}

	return dst.Bytes()
}
