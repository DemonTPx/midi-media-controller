package main

import "bytes"

var font = map[byte]uint8{
	'0':  0b0111111,
	'1':  0b0000110,
	'2':  0b1011011,
	'3':  0b1001111,
	'4':  0b1100110,
	'5':  0b1101101,
	'6':  0b1111101,
	'7':  0b0100111,
	'8':  0b1111111,
	'9':  0b1101111,
	'A':  0b1110111,
	'B':  0b1111111,
	'C':  0b0111001,
	'D':  0b0111111,
	'E':  0b1111001,
	'F':  0b1110001,
	'G':  0b0111101,
	'H':  0b1110110,
	'I':  0b0110000,
	'J':  0b0001110,
	'K':  0b1110101,
	'L':  0b0111000,
	'M':  0b0010101,
	'N':  0b0110111,
	'O':  0b0111111,
	'P':  0b1110011,
	'Q':  0b0111111,
	'R':  0b1110111,
	'S':  0b1101101,
	'T':  0b1111000,
	'U':  0b0111110,
	'V':  0b0111110,
	'W':  0b0101010,
	'X':  0b1001001,
	'Y':  0b1101110,
	'Z':  0b1011011,
	'a':  0b1011111,
	'b':  0b1111100,
	'c':  0b1011000,
	'd':  0b1011110,
	'e':  0b1111011,
	'f':  0b1110001,
	'g':  0b1101111,
	'h':  0b1110100,
	'i':  0b0010000,
	'j':  0b0001100,
	'k':  0b1110101,
	'l':  0b0110000,
	'm':  0b0010100,
	'n':  0b1010100,
	'o':  0b1011100,
	'p':  0b1110011,
	'q':  0b1100111,
	'r':  0b1010000,
	's':  0b1101101,
	't':  0b1111000,
	'u':  0b0011100,
	'v':  0b0011100,
	'w':  0b0010100,
	'x':  0b1001000,
	'y':  0b1101110,
	'z':  0b1011011,
	':':  0b0001001,
	'-':  0b1000000,
	')':  0b0001111,
	'(':  0b0111001,
	' ':  0,
	'.':  0b0001000,
	'"':  0b0100010,
	'_':  0b0001000,
	'\'': 0b0100000,
}

var dotted = []byte("QRUu")

func lcd7bitRender(text []byte) []uint8 {
	data := make([]uint8, len(text))

	for i := range text {
		data[i] = lcd7bitLetter(text[i])
	}

	return data
}

func lcd7bitLetter(c byte) uint8 {
	if i, ok := font[c]; ok {
		return i
	}

	return 0
}

func lcd7bitRenderDots(text []byte) []uint8 {
	data := []uint8{0, 0}

	for i := range text {
		if !bytes.ContainsRune(dotted, rune(text[i])) {
			continue
		}
		r := i / 7
		data[r] = data[r] | (1 << (i % 7))
	}

	return data
}
