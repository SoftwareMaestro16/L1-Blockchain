package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
)

func TestAVMCLICommandConstruction(t *testing.T) {
	for _, tc := range []struct {
		name	string
		root	*cobraCommandShim
		path	[]string
	}{
		{name: "tx store-code", root: shimCommand(NewAVMTxCmd()), path: []string{"store-code"}},
		{name: "tx deploy", root: shimCommand(NewAVMTxCmd()), path: []string{"deploy"}},
		{name: "tx execute", root: shimCommand(NewAVMTxCmd()), path: []string{"execute"}},
		{name: "query code", root: shimCommand(NewAVMQueryCmd()), path: []string{"code"}},
		{name: "query contract", root: shimCommand(NewAVMQueryCmd()), path: []string{"contract"}},
		{name: "query storage", root: shimCommand(NewAVMQueryCmd()), path: []string{"storage"}},
		{name: "query receipts", root: shimCommand(NewAVMQueryCmd()), path: []string{"receipts"}},
		{name: "debug encode-message", root: shimCommand(NewAVMDebugCmd()), path: []string{"encode-message"}},
		{name: "debug decode-receipt", root: shimCommand(NewAVMDebugCmd()), path: []string{"decode-receipt"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.root.Find(t, tc.path...)
			require.NotEmpty(t, cmd.Short)
		})
	}
}

func TestAVMCLIFeeValidationRejectsMissingOrNonNaetFees(t *testing.T) {
	_, err := executeAVMCommand(NewAVMTxCmd(), "deploy", testHash("code"), "--from", aeAddressForCLI(0x11), "--height", "1")
	require.ErrorContains(t, err, "requires --fees")

	_, err = executeAVMCommand(NewAVMTxCmd(), "deploy", testHash("code"), "--from", aeAddressForCLI(0x11), "--height", "1", "--fees", "1uatom")
	require.ErrorContains(t, err, "must use naet denom")

	out, err := executeAVMCommand(NewAVMTxCmd(), "deploy", testHash("code"), "--from", aeAddressForCLI(0x11), "--height", "1", "--fees", "1"+appparams.BaseDenom)
	require.NoError(t, err)
	require.Contains(t, out, "DeployContract")
}

func TestAVMCLIE2ESmokeDeployExecuteQuery(t *testing.T) {
	codeID := testHash("smoke-code")
	creator := aeAddressForCLI(0x11)
	contract := aeAddressForCLI(0x22)

	deployOut, err := executeAVMCommand(
		NewAVMTxCmd(),
		"deploy", codeID,
		"--from", creator,
		"--height", "9",
		"--fees", "7"+appparams.BaseDenom,
		"--body-json", `{"symbol":"TST","decimals":9}`,
		"--salt", "token-master",
	)
	require.NoError(t, err)
	var deploy struct {
		Service	string	`json:"service"`
		Method	string	`json:"method"`
		TypeURL	string	`json:"type_url"`
		Request	struct {
			Creator		string	`json:"creator"`
			CodeID		string	`json:"code_id"`
			InitPayload	string	`json:"init_payload_base64"`
			Height		uint64	`json:"height"`
		}	`json:"request"`
		Expected	struct {
			ContractAddressUser	string	`json:"contract_address_user"`
			ContractAddressRaw	string	`json:"contract_address_raw"`
		}	`json:"expected_response_fields"`
	}
	require.NoError(t, json.Unmarshal([]byte(deployOut), &deploy), deployOut)
	require.Equal(t, "l1.contracts.v1.Msg", deploy.Service)
	require.Equal(t, "DeployContract", deploy.Method)
	require.Equal(t, codeID, deploy.Request.CodeID)
	require.Equal(t, creator, deploy.Request.Creator)
	require.Equal(t, uint64(9), deploy.Request.Height)
	require.Equal(t, "AE...", deploy.Expected.ContractAddressUser)
	require.Equal(t, "4:...", deploy.Expected.ContractAddressRaw)
	body, err := base64.StdEncoding.DecodeString(deploy.Request.InitPayload)
	require.NoError(t, err)
	require.Equal(t, `{"decimals":9,"symbol":"TST"}`, string(body))

	execOut, err := executeAVMCommand(
		NewAVMTxCmd(),
		"execute", contract,
		"--from", creator,
		"--height", "10",
		"--gas-limit", "500000",
		"--fees", "3"+appparams.BaseDenom,
		"--body-json", `{"op":"mint","amount":"100"}`,
	)
	require.NoError(t, err)
	require.Contains(t, execOut, "ExecuteExternal")
	require.Contains(t, execOut, "receipt_id")
	require.Contains(t, execOut, `"exit_code": 0`)

	queryOut, err := executeAVMCommand(NewAVMQueryCmd(), "contract", contract)
	require.NoError(t, err)
	require.Contains(t, queryOut, "l1.contracts.v1.Query")
	require.Contains(t, queryOut, "Contract")
	require.Contains(t, queryOut, contract)

	storageOut, err := executeAVMCommand(NewAVMQueryCmd(), "storage", contract, "--key-prefix-hex", "01", "--limit", "5")
	require.NoError(t, err)
	require.Contains(t, storageOut, "ContractStorage")
	require.Contains(t, storageOut, `"limit": 5`)
}

func TestAVMCLIDecodeReceiptStableJSON(t *testing.T) {
	receipt := async.ExecutionReceipt{
		Sequence:	7,
		Opcode:		99,
		QueryID:	42,
		ResultCode:	async.ResultOK,
		GasUsed:	1234,
		StorageFeeNaet:	sdkmath.NewInt(5),
		ForwardFeeNaet:	sdkmath.NewInt(3),
	}
	bz, err := json.Marshal(receipt)
	require.NoError(t, err)

	first, err := executeAVMCommand(NewAVMDebugCmd(), "decode-receipt", "--receipt-json", string(bz))
	require.NoError(t, err)
	second, err := executeAVMCommand(NewAVMDebugCmd(), "decode-receipt", "--receipt-json", string(bz))
	require.NoError(t, err)
	require.Equal(t, first, second)

	var decoded struct {
		ReceiptID	string	`json:"receipt_id"`
		ExitCode	uint32	`json:"exit_code"`
		GasUsed		uint64	`json:"gas_used"`
		StorageFeeNaet	string	`json:"storage_fee_naet"`
		ForwardFeeNaet	string	`json:"forward_fee_naet"`
		RetryScheduled	bool	`json:"retry_scheduled"`
	}
	require.NoError(t, json.Unmarshal([]byte(first), &decoded), first)
	require.NotEmpty(t, decoded.ReceiptID)
	require.Equal(t, uint32(0), decoded.ExitCode)
	require.Equal(t, uint64(1234), decoded.GasUsed)
	require.Equal(t, "5", decoded.StorageFeeNaet)
	require.Equal(t, "3", decoded.ForwardFeeNaet)
	require.False(t, decoded.RetryScheduled)
}

func TestAVMCLIEncodeMessageCanonicalizesJSON(t *testing.T) {
	out, err := executeAVMCommand(NewAVMDebugCmd(), "encode-message", "--opcode", "12", "--query-id", "77", "--body-json", `{"b":2,"a":1}`)
	require.NoError(t, err)

	var decoded struct {
		BodyBase64	string	`json:"body_base64"`
		Opcode		uint32	`json:"opcode"`
		QueryID		uint64	`json:"query_id"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &decoded), out)
	body, err := base64.StdEncoding.DecodeString(decoded.BodyBase64)
	require.NoError(t, err)
	require.Equal(t, `{"a":1,"b":2}`, string(body))
	require.Equal(t, uint32(12), decoded.Opcode)
	require.Equal(t, uint64(77), decoded.QueryID)
}

type cobraCommandShim struct {
	cmd interface {
		Find([]string) (*cobra.Command, []string, error)
	}
}

func shimCommand(cmd *cobra.Command) *cobraCommandShim {
	return &cobraCommandShim{cmd: cmd}
}

func (s *cobraCommandShim) Find(t *testing.T, path ...string) *cobra.Command {
	t.Helper()
	cmd, _, err := s.cmd.Find(path)
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, path[len(path)-1], cmd.Name())
	return cmd
}

func executeAVMCommand(cmd *cobra.Command, args ...string) (string, error) {
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func aeAddressForCLI(fill byte) string {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = fill
	}
	return addressing.FormatAccAddress(bz)
}

func testHash(seed string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("avm-cli/%s", seed)))
	return hex.EncodeToString(sum[:])
}
