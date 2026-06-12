package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	AVMComparisonFeatureExecution	AVMExecutionComparisonFeature	= "Execution"
	AVMComparisonFeatureState	AVMExecutionComparisonFeature	= "State"
	AVMComparisonFeatureMessaging	AVMExecutionComparisonFeature	= "Messaging"
	AVMComparisonFeatureScheduling	AVMExecutionComparisonFeature	= "Scheduling"
	AVMComparisonFeatureContracts	AVMExecutionComparisonFeature	= "Contracts"
	AVMComparisonFeatureUXModel	AVMExecutionComparisonFeature	= "UX model"

	MaxAVMComparisonRows		= 32
	MaxAVMComparisonTextBytes	= 128
)

type AVMExecutionComparisonFeature string

type AVMExecutionComparisonRow struct {
	Feature			AVMExecutionComparisonFeature
	ClassicCosmosSDK	string
	AVM			string
	RequiresAVMExtension	bool
	RowHash			string
}

type AVMExecutionModelComparison struct {
	ModelName	string
	Rows		[]AVMExecutionComparisonRow
	MatrixHash	string
}

func DefaultAVMExecutionModelComparison() (AVMExecutionModelComparison, error) {
	rows := []AVMExecutionComparisonRow{
		{Feature: AVMComparisonFeatureExecution, ClassicCosmosSDK: "synchronous", AVM: "sync + async", RequiresAVMExtension: true},
		{Feature: AVMComparisonFeatureState, ClassicCosmosSDK: "KVStore", AVM: "KVStore + zone roots", RequiresAVMExtension: true},
		{Feature: AVMComparisonFeatureMessaging, ClassicCosmosSDK: "tx-only", AVM: "message-driven", RequiresAVMExtension: true},
		{Feature: AVMComparisonFeatureScheduling, ClassicCosmosSDK: "block-bound", AVM: "cross-block", RequiresAVMExtension: true},
		{Feature: AVMComparisonFeatureContracts, ClassicCosmosSDK: "module-based", AVM: "actor + module hybrid", RequiresAVMExtension: true},
		{Feature: AVMComparisonFeatureUXModel, ClassicCosmosSDK: "CLI/API", AVM: "interface-driven", RequiresAVMExtension: true},
	}
	for i := range rows {
		row, err := NewAVMExecutionComparisonRow(rows[i])
		if err != nil {
			return AVMExecutionModelComparison{}, err
		}
		rows[i] = row
	}
	return NewAVMExecutionModelComparison(AVMExecutionModelComparison{
		ModelName:	"Classic Cosmos SDK vs AVM",
		Rows:		rows,
	})
}

func NewAVMExecutionComparisonRow(row AVMExecutionComparisonRow) (AVMExecutionComparisonRow, error) {
	row = canonicalAVMExecutionComparisonRow(row)
	row.RowHash = ComputeAVMExecutionComparisonRowHash(row)
	return row, row.Validate()
}

func (r AVMExecutionComparisonRow) Validate() error {
	r = canonicalAVMExecutionComparisonRow(r)
	if !IsAVMExecutionComparisonFeature(r.Feature) {
		return fmt.Errorf("invalid AVM execution comparison feature %q", r.Feature)
	}
	if err := validateAVMComparisonText("Classic Cosmos SDK comparison value", r.ClassicCosmosSDK); err != nil {
		return err
	}
	if err := validateAVMComparisonText("AVM comparison value", r.AVM); err != nil {
		return err
	}
	if r.ClassicCosmosSDK == r.AVM {
		return errors.New("AVM execution comparison row must distinguish AVM from Classic Cosmos SDK")
	}
	if r.RowHash == "" {
		return errors.New("AVM execution comparison row hash is required")
	}
	if err := validateAVMComparisonHash("AVM execution comparison row hash", r.RowHash); err != nil {
		return err
	}
	if r.RowHash != ComputeAVMExecutionComparisonRowHash(r) {
		return errors.New("AVM execution comparison row hash mismatch")
	}
	return nil
}

func NewAVMExecutionModelComparison(model AVMExecutionModelComparison) (AVMExecutionModelComparison, error) {
	model = canonicalAVMExecutionModelComparison(model)
	model.MatrixHash = ComputeAVMExecutionModelComparisonHash(model)
	return model, model.Validate()
}

func (m AVMExecutionModelComparison) Validate() error {
	m = canonicalAVMExecutionModelComparison(m)
	if err := validateAVMComparisonText("AVM execution comparison model name", m.ModelName); err != nil {
		return err
	}
	if len(m.Rows) == 0 || len(m.Rows) > MaxAVMComparisonRows {
		return fmt.Errorf("AVM execution comparison rows must be 1..%d", MaxAVMComparisonRows)
	}
	required := AllAVMExecutionComparisonFeatures()
	if len(m.Rows) != len(required) {
		return errors.New("AVM execution comparison must contain every section 16 feature")
	}
	seen := make(map[AVMExecutionComparisonFeature]struct{}, len(m.Rows))
	for i, row := range m.Rows {
		if err := row.Validate(); err != nil {
			return err
		}
		if _, found := seen[row.Feature]; found {
			return fmt.Errorf("duplicate AVM execution comparison feature %q", row.Feature)
		}
		seen[row.Feature] = struct{}{}
		if i > 0 && m.Rows[i-1].Feature >= row.Feature {
			return errors.New("AVM execution comparison rows must be sorted canonically")
		}
	}
	for _, feature := range required {
		if _, found := seen[feature]; !found {
			return fmt.Errorf("missing AVM execution comparison feature %q", feature)
		}
	}
	if m.MatrixHash == "" {
		return errors.New("AVM execution comparison matrix hash is required")
	}
	if err := validateAVMComparisonHash("AVM execution comparison matrix hash", m.MatrixHash); err != nil {
		return err
	}
	if m.MatrixHash != ComputeAVMExecutionModelComparisonHash(m) {
		return errors.New("AVM execution comparison matrix hash mismatch")
	}
	return nil
}

func RenderAVMExecutionModelComparisonMarkdown(model AVMExecutionModelComparison) (string, error) {
	model = canonicalAVMExecutionModelComparison(model)
	if err := model.Validate(); err != nil {
		return "", err
	}
	lines := []string{
		"| Feature | Classic Cosmos SDK | AVM |",
		"| --- | --- | --- |",
	}
	for _, row := range sortAVMExecutionComparisonRowsByDocumentOrder(model.Rows) {
		lines = append(lines, fmt.Sprintf("| %s | %s | %s |", row.Feature, row.ClassicCosmosSDK, row.AVM))
	}
	return strings.Join(lines, "\n"), nil
}

func AllAVMExecutionComparisonFeatures() []AVMExecutionComparisonFeature {
	features := []AVMExecutionComparisonFeature{
		AVMComparisonFeatureContracts,
		AVMComparisonFeatureExecution,
		AVMComparisonFeatureMessaging,
		AVMComparisonFeatureScheduling,
		AVMComparisonFeatureState,
		AVMComparisonFeatureUXModel,
	}
	sort.Slice(features, func(i, j int) bool { return features[i] < features[j] })
	return features
}

func IsAVMExecutionComparisonFeature(feature AVMExecutionComparisonFeature) bool {
	switch feature {
	case AVMComparisonFeatureExecution,
		AVMComparisonFeatureState,
		AVMComparisonFeatureMessaging,
		AVMComparisonFeatureScheduling,
		AVMComparisonFeatureContracts,
		AVMComparisonFeatureUXModel:
		return true
	default:
		return false
	}
}

func ComputeAVMExecutionComparisonRowHash(row AVMExecutionComparisonRow) string {
	row = canonicalAVMExecutionComparisonRow(row)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-execution-comparison-row-v1")
	writeEnginePart(h, string(row.Feature))
	writeEnginePart(h, row.ClassicCosmosSDK)
	writeEnginePart(h, row.AVM)
	writeEngineBool(h, row.RequiresAVMExtension)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMExecutionModelComparisonHash(model AVMExecutionModelComparison) string {
	model = canonicalAVMExecutionModelComparison(model)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-execution-model-comparison-v1")
	writeEnginePart(h, model.ModelName)
	writeEngineUint64(h, uint64(len(model.Rows)))
	for _, row := range model.Rows {
		writeEnginePart(h, row.RowHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMExecutionComparisonRow(row AVMExecutionComparisonRow) AVMExecutionComparisonRow {
	row.Feature = AVMExecutionComparisonFeature(strings.TrimSpace(string(row.Feature)))
	row.ClassicCosmosSDK = strings.TrimSpace(row.ClassicCosmosSDK)
	row.AVM = strings.TrimSpace(row.AVM)
	row.RowHash = strings.TrimSpace(row.RowHash)
	return row
}

func canonicalAVMExecutionModelComparison(model AVMExecutionModelComparison) AVMExecutionModelComparison {
	model.ModelName = strings.TrimSpace(model.ModelName)
	model.Rows = append([]AVMExecutionComparisonRow(nil), model.Rows...)
	for i := range model.Rows {
		model.Rows[i] = canonicalAVMExecutionComparisonRow(model.Rows[i])
	}
	sort.SliceStable(model.Rows, func(i, j int) bool {
		return model.Rows[i].Feature < model.Rows[j].Feature
	})
	model.MatrixHash = strings.TrimSpace(model.MatrixHash)
	return model
}

func sortAVMExecutionComparisonRowsByDocumentOrder(rows []AVMExecutionComparisonRow) []AVMExecutionComparisonRow {
	out := append([]AVMExecutionComparisonRow(nil), rows...)
	order := map[AVMExecutionComparisonFeature]int{
		AVMComparisonFeatureExecution:	0,
		AVMComparisonFeatureState:	1,
		AVMComparisonFeatureMessaging:	2,
		AVMComparisonFeatureScheduling:	3,
		AVMComparisonFeatureContracts:	4,
		AVMComparisonFeatureUXModel:	5,
	}
	sort.SliceStable(out, func(i, j int) bool {
		return order[out[i].Feature] < order[out[j].Feature]
	})
	return out
}

func validateAVMComparisonText(fieldName, value string) error {
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not have surrounding whitespace", fieldName)
	}
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(value) > MaxAVMComparisonTextBytes {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, MaxAVMComparisonTextBytes)
	}
	for _, r := range value {
		if r < 0x20 || r == '|' {
			return fmt.Errorf("%s contains invalid character", fieldName)
		}
	}
	return nil
}

func validateAVMComparisonHash(fieldName, value string) error {
	if len(value) != 64 {
		return fmt.Errorf("%s must be a sha256 hex hash", fieldName)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be a sha256 hex hash", fieldName)
	}
	return nil
}
