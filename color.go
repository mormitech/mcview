package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"hash/crc32"
	"os"
	"strings"
)

// A picture that is edited is shown from a PNG that mormiView makes of the decoded pixels. Pixels carry no color space of their own: a
// file in Display P3 (most phone photos) is shown right because the web view reads its profile, and a PNG without that profile would be
// shown as sRGB, a little different. So the color information of the file is copied into the PNG: the pixels stay as they were
// decoded, and the screen shows the same colors before and after an edit.

// colorChunks returns the PNG chunks (complete, with length and CRC) that tell the color space of the picture in a file: iCCP, sRGB,
// gAMA, cHRM and cICP of a PNG, or an iCCP made from the ICC profile of a JPEG or a WebP. Nothing is returned for a file without one.
func colorChunks(path string) [][]byte {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	switch {
	case bytes.HasPrefix(b, []byte("\x89PNG\r\n\x1a\n")):
		return pngColorChunks(b)
	case bytes.HasPrefix(b, []byte{0xff, 0xd8}):
		return iccChunk(jpegProfile(b))
	case len(b) > 12 && string(b[:4]) == "RIFF" && string(b[8:12]) == "WEBP" && strings.EqualFold(path[strings.LastIndex(path, ".")+1:], "webp"):
		return iccChunk(webpProfile(b))
	}
	return nil
}

var colorTypes = map[string]bool{"iCCP": true, "sRGB": true, "gAMA": true, "cHRM": true, "cICP": true}

func pngColorChunks(b []byte) [][]byte {
	var out [][]byte
	for i := 8; i+12 <= len(b); {
		n := int(binary.BigEndian.Uint32(b[i:]))
		end := i + 12 + n
		if n < 0 || end > len(b) {
			break
		}
		typ := string(b[i+4 : i+8])
		if colorTypes[typ] {
			out = append(out, append([]byte(nil), b[i:end]...))
		}
		if typ == "IDAT" {
			break
		}
		i = end
	}
	return out
}

func jpegProfile(b []byte) []byte {
	parts := map[int][]byte{}
	for i := 2; i+4 <= len(b) && b[i] == 0xff; {
		marker := b[i+1]
		if marker == 0xda || marker == 0xd9 { // the picture data starts
			break
		}
		n := int(binary.BigEndian.Uint16(b[i+2:]))
		if n < 2 || i+2+n > len(b) {
			break
		}
		seg := b[i+4 : i+2+n]
		if marker == 0xe2 && len(seg) > 14 && string(seg[:12]) == "ICC_PROFILE\x00" {
			parts[int(seg[12])] = seg[14:]
		}
		i += 2 + n
	}
	var out []byte
	for k := 1; k <= len(parts); k++ {
		p, ok := parts[k]
		if !ok {
			return nil
		}
		out = append(out, p...)
	}
	return out
}

func webpProfile(b []byte) []byte {
	for i := 12; i+8 <= len(b); {
		n := int(binary.LittleEndian.Uint32(b[i+4:]))
		if i+8+n > len(b) {
			break
		}
		if string(b[i:i+4]) == "ICCP" {
			return b[i+8 : i+8+n]
		}
		i += 8 + n + n%2
	}
	return nil
}

// iccChunk is the iCCP chunk of an ICC profile (none for no profile).
func iccChunk(profile []byte) [][]byte {
	if len(profile) == 0 {
		return nil
	}
	var z bytes.Buffer
	z.WriteString("ICC\x00\x00") // the name of the profile, and the method of compression (zlib)
	w := zlib.NewWriter(&z)
	w.Write(profile)
	w.Close()
	return [][]byte{pngChunk("iCCP", z.Bytes())}
}

func pngChunk(typ string, data []byte) []byte {
	c := make([]byte, 0, len(data)+12)
	c = binary.BigEndian.AppendUint32(c, uint32(len(data)))
	c = append(c, typ...)
	c = append(c, data...)
	return binary.BigEndian.AppendUint32(c, crc32.ChecksumIEEE(c[4:]))
}

// withChunks puts chunks into a PNG right after its header chunk, where the color chunks belong.
func withChunks(png []byte, chunks [][]byte) []byte {
	const afterIHDR = 8 + 12 + 13
	if len(chunks) == 0 || len(png) < afterIHDR {
		return png
	}
	out := append([]byte(nil), png[:afterIHDR]...)
	for _, c := range chunks {
		out = append(out, c...)
	}
	return append(out, png[afterIHDR:]...)
}
