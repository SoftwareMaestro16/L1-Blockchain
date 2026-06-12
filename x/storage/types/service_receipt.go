package types

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

type ServiceStorageReceipt struct {
	ReceiptID	string
	ServiceID	string
	ObjectID	string
	RequestHash	string
	ContentHash	string
	ProviderID	string
	AccessHeight	uint64
	Signature	string
	ProofOptional	string
	ReceiptHash	string
}

type ServiceStorageReceiptAnchor struct {
	ServiceID	string
	Receipts	[]ServiceStorageReceipt
	Height		uint64
	RootHash	string
	AnchorHash	string
}

type OnChainStorageFeeModel struct {
	Denom		string
	PricePerByte	uint64
	MinimumFee	uint64
	MaxPayloadBytes	uint64
	ModelHash	string
}

type OnChainStorageFeeQuote struct {
	DeclarationHash	string
	PayloadBytes	uint64
	Denom		string
	FeeAmount	uint64
	QuoteHash	string
}

func NewServiceStorageReceipt(receipt ServiceStorageReceipt) (ServiceStorageReceipt, error) {
	if receipt.ReceiptHash != "" {
		return ServiceStorageReceipt{}, errors.New("service storage receipt hash must be empty before construction")
	}
	receipt = canonicalServiceStorageReceipt(receipt)
	if receipt.ReceiptID == "" {
		receipt.ReceiptID = ComputeServiceStorageReceiptID(receipt)
	}
	if err := receipt.ValidateFormat(); err != nil {
		return ServiceStorageReceipt{}, err
	}
	receipt.ReceiptHash = ComputeServiceStorageReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func AnchorServiceStorageReceipts(serviceID string, receipts []ServiceStorageReceipt, height uint64) (ServiceStorageReceiptAnchor, error) {
	if err := validateStorageToken("service storage anchor service id", serviceID); err != nil {
		return ServiceStorageReceiptAnchor{}, err
	}
	anchor := ServiceStorageReceiptAnchor{
		ServiceID:	serviceID,
		Receipts:	normalizeServiceStorageReceipts(receipts),
		Height:		height,
	}
	if err := anchor.ValidateFormat(); err != nil {
		return ServiceStorageReceiptAnchor{}, err
	}
	anchor.RootHash = ComputeServiceStorageReceiptRoot(anchor.Receipts)
	anchor.AnchorHash = ComputeServiceStorageReceiptAnchorHash(anchor)
	return anchor, anchor.Validate()
}

func VerifyServiceStorageReceiptProof(receipt ServiceStorageReceipt, object StorageObject, proof StorageRetrievalProof) error {
	if err := receipt.Validate(); err != nil {
		return err
	}
	if err := object.Validate(); err != nil {
		return err
	}
	if err := proof.Validate(); err != nil {
		return err
	}
	if receipt.ObjectID != object.ObjectID || receipt.ObjectID != proof.ObjectID {
		return errors.New("service storage receipt proof object mismatch")
	}
	if receipt.ContentHash != object.ContentHash || receipt.ContentHash != proof.ContentHash {
		return errors.New("service storage receipt proof content hash mismatch")
	}
	if receipt.ProofOptional == "" {
		return errors.New("service storage receipt proof is required for verification hook")
	}
	if receipt.ProofOptional != proof.ProofHash {
		return errors.New("service storage receipt proof hash mismatch")
	}
	return nil
}

func NewOnChainStorageFeeModel(model OnChainStorageFeeModel) (OnChainStorageFeeModel, error) {
	if model.ModelHash != "" {
		return OnChainStorageFeeModel{}, errors.New("on-chain storage fee model hash must be empty before construction")
	}
	model = canonicalOnChainStorageFeeModel(model)
	if err := model.ValidateFormat(); err != nil {
		return OnChainStorageFeeModel{}, err
	}
	model.ModelHash = ComputeOnChainStorageFeeModelHash(model)
	return model, model.Validate()
}

func QuoteOnChainStorageFee(declaration StorageDeclaration, model OnChainStorageFeeModel, payloadBytes uint64) (OnChainStorageFeeQuote, error) {
	if err := declaration.Validate(); err != nil {
		return OnChainStorageFeeQuote{}, err
	}
	if err := model.Validate(); err != nil {
		return OnChainStorageFeeQuote{}, err
	}
	if declaration.StorageModel != StorageModelPersistentOnChain {
		return OnChainStorageFeeQuote{}, errors.New("on-chain storage fee quote requires persistent on-chain storage declaration")
	}
	if payloadBytes == 0 {
		return OnChainStorageFeeQuote{}, errors.New("on-chain storage fee quote payload bytes must be positive")
	}
	if payloadBytes > declaration.MaxPayloadBytes || payloadBytes > model.MaxPayloadBytes {
		return OnChainStorageFeeQuote{}, errors.New("on-chain storage fee quote payload exceeds max payload bytes")
	}
	if payloadBytes > math.MaxUint64/model.PricePerByte {
		return OnChainStorageFeeQuote{}, errors.New("on-chain storage fee quote overflow")
	}
	fee := payloadBytes * model.PricePerByte
	if fee < model.MinimumFee {
		fee = model.MinimumFee
	}
	quote := OnChainStorageFeeQuote{
		DeclarationHash:	declaration.DeclarationHash,
		PayloadBytes:		payloadBytes,
		Denom:			model.Denom,
		FeeAmount:		fee,
	}
	quote.QuoteHash = ComputeOnChainStorageFeeQuoteHash(quote)
	return quote, quote.Validate()
}

func (receipt ServiceStorageReceipt) ValidateFormat() error {
	receipt = canonicalServiceStorageReceipt(receipt)
	if err := validateStorageToken("service storage receipt id", receipt.ReceiptID); err != nil {
		return err
	}
	if err := validateStorageToken("service storage receipt service id", receipt.ServiceID); err != nil {
		return err
	}
	if err := validateStorageToken("service storage receipt object id", receipt.ObjectID); err != nil {
		return err
	}
	if err := validateStorageHash("service storage receipt request hash", receipt.RequestHash); err != nil {
		return err
	}
	if err := validateStorageHash("service storage receipt content hash", receipt.ContentHash); err != nil {
		return err
	}
	if err := validateStorageToken("service storage receipt provider id", receipt.ProviderID); err != nil {
		return err
	}
	if receipt.AccessHeight == 0 {
		return errors.New("service storage receipt access height must be positive")
	}
	if err := validateStorageHash("service storage receipt signature", receipt.Signature); err != nil {
		return err
	}
	if receipt.ProofOptional != "" {
		if err := validateStorageHash("service storage receipt proof", receipt.ProofOptional); err != nil {
			return err
		}
	}
	if receipt.ReceiptHash != "" {
		return validateStorageHash("service storage receipt hash", receipt.ReceiptHash)
	}
	return nil
}

func (receipt ServiceStorageReceipt) Validate() error {
	receipt = canonicalServiceStorageReceipt(receipt)
	if err := receipt.ValidateFormat(); err != nil {
		return err
	}
	if receipt.ReceiptHash == "" {
		return errors.New("service storage receipt hash is required")
	}
	if receipt.ReceiptHash != ComputeServiceStorageReceiptHash(receipt) {
		return errors.New("service storage receipt hash mismatch")
	}
	return nil
}

func (anchor ServiceStorageReceiptAnchor) ValidateFormat() error {
	anchor.ServiceID = strings.TrimSpace(anchor.ServiceID)
	if err := validateStorageToken("service storage anchor service id", anchor.ServiceID); err != nil {
		return err
	}
	if len(anchor.Receipts) == 0 {
		return errors.New("service storage anchor requires receipts")
	}
	if anchor.Height == 0 {
		return errors.New("service storage anchor height must be positive")
	}
	previous := ""
	for _, receipt := range anchor.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if receipt.ServiceID != anchor.ServiceID {
			return errors.New("service storage anchor receipt service mismatch")
		}
		if previous != "" && previous >= receipt.ReceiptID {
			return errors.New("service storage anchor receipts must be sorted canonically")
		}
		previous = receipt.ReceiptID
	}
	if anchor.RootHash != "" {
		if err := validateStorageHash("service storage anchor root", anchor.RootHash); err != nil {
			return err
		}
	}
	if anchor.AnchorHash != "" {
		if err := validateStorageHash("service storage anchor hash", anchor.AnchorHash); err != nil {
			return err
		}
	}
	return nil
}

func (anchor ServiceStorageReceiptAnchor) Validate() error {
	if err := anchor.ValidateFormat(); err != nil {
		return err
	}
	if anchor.RootHash == "" || anchor.AnchorHash == "" {
		return errors.New("service storage anchor hashes are required")
	}
	if anchor.RootHash != ComputeServiceStorageReceiptRoot(anchor.Receipts) {
		return errors.New("service storage anchor root mismatch")
	}
	if anchor.AnchorHash != ComputeServiceStorageReceiptAnchorHash(anchor) {
		return errors.New("service storage anchor hash mismatch")
	}
	return nil
}

func (model OnChainStorageFeeModel) ValidateFormat() error {
	model = canonicalOnChainStorageFeeModel(model)
	if err := validateStorageToken("on-chain storage fee denom", model.Denom); err != nil {
		return err
	}
	if model.PricePerByte == 0 {
		return errors.New("on-chain storage fee price per byte must be positive")
	}
	if model.MaxPayloadBytes == 0 || model.MaxPayloadBytes > MaxStorageObjectSize {
		return fmt.Errorf("on-chain storage fee max payload bytes must be between 1 and %d", MaxStorageObjectSize)
	}
	if model.ModelHash != "" {
		return validateStorageHash("on-chain storage fee model hash", model.ModelHash)
	}
	return nil
}

func (model OnChainStorageFeeModel) Validate() error {
	model = canonicalOnChainStorageFeeModel(model)
	if err := model.ValidateFormat(); err != nil {
		return err
	}
	if model.ModelHash == "" {
		return errors.New("on-chain storage fee model hash is required")
	}
	if model.ModelHash != ComputeOnChainStorageFeeModelHash(model) {
		return errors.New("on-chain storage fee model hash mismatch")
	}
	return nil
}

func (quote OnChainStorageFeeQuote) Validate() error {
	if err := validateStorageHash("on-chain storage fee quote declaration hash", quote.DeclarationHash); err != nil {
		return err
	}
	if quote.PayloadBytes == 0 {
		return errors.New("on-chain storage fee quote payload bytes must be positive")
	}
	if err := validateStorageToken("on-chain storage fee quote denom", quote.Denom); err != nil {
		return err
	}
	if quote.FeeAmount == 0 {
		return errors.New("on-chain storage fee quote amount must be positive")
	}
	if err := validateStorageHash("on-chain storage fee quote hash", quote.QuoteHash); err != nil {
		return err
	}
	if quote.QuoteHash != ComputeOnChainStorageFeeQuoteHash(quote) {
		return errors.New("on-chain storage fee quote hash mismatch")
	}
	return nil
}

func ComputeServiceStorageReceiptID(receipt ServiceStorageReceipt) string {
	receipt = canonicalServiceStorageReceipt(receipt)
	return "service-receipt/" + storageHashParts("service-storage-receipt-id-v1", receipt.ServiceID, receipt.ObjectID, receipt.RequestHash, receipt.ProviderID, fmt.Sprintf("%020d", receipt.AccessHeight))
}

func ComputeServiceStorageReceiptHash(receipt ServiceStorageReceipt) string {
	receipt = canonicalServiceStorageReceipt(receipt)
	return storageHashParts(
		"service-storage-receipt-v1",
		receipt.ReceiptID,
		receipt.ServiceID,
		receipt.ObjectID,
		receipt.RequestHash,
		receipt.ContentHash,
		receipt.ProviderID,
		fmt.Sprintf("%020d", receipt.AccessHeight),
		receipt.Signature,
		receipt.ProofOptional,
	)
}

func ComputeServiceStorageReceiptRoot(receipts []ServiceStorageReceipt) string {
	ordered := normalizeServiceStorageReceipts(receipts)
	parts := []string{"service-storage-receipt-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, receipt.ReceiptHash)
	}
	return storageHashParts(parts...)
}

func ComputeServiceStorageReceiptAnchorHash(anchor ServiceStorageReceiptAnchor) string {
	return storageHashParts("service-storage-receipt-anchor-v1", anchor.ServiceID, anchor.RootHash, fmt.Sprintf("%020d", anchor.Height))
}

func ComputeOnChainStorageFeeModelHash(model OnChainStorageFeeModel) string {
	model = canonicalOnChainStorageFeeModel(model)
	return storageHashParts("on-chain-storage-fee-model-v1", model.Denom, fmt.Sprintf("%020d", model.PricePerByte), fmt.Sprintf("%020d", model.MinimumFee), fmt.Sprintf("%020d", model.MaxPayloadBytes))
}

func ComputeOnChainStorageFeeQuoteHash(quote OnChainStorageFeeQuote) string {
	return storageHashParts("on-chain-storage-fee-quote-v1", quote.DeclarationHash, fmt.Sprintf("%020d", quote.PayloadBytes), quote.Denom, fmt.Sprintf("%020d", quote.FeeAmount))
}

func canonicalServiceStorageReceipt(receipt ServiceStorageReceipt) ServiceStorageReceipt {
	receipt.ReceiptID = strings.TrimSpace(receipt.ReceiptID)
	receipt.ServiceID = strings.TrimSpace(receipt.ServiceID)
	receipt.ObjectID = strings.TrimSpace(receipt.ObjectID)
	receipt.RequestHash = strings.ToLower(strings.TrimSpace(receipt.RequestHash))
	receipt.ContentHash = strings.ToLower(strings.TrimSpace(receipt.ContentHash))
	receipt.ProviderID = strings.TrimSpace(receipt.ProviderID)
	receipt.Signature = strings.ToLower(strings.TrimSpace(receipt.Signature))
	receipt.ProofOptional = strings.ToLower(strings.TrimSpace(receipt.ProofOptional))
	receipt.ReceiptHash = strings.ToLower(strings.TrimSpace(receipt.ReceiptHash))
	return receipt
}

func canonicalOnChainStorageFeeModel(model OnChainStorageFeeModel) OnChainStorageFeeModel {
	model.Denom = strings.TrimSpace(model.Denom)
	model.ModelHash = strings.ToLower(strings.TrimSpace(model.ModelHash))
	return model
}

func normalizeServiceStorageReceipts(receipts []ServiceStorageReceipt) []ServiceStorageReceipt {
	out := append([]ServiceStorageReceipt(nil), receipts...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReceiptID < out[j].ReceiptID })
	return out
}
