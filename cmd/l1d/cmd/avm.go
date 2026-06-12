package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
)

const (
	avmMsgService	= "l1.contracts.v1.Msg"
	avmQueryService	= "l1.contracts.v1.Query"

	flagAVMAuthority	= "authority"
	flagAVMBytecodeFile	= "bytecode-file"
	flagAVMBytecodeHex	= "bytecode-hex"
	flagAVMCodeHash		= "code-hash"
	flagAVMCodeBytes	= "code-bytes"
	flagAVMBodyJSON		= "body-json"
	flagAVMBodyFile		= "body-file"
	flagAVMBodyHex		= "body-hex"
	flagAVMInitialBalance	= "initial-balance"
	flagAVMAdmin		= "admin"
	flagAVMHeight		= "height"
	flagAVMNamespace	= "namespace"
	flagAVMSalt		= "salt"
	flagAVMFunds		= "funds"
	flagAVMGasLimit		= "gas-limit"
	flagAVMKeyPrefixHex	= "key-prefix-hex"
	flagAVMLimit		= "limit"
	flagAVMOpcode		= "opcode"
	flagAVMQueryID		= "query-id"
	flagAVMReceiptJSON	= "receipt-json"
	flagAVMReceiptFile	= "receipt-file"
)

type avmServicePayload struct {
	Service		string	`json:"service"`
	Method		string	`json:"method"`
	FullMethod	string	`json:"full_method"`
	TypeURL		string	`json:"type_url,omitempty"`
	Request		any	`json:"request"`
}

type avmStoreCodeRequest struct {
	Authority	string	`json:"authority"`
	CodeHash	string	`json:"code_hash,omitempty"`
	CodeBytes	uint64	`json:"code_bytes,omitempty"`
	Bytecode	string	`json:"bytecode_base64,omitempty"`
}

type avmDeployRequest struct {
	Creator		string	`json:"creator"`
	CodeID		string	`json:"code_id"`
	ChainID		string	`json:"chain_id,omitempty"`
	Namespace	string	`json:"namespace,omitempty"`
	Salt		string	`json:"salt,omitempty"`
	InitPayload	string	`json:"init_payload_base64,omitempty"`
	InitialBalance	uint64	`json:"initial_balance,omitempty"`
	Admin		string	`json:"admin,omitempty"`
	Height		uint64	`json:"height"`
}

type avmExecuteRequest struct {
	Sender		string	`json:"sender"`
	ContractAddress	string	`json:"contract_address"`
	Payload		string	`json:"payload_base64,omitempty"`
	Funds		uint64	`json:"funds,omitempty"`
	GasLimit	uint64	`json:"gas_limit"`
	Height		uint64	`json:"height"`
}

type avmQueryRequest struct {
	ContractAddress	string	`json:"contract_address,omitempty"`
	CodeID		string	`json:"code_id,omitempty"`
	KeyPrefix	string	`json:"key_prefix_base64,omitempty"`
	Limit		uint32	`json:"limit,omitempty"`
}

type avmDeployHint struct {
	ContractAddressUser	string	`json:"contract_address_user"`
	ContractAddressRaw	string	`json:"contract_address_raw"`
}

type avmExecuteHint struct {
	ExitCode	uint32	`json:"exit_code"`
	ReceiptID	string	`json:"receipt_id"`
}

func NewAVMTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"avm",
		Short:	"AVM contract transaction helpers",
	}
	cmd.AddCommand(
		newAVMStoreCodeCmd(),
		newAVMDeployCmd(),
		newAVMExecuteCmd(),
	)
	return cmd
}

func NewAVMQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"avm",
		Short:	"AVM contract query helpers",
	}
	cmd.AddCommand(
		newAVMCodeQueryCmd(),
		newAVMContractQueryCmd(),
		newAVMStorageQueryCmd(),
		newAVMReceiptsQueryCmd(),
	)
	return cmd
}

func NewAVMDebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"avm",
		Short:	"AVM developer debug helpers",
	}
	cmd.AddCommand(
		newAVMEncodeMessageCmd(),
		newAVMDecodeReceiptCmd(),
	)
	return cmd
}

func newAVMStoreCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"store-code",
		Short:	"Build l1.contracts.v1.Msg/StoreCode request",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateAVMTxFees(cmd); err != nil {
				return err
			}
			authority, err := avmAuthority(cmd)
			if err != nil {
				return err
			}
			bytecode, err := readOptionalBytes(cmd, flagAVMBytecodeFile, flagAVMBytecodeHex)
			if err != nil {
				return err
			}
			codeHash, _ := cmd.Flags().GetString(flagAVMCodeHash)
			codeBytes, _ := cmd.Flags().GetUint64(flagAVMCodeBytes)
			if len(bytecode) == 0 && (strings.TrimSpace(codeHash) == "" || codeBytes == 0) {
				return errors.New("store-code requires bytecode or code-hash plus code-bytes")
			}
			req := contractstypes.MsgStoreCode{
				Authority:	authority,
				CodeHash:	strings.TrimSpace(codeHash),
				CodeBytes:	codeBytes,
				Bytecode:	append([]byte(nil), bytecode...),
			}
			if len(req.Bytecode) > 0 {
				req.CodeBytes = uint64(len(req.Bytecode))
				if req.CodeHash == "" {
					req.CodeHash = contractstypes.CanonicalCodeHash(req.Bytecode)
				}
			}
			return writeCommandJSON(cmd, avmServicePayload{
				Service:	avmMsgService,
				Method:		"StoreCode",
				FullMethod:	"/" + avmMsgService + "/StoreCode",
				TypeURL:	contractstypes.MsgStoreCodeTypeURL,
				Request: avmStoreCodeRequest{
					Authority:	req.Authority,
					CodeHash:	req.CodeHash,
					CodeBytes:	req.CodeBytes,
					Bytecode:	base64OrEmpty(req.Bytecode),
				},
			})
		},
	}
	addAVMTxFlags(cmd)
	cmd.Flags().String(flagAVMAuthority, "", "governance/system authority; defaults to --from")
	cmd.Flags().String(flagAVMBytecodeFile, "", "AVM bytecode file")
	cmd.Flags().String(flagAVMBytecodeHex, "", "hex-encoded AVM bytecode")
	cmd.Flags().String(flagAVMCodeHash, "", "known AVM code hash")
	cmd.Flags().Uint64(flagAVMCodeBytes, 0, "known AVM code size in bytes")
	return cmd
}

func newAVMDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"deploy [code-id]",
		Short:	"Build l1.contracts.v1.Msg/DeployContract request",
		Args:	cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateAVMTxFees(cmd); err != nil {
				return err
			}
			creator, err := requiredFlag(cmd, flags.FlagFrom, "deploy creator")
			if err != nil {
				return err
			}
			body, err := readBodyBytes(cmd)
			if err != nil {
				return err
			}
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			namespace, _ := cmd.Flags().GetString(flagAVMNamespace)
			salt, _ := cmd.Flags().GetString(flagAVMSalt)
			initialBalance, _ := cmd.Flags().GetUint64(flagAVMInitialBalance)
			admin, _ := cmd.Flags().GetString(flagAVMAdmin)
			height, _ := cmd.Flags().GetUint64(flagAVMHeight)
			req := avmDeployRequest{
				Creator:	creator,
				CodeID:		strings.TrimSpace(args[0]),
				ChainID:	strings.TrimSpace(chainID),
				Namespace:	strings.TrimSpace(namespace),
				Salt:		strings.TrimSpace(salt),
				InitPayload:	base64OrEmpty(body),
				InitialBalance:	initialBalance,
				Admin:		strings.TrimSpace(admin),
				Height:		height,
			}
			if req.CodeID == "" {
				return errors.New("deploy code id is required")
			}
			if req.Height == 0 {
				return errors.New("deploy height must be positive")
			}
			return writeCommandJSON(cmd, struct {
				avmServicePayload
				ContractAddressUser	string		`json:"contract_address_user"`
				ContractAddressRaw	string		`json:"contract_address_raw"`
				Expected		avmDeployHint	`json:"expected_response_fields"`
			}{
				avmServicePayload: avmServicePayload{
					Service:	avmMsgService,
					Method:		"DeployContract",
					FullMethod:	"/" + avmMsgService + "/DeployContract",
					TypeURL:	contractstypes.MsgDeployContractTypeURL,
					Request:	req,
				},
				ContractAddressUser:	"AE...",
				ContractAddressRaw:	"4:...",
				Expected: avmDeployHint{
					ContractAddressUser:	"AE...",
					ContractAddressRaw:	"4:...",
				},
			})
		},
	}
	addAVMTxFlags(cmd)
	addAVMBodyFlags(cmd)
	cmd.Flags().String(flagAVMNamespace, "", "contract namespace")
	cmd.Flags().String(flagAVMSalt, "", "contract salt")
	cmd.Flags().Uint64(flagAVMInitialBalance, 0, "initial native balance in naet")
	cmd.Flags().String(flagAVMAdmin, "", "contract admin AE address")
	cmd.Flags().Uint64(flagAVMHeight, 1, "execution height")
	return cmd
}

func newAVMExecuteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"execute [contract-address]",
		Short:	"Build l1.contracts.v1.Msg/ExecuteExternal request",
		Args:	cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateAVMTxFees(cmd); err != nil {
				return err
			}
			sender, err := requiredFlag(cmd, flags.FlagFrom, "execute sender")
			if err != nil {
				return err
			}
			body, err := readBodyBytes(cmd)
			if err != nil {
				return err
			}
			funds, _ := cmd.Flags().GetUint64(flagAVMFunds)
			gasLimit, _ := cmd.Flags().GetUint64(flagAVMGasLimit)
			height, _ := cmd.Flags().GetUint64(flagAVMHeight)
			req := avmExecuteRequest{
				Sender:			sender,
				ContractAddress:	strings.TrimSpace(args[0]),
				Payload:		base64OrEmpty(body),
				Funds:			funds,
				GasLimit:		gasLimit,
				Height:			height,
			}
			if req.ContractAddress == "" {
				return errors.New("execute contract address is required")
			}
			if req.GasLimit == 0 {
				return errors.New("execute gas limit must be positive")
			}
			if req.Height == 0 {
				return errors.New("execute height must be positive")
			}
			return writeCommandJSON(cmd, struct {
				avmServicePayload
				ExitCode	uint32		`json:"exit_code"`
				ReceiptID	string		`json:"receipt_id"`
				Expected	avmExecuteHint	`json:"expected_response_fields"`
			}{
				avmServicePayload: avmServicePayload{
					Service:	avmMsgService,
					Method:		"ExecuteExternal",
					FullMethod:	"/" + avmMsgService + "/ExecuteExternal",
					TypeURL:	contractstypes.MsgExecuteExternalTypeURL,
					Request:	req,
				},
				ExitCode:	0,
				ReceiptID:	"receipt_id",
				Expected: avmExecuteHint{
					ExitCode:	0,
					ReceiptID:	"receipt_id",
				},
			})
		},
	}
	addAVMTxFlags(cmd)
	addAVMBodyFlags(cmd)
	cmd.Flags().Uint64(flagAVMFunds, 0, "native funds in naet")
	cmd.Flags().Uint64(flagAVMGasLimit, 100_000, "AVM gas limit")
	cmd.Flags().Uint64(flagAVMHeight, 1, "execution height")
	return cmd
}

func newAVMCodeQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"code [code-id]",
		Short:	"Build l1.contracts.v1.Query/Code request",
		Args:	cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeCommandJSON(cmd, avmServicePayload{
				Service:	avmQueryService,
				Method:		"Code",
				FullMethod:	"/" + avmQueryService + "/Code",
				Request:	avmQueryRequest{CodeID: strings.TrimSpace(args[0])},
			})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newAVMContractQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"contract [contract-address]",
		Short:	"Build l1.contracts.v1.Query/Contract request",
		Args:	cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeCommandJSON(cmd, avmServicePayload{
				Service:	avmQueryService,
				Method:		"Contract",
				FullMethod:	"/" + avmQueryService + "/Contract",
				Request:	avmQueryRequest{ContractAddress: strings.TrimSpace(args[0])},
			})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newAVMStorageQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"storage [contract-address]",
		Short:	"Build l1.contracts.v1.Query/ContractStorage request",
		Args:	cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyPrefix, err := optionalHexFlag(cmd, flagAVMKeyPrefixHex)
			if err != nil {
				return err
			}
			limit, _ := cmd.Flags().GetUint32(flagAVMLimit)
			if limit == 0 {
				return errors.New("storage query limit must be positive")
			}
			return writeCommandJSON(cmd, avmServicePayload{
				Service:	avmQueryService,
				Method:		"ContractStorage",
				FullMethod:	"/" + avmQueryService + "/ContractStorage",
				Request:	avmQueryRequest{ContractAddress: strings.TrimSpace(args[0]), KeyPrefix: base64OrEmpty(keyPrefix), Limit: limit},
			})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(flagAVMKeyPrefixHex, "", "hex-encoded storage key prefix")
	cmd.Flags().Uint32(flagAVMLimit, 50, "bounded storage query limit")
	return cmd
}

func newAVMReceiptsQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"receipts [contract-address]",
		Short:	"Build l1.contracts.v1.Query/ContractReceipts request",
		Args:	cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, _ := cmd.Flags().GetUint32(flagAVMLimit)
			if limit == 0 {
				return errors.New("receipts query limit must be positive")
			}
			return writeCommandJSON(cmd, avmServicePayload{
				Service:	avmQueryService,
				Method:		"ContractReceipts",
				FullMethod:	"/" + avmQueryService + "/ContractReceipts",
				Request:	avmQueryRequest{ContractAddress: strings.TrimSpace(args[0]), Limit: limit},
			})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Uint32(flagAVMLimit, 50, "bounded receipts query limit")
	return cmd
}

func newAVMEncodeMessageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"encode-message",
		Short:	"Encode an AVM message body from JSON",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := readBodyBytes(cmd)
			if err != nil {
				return err
			}
			opcode, _ := cmd.Flags().GetUint32(flagAVMOpcode)
			queryID, _ := cmd.Flags().GetUint64(flagAVMQueryID)
			sum := sha256.Sum256(body)
			return writeCommandJSON(cmd, map[string]any{
				"body_base64":	base64.StdEncoding.EncodeToString(body),
				"body_hex":	hex.EncodeToString(body),
				"body_sha256":	hex.EncodeToString(sum[:]),
				"opcode":	opcode,
				"query_id":	queryID,
			})
		},
	}
	addAVMBodyFlags(cmd)
	cmd.Flags().Uint32(flagAVMOpcode, 0, "AVM opcode")
	cmd.Flags().Uint64(flagAVMQueryID, 0, "AVM query id")
	return cmd
}

func newAVMDecodeReceiptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"decode-receipt",
		Short:	"Normalize an AVM execution receipt into stable JSON",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			raw, err := readReceiptJSON(cmd)
			if err != nil {
				return err
			}
			var receipt async.ExecutionReceipt
			if err := json.Unmarshal(raw, &receipt); err != nil {
				return fmt.Errorf("decode receipt JSON: %w", err)
			}
			id := avmReceiptID(receipt)
			return writeCommandJSON(cmd, map[string]any{
				"receipt_id":		id,
				"sequence":		receipt.Sequence,
				"source":		receipt.Source.String(),
				"destination":		receipt.Destination.String(),
				"opcode":		receipt.Opcode,
				"query_id":		receipt.QueryID,
				"exit_code":		receipt.ResultCode,
				"gas_used":		receipt.GasUsed,
				"storage_fee_naet":	receipt.StorageFeeNaet.String(),
				"forward_fee_naet":	receipt.ForwardFeeNaet.String(),
				"bounced":		receipt.Bounced,
				"retry_count":		receipt.RetryCount,
				"retry_scheduled":	receipt.RetryScheduled,
				"error":		receipt.Error,
			})
		},
	}
	cmd.Flags().String(flagAVMReceiptJSON, "", "receipt JSON")
	cmd.Flags().String(flagAVMReceiptFile, "", "receipt JSON file")
	return cmd
}

func addAVMTxFlags(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)
}

func addAVMBodyFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagAVMBodyJSON, "", "JSON message body")
	cmd.Flags().String(flagAVMBodyFile, "", "file containing JSON message body")
	cmd.Flags().String(flagAVMBodyHex, "", "hex-encoded message body")
}

func validateAVMTxFees(cmd *cobra.Command) error {
	feeText, _ := cmd.Flags().GetString(flags.FlagFees)
	feeText = strings.TrimSpace(feeText)
	if feeText == "" {
		return errors.New("AVM tx requires --fees with naet denom")
	}
	coins, err := sdk.ParseCoinsNormalized(feeText)
	if err != nil {
		return fmt.Errorf("invalid AVM tx fees: %w", err)
	}
	if coins.Empty() {
		return errors.New("AVM tx fees must not be empty")
	}
	for _, coin := range coins {
		if coin.Denom != appparams.BaseDenom {
			return fmt.Errorf("AVM tx fees must use %s denom", appparams.BaseDenom)
		}
	}
	return nil
}

func avmAuthority(cmd *cobra.Command) (string, error) {
	authority, _ := cmd.Flags().GetString(flagAVMAuthority)
	authority = strings.TrimSpace(authority)
	if authority != "" {
		return authority, nil
	}
	return requiredFlag(cmd, flags.FlagFrom, "store code authority")
}

func requiredFlag(cmd *cobra.Command, name string, label string) (string, error) {
	value, _ := cmd.Flags().GetString(name)
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", label)
	}
	return value, nil
}

func readBodyBytes(cmd *cobra.Command) ([]byte, error) {
	bodyJSON, _ := cmd.Flags().GetString(flagAVMBodyJSON)
	bodyFile, _ := cmd.Flags().GetString(flagAVMBodyFile)
	bodyHex, _ := cmd.Flags().GetString(flagAVMBodyHex)
	set := countSet(bodyJSON, bodyFile, bodyHex)
	if set == 0 {
		return nil, nil
	}
	if set > 1 {
		return nil, errors.New("set only one of body-json, body-file, or body-hex")
	}
	if strings.TrimSpace(bodyHex) != "" {
		return hex.DecodeString(strings.TrimSpace(bodyHex))
	}
	if strings.TrimSpace(bodyFile) != "" {
		bz, err := os.ReadFile(strings.TrimSpace(bodyFile))
		if err != nil {
			return nil, err
		}
		return canonicalJSONBytes(bz)
	}
	return canonicalJSONBytes([]byte(bodyJSON))
}

func readOptionalBytes(cmd *cobra.Command, fileFlag string, hexFlag string) ([]byte, error) {
	fileName, _ := cmd.Flags().GetString(fileFlag)
	hexText, _ := cmd.Flags().GetString(hexFlag)
	if strings.TrimSpace(fileName) != "" && strings.TrimSpace(hexText) != "" {
		return nil, fmt.Errorf("set only one of %s or %s", fileFlag, hexFlag)
	}
	if strings.TrimSpace(fileName) != "" {
		return os.ReadFile(strings.TrimSpace(fileName))
	}
	if strings.TrimSpace(hexText) != "" {
		return hex.DecodeString(strings.TrimSpace(hexText))
	}
	return nil, nil
}

func optionalHexFlag(cmd *cobra.Command, name string) ([]byte, error) {
	value, _ := cmd.Flags().GetString(name)
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	return hex.DecodeString(value)
}

func canonicalJSONBytes(raw []byte) ([]byte, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, errors.New("JSON body is empty")
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("invalid JSON body: %w", err)
	}
	return json.Marshal(value)
}

func readReceiptJSON(cmd *cobra.Command) ([]byte, error) {
	text, _ := cmd.Flags().GetString(flagAVMReceiptJSON)
	fileName, _ := cmd.Flags().GetString(flagAVMReceiptFile)
	if strings.TrimSpace(text) != "" && strings.TrimSpace(fileName) != "" {
		return nil, errors.New("set only one of receipt-json or receipt-file")
	}
	if strings.TrimSpace(fileName) != "" {
		return os.ReadFile(strings.TrimSpace(fileName))
	}
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("receipt JSON is required")
	}
	return []byte(text), nil
}

func avmReceiptID(receipt async.ExecutionReceipt) string {
	buf := bytes.NewBuffer(nil)
	_, _ = fmt.Fprintf(buf, "%020d/%s/%s/%010d/%020d/%010d/%020d/%s",
		receipt.Sequence,
		receipt.Source.String(),
		receipt.Destination.String(),
		receipt.Opcode,
		receipt.QueryID,
		receipt.ResultCode,
		receipt.GasUsed,
		receipt.Error,
	)
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:])
}

func writeCommandJSON(cmd *cobra.Command, value any) error {
	bz, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(bz))
	return err
}

func base64OrEmpty(bz []byte) string {
	if len(bz) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(bz)
}

func countSet(values ...string) int {
	var count int
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			count++
		}
	}
	return count
}
