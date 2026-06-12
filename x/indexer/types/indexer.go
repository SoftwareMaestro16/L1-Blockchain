package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	KindState	= "state"
	KindEvent	= "event"
	KindMemo	= "memo"
	KindDomain	= "domain"
)

type Field struct {
	Key	string
	Value	string
}

type Record struct {
	Kind	string
	Key	string
	Owner	string
	Height	uint64
	TxHash	[]byte
	Value	[]byte
	Fields	[]Field
}

type Query struct {
	Kind	string
	Key	string
	Field	Field
	Limit	uint32
}

type Projection struct {
	Records	[]Record
	byKind	map[string][]Record
	byKey	map[string][]Record
	byField	map[string][]Record
}

func BuildProjection(records []Record) (Projection, error) {
	projection := Projection{
		byKind:		make(map[string][]Record),
		byKey:		make(map[string][]Record),
		byField:	make(map[string][]Record),
	}
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return Projection{}, err
		}
		record = record.Clone()
		projection.Records = append(projection.Records, record)
		projection.byKind[record.Kind] = append(projection.byKind[record.Kind], record)
		projection.byKey[record.Key] = append(projection.byKey[record.Key], record)
		for _, field := range record.Fields {
			projection.byField[fieldIndexKey(field)] = append(projection.byField[fieldIndexKey(field)], record)
		}
	}
	sortRecords(projection.Records)
	sortProjectionBuckets(projection.byKind)
	sortProjectionBuckets(projection.byKey)
	sortProjectionBuckets(projection.byField)
	return projection, nil
}

func (p Projection) Search(query Query) ([]Record, error) {
	if query.Limit == 0 {
		return nil, errors.New("indexer query limit must be positive")
	}
	var candidates []Record
	switch {
	case query.Kind != "":
		if !IsKind(query.Kind) {
			return nil, fmt.Errorf("invalid indexer kind %q", query.Kind)
		}
		candidates = p.byKind[query.Kind]
	case query.Key != "":
		candidates = p.byKey[query.Key]
	case query.Field.Key != "":
		if query.Field.Value == "" {
			return nil, errors.New("indexer field value is required")
		}
		candidates = p.byField[fieldIndexKey(query.Field)]
	default:
		return nil, errors.New("indexer query requires kind, key, or field")
	}
	out := cloneRecords(candidates)
	if len(out) > int(query.Limit) {
		out = out[:query.Limit]
	}
	return out, nil
}

func (r Record) Validate() error {
	if !IsKind(r.Kind) {
		return fmt.Errorf("invalid indexer kind %q", r.Kind)
	}
	if strings.TrimSpace(r.Key) == "" {
		return errors.New("indexer key is required")
	}
	for _, field := range r.Fields {
		if strings.TrimSpace(field.Key) == "" {
			return errors.New("indexer field key is required")
		}
		if field.Value == "" {
			return errors.New("indexer field value is required")
		}
	}
	if !fieldsSorted(r.Fields) {
		return errors.New("indexer fields must be sorted")
	}
	return nil
}

func (r Record) Clone() Record {
	out := r
	out.TxHash = append([]byte(nil), r.TxHash...)
	out.Value = append([]byte(nil), r.Value...)
	out.Fields = append([]Field(nil), r.Fields...)
	return out
}

func CanonicalFields(fields []Field) []Field {
	out := append([]Field(nil), fields...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Key == out[j].Key {
			return out[i].Value < out[j].Value
		}
		return out[i].Key < out[j].Key
	})
	return out
}

func ConsensusRequired() bool {
	return false
}

func IsKind(kind string) bool {
	switch kind {
	case KindState, KindEvent, KindMemo, KindDomain:
		return true
	default:
		return false
	}
}

func fieldIndexKey(field Field) string {
	return field.Key + "\x00" + field.Value
}

func fieldsSorted(fields []Field) bool {
	for i := 1; i < len(fields); i++ {
		if fields[i-1].Key > fields[i].Key {
			return false
		}
		if fields[i-1].Key == fields[i].Key && fields[i-1].Value > fields[i].Value {
			return false
		}
	}
	return true
}

func sortProjectionBuckets(buckets map[string][]Record) {
	for key := range buckets {
		sortRecords(buckets[key])
	}
}

func sortRecords(records []Record) {
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Height != records[j].Height {
			return records[i].Height < records[j].Height
		}
		if records[i].Kind != records[j].Kind {
			return records[i].Kind < records[j].Kind
		}
		if records[i].Key != records[j].Key {
			return records[i].Key < records[j].Key
		}
		return string(records[i].TxHash) < string(records[j].TxHash)
	})
}

func cloneRecords(records []Record) []Record {
	out := make([]Record, len(records))
	for i, record := range records {
		out[i] = record.Clone()
	}
	return out
}
