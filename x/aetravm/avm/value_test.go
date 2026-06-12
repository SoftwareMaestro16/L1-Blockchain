package avm

import (
	"math/big"
	"testing"
)

func TestValueTagGoldenVectors(t *testing.T) {
	golden := map[ValueTag]byte{
		TagNull:		0,
		TagBool:		1,
		TagInt8:		2,
		TagInt16:		3,
		TagInt32:		4,
		TagInt64:		5,
		TagInt128:		6,
		TagInt256:		7,
		TagUint8:		8,
		TagUint16:		9,
		TagUint32:		10,
		TagUint64:		11,
		TagUint128:		12,
		TagUint256:		13,
		TagCoins:		14,
		TagTimestamp:		15,
		TagAddress:		16,
		TagHash:		17,
		TagBytes:		18,
		TagString:		19,
		TagTuple:		20,
		TagChunkRef:		21,
		TagReaderCursor:	22,
		TagWriterHandle:	23,
		TagExecFrameRef:	24,
	}
	for tag, expected := range golden {
		if byte(tag) != expected {
			t.Errorf("tag %s: expected byte %d, got %d", tag, expected, byte(tag))
		}
	}
}

func TestValueTagString(t *testing.T) {
	tests := map[ValueTag]string{
		TagNull:		"null",
		TagBool:		"bool",
		TagInt64:		"int64",
		TagUint256:		"uint256",
		TagCoins:		"coins",
		TagAddress:		"address",
		TagHash:		"hash",
		TagString:		"string",
		TagTuple:		"tuple",
		TagChunkRef:		"chunk_ref",
		TagReaderCursor:	"reader_cursor",
		TagWriterHandle:	"writer_handle",
		TagExecFrameRef:	"exec_frame_ref",
	}
	for tag, expected := range tests {
		if tag.String() != expected {
			t.Errorf("expected %s, got %s", expected, tag.String())
		}
	}
}

func TestCanonicalEncodeDecodeBool(t *testing.T) {
	for _, v := range []bool{true, false} {
		val := ValueBool(v)
		encoded, err := CanonicalEncode(val)
		if err != nil {
			t.Fatalf("encode bool %v: %v", v, err)
		}
		decoded, n, err := CanonicalDecode(encoded)
		if err != nil {
			t.Fatalf("decode bool %v: %v", v, err)
		}
		if n != len(encoded) {
			t.Errorf("bytes consumed mismatch: %d vs %d", n, len(encoded))
		}
		got, _ := decoded.AsBool()
		if got != v {
			t.Errorf("round trip: expected %v, got %v", v, got)
		}
	}
}

func TestCanonicalEncodeDecodeIntegers(t *testing.T) {
	tests := []struct {
		name	string
		value	RuntimeValue
		tag	ValueTag
	}{
		{"int8_max", ValueInt8(127), TagInt8},
		{"int8_min", ValueInt8(-128), TagInt8},
		{"uint8_200", ValueUint8(200), TagUint8},
		{"int64_big", ValueInt64(1 << 40), TagInt64},
		{"uint64_max", ValueUint64(18446744073709551615), TagUint64},
		{"uint128_1", ValueBigUint128(big.NewInt(1)), TagUint128},
		{"timestamp", ValueTimestamp(1700000000), TagTimestamp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := CanonicalEncode(tt.value)
			if err != nil {
				t.Fatalf("encode %s: %v", tt.name, err)
			}
			if encoded[0] != byte(tt.tag) {
				t.Errorf("tag byte: expected %d, got %d", tt.tag, encoded[0])
			}
			decoded, n, err := CanonicalDecode(encoded)
			if err != nil {
				t.Fatalf("decode %s: %v", tt.name, err)
			}
			if n != len(encoded) {
				t.Errorf("bytes consumed mismatch: %d vs %d", n, len(encoded))
			}
			if decoded.Tag != tt.tag {
				t.Errorf("tag mismatch: expected %v, got %v", tt.tag, decoded.Tag)
			}
		})
	}
}

func TestCanonicalEncodeDecodeString(t *testing.T) {
	tests := []string{"", "hello", "Aetra VM", "🦀"}
	for _, s := range tests {
		val := ValueString(s)
		encoded, err := CanonicalEncode(val)
		if err != nil {
			t.Fatalf("encode string %q: %v", s, err)
		}
		decoded, n, err := CanonicalDecode(encoded)
		if err != nil {
			t.Fatalf("decode string %q: %v", s, err)
		}
		if n != len(encoded) {
			t.Errorf("bytes consumed mismatch for %q", s)
		}
		got, _ := decoded.AsString()
		if got != s {
			t.Errorf("round trip: expected %q, got %q", s, got)
		}
	}
}

func TestCanonicalEncodeInvalidUTF8(t *testing.T) {
	val := RuntimeValue{Tag: TagString, strVal: string([]byte{0xff, 0xfe, 0xfd})}
	_, err := CanonicalEncode(val)
	if err == nil {
		t.Error("expected error for invalid UTF-8 string")
	}
}

func TestCanonicalEncodeDecodeBytes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	val := ValueBytes(data)
	encoded, err := CanonicalEncode(val)
	if err != nil {
		t.Fatalf("encode bytes: %v", err)
	}
	decoded, n, err := CanonicalDecode(encoded)
	if err != nil {
		t.Fatalf("decode bytes: %v", err)
	}
	if n != len(encoded) {
		t.Errorf("bytes consumed mismatch")
	}
	got, _ := decoded.AsBytes()
	if len(got) != len(data) {
		t.Errorf("length mismatch: %d vs %d", len(got), len(data))
	}
	for i := range data {
		if got[i] != data[i] {
			t.Errorf("byte %d: expected %02x, got %02x", i, data[i], got[i])
		}
	}
}

func TestCanonicalEncodeDecodeHash(t *testing.T) {
	var h [32]byte
	for i := range h {
		h[i] = byte(i)
	}
	val := ValueHash(h)
	encoded, err := CanonicalEncode(val)
	if err != nil {
		t.Fatalf("encode hash: %v", err)
	}
	if len(encoded) != 33 {
		t.Errorf("hash encoded length: expected 33, got %d", len(encoded))
	}
	decoded, n, err := CanonicalDecode(encoded)
	if err != nil {
		t.Fatalf("decode hash: %v", err)
	}
	got, _ := decoded.AsHash()
	if got != h {
		t.Error("hash round trip mismatch")
	}
	if n != 33 {
		t.Errorf("bytes consumed: expected 33, got %d", n)
	}
}

func TestCanonicalEncodeDecodeAddress(t *testing.T) {
	addr := "AE:test:aetra123"
	val := ValueAddress(addr)
	encoded, err := CanonicalEncode(val)
	if err != nil {
		t.Fatalf("encode address: %v", err)
	}
	decoded, n, err := CanonicalDecode(encoded)
	if err != nil {
		t.Fatalf("decode address: %v", err)
	}
	got, _ := decoded.AsAddress()
	if got != addr {
		t.Errorf("round trip: expected %q, got %q", addr, got)
	}
	if n != len(encoded) {
		t.Errorf("bytes consumed mismatch")
	}
}

func TestCanonicalEncodeDecodeTuple(t *testing.T) {
	inner := ValueTuple([]RuntimeValue{
		ValueUint64(42),
		ValueBool(true),
	})
	outer := ValueTuple([]RuntimeValue{
		ValueString("hello"),
		inner,
		ValueNull(),
	})

	encoded, err := CanonicalEncode(outer)
	if err != nil {
		t.Fatalf("encode tuple: %v", err)
	}

	decoded, n, err := CanonicalDecode(encoded)
	if err != nil {
		t.Fatalf("decode tuple: %v", err)
	}
	if n != len(encoded) {
		t.Errorf("bytes consumed: %d vs %d", n, len(encoded))
	}
	tup, _ := decoded.AsTuple()
	if len(tup) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(tup))
	}
	s, _ := tup[0].AsString()
	if s != "hello" {
		t.Errorf("element 0: expected 'hello', got %q", s)
	}
	innerDecoded, _ := tup[1].AsTuple()
	if len(innerDecoded) != 2 {
		t.Errorf("inner tuple: expected 2 elements, got %d", len(innerDecoded))
	}
	if !tup[2].IsNull() {
		t.Error("element 2: expected null")
	}
}

func TestCanonicalEmptyTuple(t *testing.T) {
	val := ValueEmptyTuple()
	encoded, err := CanonicalEncode(val)
	if err != nil {
		t.Fatalf("encode empty tuple: %v", err)
	}
	decoded, n, err := CanonicalDecode(encoded)
	if err != nil {
		t.Fatalf("decode empty tuple: %v", err)
	}
	tup, _ := decoded.AsTuple()
	if len(tup) != 0 {
		t.Errorf("expected empty tuple, got %d elements", len(tup))
	}
	if n != len(encoded) {
		t.Errorf("bytes consumed mismatch")
	}
}

func TestOptionNoneIsTagNull(t *testing.T) {
	none := OptionNone()
	if !none.IsNull() {
		t.Error("Option.None should be TagNull")
	}
	if !IsOptionNone(none) {
		t.Error("IsOptionNone should return true for null")
	}
}

func TestOptionSomePreservesValue(t *testing.T) {
	some := OptionSome(ValueUint64(42))
	if some.Tag != TagUint64 {
		t.Errorf("Option.Some should preserve tag, got %v", some.Tag)
	}
	if IsOptionNone(some) {
		t.Error("Option.Some should not be None")
	}
}

func TestOptionRoundTrip(t *testing.T) {
	none := OptionNone()
	encNone, err := CanonicalEncode(none)
	if err != nil {
		t.Fatal(err)
	}
	if len(encNone) != 1 || encNone[0] != 0 {
		t.Errorf("Option.None should encode as single 0x00 byte, got %v", encNone)
	}

	some := OptionSome(ValueBool(true))
	encSome, _ := CanonicalEncode(some)
	if encSome[0] != byte(TagBool) {
		t.Errorf("Option.Some should encode with value tag, got %d", encSome[0])
	}
}

func TestExplicitCastSameType(t *testing.T) {
	v := ValueUint64(42)
	result, code := ExplicitCast(v, TagUint64)
	if code != CastOK {
		t.Errorf("same-type cast should be OK, got %d", code)
	}
	casted, _ := result.AsUint64()
	if casted != 42 {
		t.Errorf("same-type cast should preserve value, got %d", casted)
	}
}

func TestExplicitCastIntToUint(t *testing.T) {
	v := ValueInt64(100)
	result, code := ExplicitCast(v, TagUint64)
	if code != CastOK {
		t.Errorf("int64→uint64 cast should be OK, got %d", code)
	}
	u, _ := result.AsUint64()
	if u != 100 {
		t.Errorf("expected 100, got %d", u)
	}
}

func TestExplicitCastTypeMismatch(t *testing.T) {
	v := ValueBool(true)
	_, code := ExplicitCast(v, TagUint64)
	if code != CastTypeMismatch {
		t.Errorf("bool→uint64 cast should be TypeMismatch, got %d", code)
	}
}

func TestExplicitCastNullToValue(t *testing.T) {
	v := ValueNull()
	_, code := ExplicitCast(v, TagUint64)
	if code != CastNullToValue {
		t.Errorf("null→uint64 cast should be NullToValue, got %d", code)
	}
}

func TestValueBitWidth(t *testing.T) {
	tests := map[ValueTag]int{
		TagInt8:	8,
		TagUint8:	8,
		TagInt16:	16,
		TagUint16:	16,
		TagInt32:	32,
		TagUint32:	32,
		TagInt64:	64,
		TagUint64:	64,
		TagInt128:	128,
		TagUint128:	128,
		TagCoins:	128,
		TagInt256:	256,
		TagUint256:	256,
	}
	for tag, expected := range tests {
		width, ok := ValueBitWidth(tag)
		if !ok {
			t.Errorf("tag %v: expected valid bit width", tag)
		}
		if width != expected {
			t.Errorf("tag %v: expected width %d, got %d", tag, expected, width)
		}
	}
	_, ok := ValueBitWidth(TagBool)
	if ok {
		t.Error("non-integer tag should not have bit width")
	}
}

func TestSignedness(t *testing.T) {
	if !IsSigned(TagInt64) {
		t.Error("int64 should be signed")
	}
	if IsSigned(TagUint64) {
		t.Error("uint64 should not be signed")
	}
	if !IsInteger(TagUint64) {
		t.Error("uint64 should be integer")
	}
	if IsInteger(TagBool) {
		t.Error("bool should not be integer")
	}
}

func TestCanonicalDeterminism(t *testing.T) {
	v1 := ValueUint64(42)
	v2 := ValueUint64(42)
	enc1, _ := CanonicalEncode(v1)
	enc2, _ := CanonicalEncode(v2)
	if len(enc1) != len(enc2) {
		t.Error("same value should produce same encoded length")
	}
	for i := range enc1 {
		if enc1[i] != enc2[i] {
			t.Errorf("byte %d differs: %02x vs %02x", i, enc1[i], enc2[i])
		}
	}
}

func TestCanonicalDeterminismHash(t *testing.T) {
	v1 := ValueString("hello")
	v2 := ValueString("hello")
	h1, _ := CanonicalHash(v1)
	h2, _ := CanonicalHash(v2)
	if h1 != h2 {
		t.Error("same value should produce same canonical hash")
	}
}

func TestCanonicalDifferentValuesDifferentHashes(t *testing.T) {
	v1 := ValueString("hello")
	v2 := ValueString("world")
	h1, _ := CanonicalHash(v1)
	h2, _ := CanonicalHash(v2)
	if h1 == h2 {
		t.Error("different values should produce different hashes")
	}
}

func TestValidateDeterminism(t *testing.T) {
	v1 := ValueUint64(100)
	v2 := ValueUint64(100)
	if err := ValidateDeterminism(v1, v2); err != nil {
		t.Errorf("identical values should pass: %v", err)
	}

	v3 := ValueUint64(200)
	if err := ValidateDeterminism(v1, v3); err == nil {
		t.Error("different values should fail determinism check")
	}
}

func TestValidateDeterminismDifferentTags(t *testing.T) {
	v1 := ValueUint64(42)
	v2 := ValueInt64(42)
	if err := ValidateDeterminism(v1, v2); err == nil {
		t.Error("different tags should fail determinism check")
	}
}

func TestMaxEncodedSize(t *testing.T) {
	if MaxEncodedSizeForTag(TagNull) != 1 {
		t.Errorf("null max size: expected 1, got %d", MaxEncodedSizeForTag(TagNull))
	}
	if MaxEncodedSizeForTag(TagBool) != 2 {
		t.Errorf("bool max size: expected 2, got %d", MaxEncodedSizeForTag(TagBool))
	}
	if MaxEncodedSizeForTag(TagUint64) != 9 {
		t.Errorf("uint64 max size: expected 9, got %d", MaxEncodedSizeForTag(TagUint64))
	}
	if MaxEncodedSizeForTag(TagHash) != 33 {
		t.Errorf("hash max size: expected 33, got %d", MaxEncodedSizeForTag(TagHash))
	}
	if MaxEncodedSizeForTag(ValueTag(99)) != 0 {
		t.Error("unknown tag should have 0 size")
	}
}

func TestGasCosts(t *testing.T) {
	if GasCostEncode(TagNull) != 1 {
		t.Errorf("null encode gas: expected 1, got %d", GasCostEncode(TagNull))
	}
	if GasCostDecode(TagUint64) != 3 {
		t.Errorf("uint64 decode gas: expected 3, got %d", GasCostDecode(TagUint64))
	}
	if GasCostEncode(TagChunkRef) != 15 {
		t.Errorf("chunkref encode gas: expected 15, got %d", GasCostEncode(TagChunkRef))
	}
}

func TestTypeSafeAccessors(t *testing.T) {
	v := ValueUint64(42)
	u, err := v.AsUint64()
	if err != nil || u != 42 {
		t.Errorf("AsUint64: expected 42, got %d, err=%v", u, err)
	}
	_, err = v.AsBool()
	if err == nil {
		t.Error("AsBool on uint64 should fail")
	}
	_, err = v.AsString()
	if err == nil {
		t.Error("AsString on uint64 should fail")
	}
	_, err = v.AsAddress()
	if err == nil {
		t.Error("AsAddress on uint64 should fail")
	}
}

func TestNullAccessor(t *testing.T) {
	v := ValueNull()
	if !v.IsNull() {
		t.Error("null should be null")
	}
	if v.IsBool() {
		t.Error("null should not be bool")
	}
}

func TestCoinsValue(t *testing.T) {
	v := ValueCoins(big.NewInt(1000000))
	if v.Tag != TagCoins {
		t.Errorf("coins tag: expected %v, got %v", TagCoins, v.Tag)
	}
	enc, err := CanonicalEncode(v)
	if err != nil {
		t.Fatalf("encode coins: %v", err)
	}
	if enc[0] != byte(TagCoins) {
		t.Errorf("coins tag byte: expected %d, got %d", TagCoins, enc[0])
	}
}
