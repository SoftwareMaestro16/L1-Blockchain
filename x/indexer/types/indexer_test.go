package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectionSearchesStateEventsMemoAndDomain(t *testing.T) {
	records := []Record{
		record(KindDomain, "alice.aet", 5, Field{Key: "domain", Value: "alice.aet"}),
		record(KindMemo, "tx:1", 3, Field{Key: "memo", Value: "hello"}),
		record(KindEvent, "event:transfer", 2, Field{Key: "event", Value: "transfer"}),
		record(KindState, "contract:counter", 1, Field{Key: "contract", Value: "counter"}),
	}
	projection, err := BuildProjection(records)
	require.NoError(t, err)

	state, err := projection.Search(Query{Kind: KindState, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "contract:counter", state[0].Key)

	domain, err := projection.Search(Query{Key: "alice.aet", Limit: 10})
	require.NoError(t, err)
	require.Equal(t, KindDomain, domain[0].Kind)
}

func TestIndexerSearchIsDeterministicAndBounded(t *testing.T) {
	projection, err := BuildProjection([]Record{
		record(KindEvent, "event:b", 2, Field{Key: "event", Value: "swap"}),
		record(KindEvent, "event:a", 1, Field{Key: "event", Value: "swap"}),
	})
	require.NoError(t, err)

	found, err := projection.Search(Query{Field: Field{Key: "event", Value: "swap"}, Limit: 1})
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.Equal(t, "event:a", found[0].Key)
}

func TestIndexerValidationAndConsensusBoundary(t *testing.T) {
	_, err := BuildProjection([]Record{{Kind: "bad", Key: "x"}})
	require.ErrorContains(t, err, "invalid")

	_, err = BuildProjection([]Record{{Kind: KindState, Key: "x", Fields: []Field{{Key: "z", Value: "2"}, {Key: "a", Value: "1"}}}})
	require.ErrorContains(t, err, "sorted")

	projection, err := BuildProjection(nil)
	require.NoError(t, err)
	_, err = projection.Search(Query{Kind: KindState})
	require.ErrorContains(t, err, "limit")
	require.False(t, ConsensusRequired())
}

func record(kind string, key string, height uint64, fields ...Field) Record {
	return Record{
		Kind:	kind,
		Key:	key,
		Height:	height,
		TxHash:	[]byte(key),
		Fields:	CanonicalFields(fields),
	}
}
