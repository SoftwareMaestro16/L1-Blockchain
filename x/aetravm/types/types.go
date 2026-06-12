package types

import (
	"fmt"
	"math/big"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

type TypeKind uint16

const (
	KindNull	TypeKind	= iota
	KindBool
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	KindUint128
	KindUint256
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindInt128
	KindInt256
	KindAddress
	KindHash
	KindCoins
	KindTimestamp
	KindTuple
	KindChunk
	KindExecutionFrame
	KindString
	KindBytes
	KindOption
	KindMap
)

// TypeID is a unique identifier for a type, typically BLAKE3(Schema).
type TypeID [32]byte

func (id TypeID) String() string {
	return fmt.Sprintf("%x", id[:])
}

// Type represents a type in the AVM system.
type Type interface {
	Kind() TypeKind
	TypeID() TypeID
	Schema() *chunk.Chunk	// The type definition as a chunk
	IsAssignableFrom(other Type) bool
	String() string
}

// Value is a runtime container for a data payload and its type information.
type Value struct {
	Type	Type
	Payload	interface{}
}

func (v Value) String() string {
	return fmt.Sprintf("Value(%s: %v)", v.Type.String(), v.Payload)
}

// BaseType implements basic Type interface for primitives.
type BaseType struct {
	kind	TypeKind
	id	TypeID
	schema	*chunk.Chunk
}

func (t *BaseType) Kind() TypeKind		{ return t.kind }
func (t *BaseType) TypeID() TypeID		{ return t.id }
func (t *BaseType) Schema() *chunk.Chunk	{ return t.schema }
func (t *BaseType) String() string		{ return fmt.Sprintf("Type(%d)", t.kind) }
func (t *BaseType) IsAssignableFrom(other Type) bool {
	return t.kind == other.Kind() && t.id == other.TypeID()
}

var (
	TypeNull	= &BaseType{kind: KindNull}
	TypeBool	= &BaseType{kind: KindBool}
	TypeUint8	= &BaseType{kind: KindUint8}
	TypeUint16	= &BaseType{kind: KindUint16}
	TypeUint32	= &BaseType{kind: KindUint32}
	TypeUint64	= &BaseType{kind: KindUint64}
	TypeUint128	= &BaseType{kind: KindUint128}
	TypeUint256	= &BaseType{kind: KindUint256}
	TypeInt8	= &BaseType{kind: KindInt8}
	TypeInt16	= &BaseType{kind: KindInt16}
	TypeInt32	= &BaseType{kind: KindInt32}
	TypeInt64	= &BaseType{kind: KindInt64}
	TypeInt128	= &BaseType{kind: KindInt128}
	TypeInt256	= &BaseType{kind: KindInt256}
	TypeAddress	= &BaseType{kind: KindAddress}
	TypeHash	= &BaseType{kind: KindHash}
	TypeCoins	= TypeUint128
	TypeTimestamp	= TypeUint64
	TypeChunk	= &BaseType{kind: KindChunk}
	TypeExecFrame	= &BaseType{kind: KindExecutionFrame}
)

// TupleType represents a fixed-size ordered collection of values of different types.
type TupleType struct {
	BaseType
	Elements	[]Type
}

// OptionType represents an optional value of type T.
type OptionType struct {
	BaseType
	Underlying	Type
}

// MapType represents a key-value mapping from K to V.
type MapType struct {
	BaseType
	Key	Type
	Value	Type
}

// StringType represents a bounded UTF-8 string.
type StringType struct {
	BaseType
	MaxLength	uint32
}

// BytesType represents a bounded byte slice.
type BytesType struct {
	BaseType
	MaxLength	uint32
}

// Helper for large integers
type Uint128 struct{ big.Int }
type Uint256 struct{ big.Int }
type Int128 struct{ big.Int }
type Int256 struct{ big.Int }
