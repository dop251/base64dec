package base64dec

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"
)

func TestDecodeBase64(t *testing.T) {
	f := func(t *testing.T, input, expectedResult string, expectedError error) {
		decoded := make([]byte, base64.RawStdEncoding.DecodedLen(len(input)))
		n, err := DecodeBase64(decoded, input)
		if err != expectedError {
			t.Fatal(err)
		}
		encoded := base64.StdEncoding.EncodeToString(decoded[:n])
		if encoded != expectedResult {
			t.Fatal(encoded)
		}
	}

	t.Run("empty input", func(t *testing.T) {
		f(t, "", "", nil)
	})

	t.Run("newlines only", func(t *testing.T) {
		f(t, "\r\n", "", nil)
	})

	t.Run("correct padding", func(t *testing.T) {
		f(t, "Z9CA-w==", "Z9CA+w==", nil)
	})

	t.Run("correct padding split by newline", func(t *testing.T) {
		f(t, "Z9CA-w=\n=", "Z9CA+w==", nil)
	})

	t.Run("correct padding with concatenation", func(t *testing.T) {
		f(t, "Z9CA-w==Z9CA-w==", "Z9CA+w==", base64.CorruptInputError(8))
	})

	t.Run("trailing newline", func(t *testing.T) {
		f(t, "Z9CA+wZ9CA-w\n", "Z9CA+wZ9CA+w", nil)
	})

	t.Run("trailing newline with padding", func(t *testing.T) {
		f(t, "Z9CA+wZ9CA-www==\n", "Z9CA+wZ9CA+www==", nil)
	})

	t.Run("garbage after newline", func(t *testing.T) {
		f(t, "Z9CA+wZ9CA-www==\n?", "Z9CA+wZ9CA+www==", base64.CorruptInputError(17))
	})

	t.Run("no padding", func(t *testing.T) {
		f(t, "Z9CA-w", "Z9CA+w==", nil)
	})

	t.Run("no padding, garbage at the end", func(t *testing.T) {
		f(t, "Z9CA-w???", "Z9CA+w==", base64.CorruptInputError(6))
	})

	t.Run("not enough padding", func(t *testing.T) {
		f(t, "Z9CA-w=", "Z9CA+w==", base64.CorruptInputError(6))
	})

	t.Run("incorrect padding", func(t *testing.T) {
		f(t, "Z9CA====", "Z9CA", base64.CorruptInputError(4))
	})

	t.Run("incorrect padding with extra base64", func(t *testing.T) {
		f(t, "Z9CA-w=Z9CA-w=", "Z9CA+w==", base64.CorruptInputError(6))
	})

	t.Run("incorrect padding with garbage", func(t *testing.T) {
		f(t, "Z9CA-w=???", "Z9CA+w==", base64.CorruptInputError(6))
	})

}

func FuzzDecodeBase64(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("\x14\xfb\x9c\x03\xd9\x7e"))
	f.Add([]byte("\x14\xfb\x9c\x03\xd9"))
	f.Add([]byte("\x14\xfb\x9c\x03"))

	f.Fuzz(func(t *testing.T, b []byte) {
		encoded := base64.StdEncoding.EncodeToString(b)
		decoded := make([]byte, len(b))
		n, err := DecodeBase64(decoded, encoded)
		if err != nil {
			t.Fatalf("%v: %v", b, err)
		}
		if !bytes.Equal(decoded[:n], b) {
			t.Fatal(b)
		}

	})
}

func FuzzDecodeBase64String(f *testing.F) {
	f.Add("Z9CA-w")
	f.Add("=\n=")
	f.Add("=")
	f.Add("====")

	f.Fuzz(func(t *testing.T, s string) {
		decoded := make([]byte, base64.RawStdEncoding.DecodedLen(len(s)))
		_, _ = DecodeBase64(decoded, s) // should not panic
	})
}

func BenchmarkDecodeBase64(b *testing.B) {
	sizes := []int{2, 4, 8, 64, 8192}
	dst := make([]byte, 8192)
	benchFunc := func(b *testing.B, benchSize int) {
		data := base64.StdEncoding.EncodeToString(make([]byte, benchSize))
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			DecodeBase64(dst, data)
		}
	}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			benchFunc(b, size)
		})
	}
}

func BenchmarkDecodeBase64Std(b *testing.B) {
	sizes := []int{2, 4, 8, 64, 8192}
	dst := make([]byte, 8192)
	benchFunc := func(b *testing.B, benchSize int) {
		data := base64.StdEncoding.EncodeToString(make([]byte, benchSize))
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			base64.StdEncoding.Decode(dst, []byte(data))
		}
	}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			benchFunc(b, size)
		})
	}
}
