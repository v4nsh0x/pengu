package interpreter

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/v4nsh0x/pengu/runtime"
)

func createCryptoModule() *runtime.Value {
	om := runtime.NewOrderedMap()

	// crypto.md5(str) - MD5 hash
	om.Set("md5", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.md5() expects 1 string argument")
		}
		hash := md5.Sum([]byte(args[0].Str))
		return runtime.NewString(hex.EncodeToString(hash[:])), nil
	}))

	// crypto.sha1(str) - SHA1 hash
	om.Set("sha1", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.sha1() expects 1 string argument")
		}
		hash := sha1.Sum([]byte(args[0].Str))
		return runtime.NewString(hex.EncodeToString(hash[:])), nil
	}))

	// crypto.sha256(str) - SHA256 hash
	om.Set("sha256", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.sha256() expects 1 string argument")
		}
		hash := sha256.Sum256([]byte(args[0].Str))
		return runtime.NewString(hex.EncodeToString(hash[:])), nil
	}))

	// crypto.sha512(str) - SHA512 hash
	om.Set("sha512", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.sha512() expects 1 string argument")
		}
		hash := sha512.Sum512([]byte(args[0].Str))
		return runtime.NewString(hex.EncodeToString(hash[:])), nil
	}))

	// crypto.hmac_sha256(message, key) - HMAC-SHA256
	om.Set("hmac_sha256", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.hmac_sha256() expects (message, key) as strings")
		}
		mac := hmac.New(sha256.New, []byte(args[1].Str))
		mac.Write([]byte(args[0].Str))
		return runtime.NewString(hex.EncodeToString(mac.Sum(nil))), nil
	}))

	// crypto.hmac_sha512(message, key) - HMAC-SHA512
	om.Set("hmac_sha512", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.hmac_sha512() expects (message, key) as strings")
		}
		mac := hmac.New(sha512.New, []byte(args[1].Str))
		mac.Write([]byte(args[0].Str))
		return runtime.NewString(hex.EncodeToString(mac.Sum(nil))), nil
	}))

	// crypto.base64_encode(str) - Base64 encode
	om.Set("base64_encode", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.base64_encode() expects 1 string argument")
		}
		encoded := base64.StdEncoding.EncodeToString([]byte(args[0].Str))
		return runtime.NewString(encoded), nil
	}))

	// crypto.base64_decode(str) - Base64 decode
	om.Set("base64_decode", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.base64_decode() expects 1 string argument")
		}
		decoded, err := base64.StdEncoding.DecodeString(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("crypto.base64_decode() failed: %v", err)
		}
		return runtime.NewString(string(decoded)), nil
	}))

	// crypto.hex_encode(str) - Hex encode
	om.Set("hex_encode", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.hex_encode() expects 1 string argument")
		}
		return runtime.NewString(hex.EncodeToString([]byte(args[0].Str))), nil
	}))

	// crypto.hex_decode(str) - Hex decode
	om.Set("hex_decode", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.hex_decode() expects 1 string argument")
		}
		decoded, err := hex.DecodeString(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("crypto.hex_decode() failed: %v", err)
		}
		return runtime.NewString(string(decoded)), nil
	}))

	// crypto.url_encode(str) - URL encode a string
	om.Set("url_encode", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.url_encode() expects 1 string argument")
		}
		return runtime.NewString(url.QueryEscape(args[0].Str)), nil
	}))

	// crypto.url_decode(str) - URL decode a string
	om.Set("url_decode", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.url_decode() expects 1 string argument")
		}
		decoded, err := url.QueryUnescape(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("crypto.url_decode() failed: %v", err)
		}
		return runtime.NewString(decoded), nil
	}))

	// crypto.random_bytes(n) - Generate n random bytes as hex string
	om.Set("random_bytes", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("crypto.random_bytes() expects 1 number argument")
		}
		n := int(args[0].Number)
		if n <= 0 || n > 1024 {
			return nil, fmt.Errorf("crypto.random_bytes() size must be between 1 and 1024")
		}
		buf := make([]byte, n)
		_, err := rand.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("crypto.random_bytes() failed: %v", err)
		}
		return runtime.NewString(hex.EncodeToString(buf)), nil
	}))

	// crypto.uuid() - Generate a UUID v4
	om.Set("uuid", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		buf := make([]byte, 16)
		_, err := rand.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("crypto.uuid() failed: %v", err)
		}
		// Set version (4) and variant bits
		buf[6] = (buf[6] & 0x0f) | 0x40
		buf[8] = (buf[8] & 0x3f) | 0x80
		uuid := fmt.Sprintf("%s-%s-%s-%s-%s",
			hex.EncodeToString(buf[0:4]),
			hex.EncodeToString(buf[4:6]),
			hex.EncodeToString(buf[6:8]),
			hex.EncodeToString(buf[8:10]),
			hex.EncodeToString(buf[10:16]))
		return runtime.NewString(uuid), nil
	}))

	// crypto.compare_hash(hash1, hash2) - Constant-time hash comparison (timing-safe)
	om.Set("compare_hash", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("crypto.compare_hash() expects 2 string arguments")
		}
		a := strings.ToLower(args[0].Str)
		b := strings.ToLower(args[1].Str)
		if len(a) != len(b) {
			return runtime.NewBool(false), nil
		}
		return runtime.NewBool(hmac.Equal([]byte(a), []byte(b))), nil
	}))

	return runtime.NewObject(om)
}
