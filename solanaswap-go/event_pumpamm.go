package solanaswapgo

import (
	"fmt"
	"strconv"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

type InputTransfer struct {
	TransferData
}

type OutputTransfer struct {
	TransferData
}

type PumpAmmPool struct {
	Pool                             solana.PublicKey
	GlobalConfig                     solana.PublicKey
	BaseMint                         solana.PublicKey
	QuoteMint                        solana.PublicKey
	PoolBaseTokenAccount             solana.PublicKey
	PoolQuoteTokenAccount            solana.PublicKey
	ProtocolFeeRecipient             solana.PublicKey
	ProtocolFeeRecipientTokenAccount solana.PublicKey
	CoinCreatorVaultAta              solana.PublicKey
	CoinCreatorVaultAuthority        solana.PublicKey
	PoolBaseTokenReserves            uint64
	PoolQuoteTokenReserves           uint64
}

type PumpAmmEvent struct {
	Timestamp                        int64            `json:"timestamp,string"`
	BaseAmountOut                    uint64           `json:"baseAmountOut,string"`
	MaxQuoteAmountIn                 uint64           `json:"maxQuoteAmountIn,string"`
	UserBaseTokenReserves            uint64           `json:"userBaseTokenReserves,string"`
	UserQuoteTokenReserves           uint64           `json:"userQuoteTokenReserves,string"`
	PoolBaseTokenReserves            uint64           `json:"poolBaseTokenReserves,string"`
	PoolQuoteTokenReserves           uint64           `json:"poolQuoteTokenReserves,string"`
	QuoteAmountIn                    uint64           `json:"quoteAmountIn,string"`
	LpFeeBasisPoints                 uint64           `json:"lpFeeBasisPoints,string"`
	LpFee                            uint64           `json:"lpFee,string"`
	ProtocolFeeBasisPoints           uint64           `json:"protocolFeeBasisPoints,string"`
	ProtocolFee                      uint64           `json:"protocolFee,string"`
	QuoteAmountInWithLpFee           uint64           `json:"quoteAmountInWithLpFee,string"`
	UserQuoteAmountIn                uint64           `json:"userQuoteAmountIn,string"`
	Pool                             solana.PublicKey `json:"pool"`
	User                             solana.PublicKey `json:"user"`
	UserBaseTokenAccount             solana.PublicKey `json:"userBaseTokenAccount"`
	UserQuoteTokenAccount            solana.PublicKey `json:"userQuoteTokenAccount"`
	ProtocolFeeRecipient             solana.PublicKey `json:"protocolFeeRecipient"`
	ProtocolFeeRecipientTokenAccount solana.PublicKey `json:"protocolFeeRecipientTokenAccount"`
}

func (p *Parser) processPumpAmmSwaps(instructionIndex int) []SwapData {
	var swaps []SwapData

	innerInstructions := p.getInnerInstructions(instructionIndex)
	if len(innerInstructions) == 0 {
		return swaps
	}

	userAccount := p.allAccountKeys[0]

	for _, inner := range innerInstructions {
		switch {
		case p.isTransferCheck(inner):
			swaps = append(swaps, p.processTransferCheckInstruction(userAccount.String(), inner)...)
		case p.isTokenTransfer(inner):
			swaps = append(swaps, p.processTokenTransferInstruction(userAccount.String(), inner)...)
		case p.isSystemTransfer(inner):
			swaps = append(swaps, p.processSystemTransferInstruction(userAccount.String(), inner)...)
		}
	}

	return swaps
}

func (p *Parser) processTransferCheckInstruction(userAccount string, inner solana.CompiledInstruction) []SwapData {
	var swaps []SwapData
	transfer := p.processTransferCheck(inner)
	if transfer == nil {
		return swaps
	}
	authority := transfer.Info.Authority
	amountStr := transfer.Info.TokenAmount.Amount
	amount, _ := strconv.ParseUint(amountStr, 10, 64)
	decimals := transfer.Info.TokenAmount.Decimals

	if authority != userAccount {
		// Output: DEX -> user
		swaps = append(swaps, SwapData{
			Type: PUMP_SWAP,
			Data: &OutputTransfer{
				TransferData: TransferData{
					Mint:     transfer.Info.Mint,
					Info:     TransferInfo{Amount: amount},
					Decimals: decimals,
				},
			},
		})
	} else {
		// Input: user -> DEX
		swaps = append(swaps, SwapData{
			Type: PUMP_SWAP,
			Data: &InputTransfer{
				TransferData: TransferData{
					Mint:     transfer.Info.Mint,
					Info:     TransferInfo{Amount: amount},
					Decimals: decimals,
				},
			},
		})
	}

	return swaps
}

func (p *Parser) processTokenTransferInstruction(userAccount string, inner solana.CompiledInstruction) []SwapData {
	var swaps []SwapData
	transfer := p.processTokenTransfer(inner)
	if transfer == nil {
		return swaps
	}
	source := transfer.Info.Source
	authority := transfer.Info.Authority
	amount := transfer.Info.Amount
	decimals := transfer.Decimals

	if authority != userAccount || source != userAccount {
		// Output: DEX -> user
		swaps = append(swaps, SwapData{
			Type: PUMP_SWAP,
			Data: &OutputTransfer{
				TransferData: TransferData{
					Mint:     transfer.Mint,
					Info:     TransferInfo{Amount: amount},
					Decimals: decimals,
				},
			},
		})
	} else {
		// Input: user -> DEX
		swaps = append(swaps, SwapData{
			Type: PUMP_SWAP,
			Data: &InputTransfer{
				TransferData: TransferData{
					Mint:     transfer.Mint,
					Info:     TransferInfo{Amount: amount},
					Decimals: decimals,
				},
			},
		})
	}

	return swaps
}

func (p *Parser) processSystemTransferInstruction(userAccount string, inner solana.CompiledInstruction) []SwapData {
	var swaps []SwapData
	transfer := p.processSystemTransfer(inner)
	if transfer == nil {
		return swaps
	}
	from := transfer.From
	amount := transfer.Amount
	decimals := uint8(9)
	mint := NATIVE_SOL_MINT_PROGRAM_ID.String()

	if from != userAccount {
		// Output: DEX -> user
		swaps = append(swaps, SwapData{
			Type: PUMP_SWAP,
			Data: &OutputTransfer{
				TransferData: TransferData{
					Mint:     mint,
					Info:     TransferInfo{Amount: amount},
					Decimals: decimals,
				},
			},
		})
	} else {
		// Input: user -> DEX
		swaps = append(swaps, SwapData{
			Type: PUMP_SWAP,
			Data: &InputTransfer{
				TransferData: TransferData{
					Mint:     mint,
					Info:     TransferInfo{Amount: amount},
					Decimals: decimals,
				},
			},
		})
	}

	return swaps
}

func (p *Parser) GetAccountMetaSlice() solana.AccountMetaSlice {

	metaSlice := make(solana.AccountMetaSlice, 0)

	for _, accIdx := range p.allAccountKeys {
		pubkey := accIdx
		metaSlice.Append(&solana.AccountMeta{
			PublicKey:  pubkey,
			IsSigner:   true,
			IsWritable: true,
		})
	}
	for _, inner := range p.txInfo.Message.GetAddressTableLookups() {
		metaSlice = append(metaSlice, &solana.AccountMeta{
			PublicKey:  inner.AccountKey,
			IsSigner:   true,
			IsWritable: true,
		})

	}
	return metaSlice
}

func (p *Parser) getPumpAmmPool() *PumpAmmPool {

	if p.txMeta == nil || p.txMeta.InnerInstructions == nil {
		return nil
	}
	for _, inner := range p.txInfo.Message.Instructions {

		if p.allAccountKeys[inner.ProgramIDIndex].Equals(PUMP_AMM_PROGRAM_ID) {
			pumpAmmPool := p.processPumpAmmAccounts(inner)
			if pumpAmmPool != nil {
				return pumpAmmPool
			}
		}

	}

	for _, inner := range p.txMeta.InnerInstructions {
		for _, inst := range inner.Instructions {
			if p.allAccountKeys[inst.ProgramIDIndex].Equals(PUMP_AMM_PROGRAM_ID) {
				pumpAmmPool := p.processPumpAmmAccounts(inst)
				if pumpAmmPool != nil {
					return pumpAmmPool
				}
			}
		}
	}
	return nil
}

func (p *Parser) processPumpAmmAccounts(inner solana.CompiledInstruction) *PumpAmmPool {

	if len(inner.Accounts) <= 19 {
		return nil
	}

	var accounts PumpAmmPool
	accounts.Pool = p.allAccountKeys[inner.Accounts[0]]
	accounts.GlobalConfig = p.allAccountKeys[inner.Accounts[2]]
	accounts.BaseMint = p.allAccountKeys[inner.Accounts[3]]
	accounts.QuoteMint = p.allAccountKeys[inner.Accounts[4]]
	accounts.PoolBaseTokenAccount = p.allAccountKeys[inner.Accounts[7]]
	accounts.PoolQuoteTokenAccount = p.allAccountKeys[inner.Accounts[8]]
	accounts.ProtocolFeeRecipient = p.allAccountKeys[inner.Accounts[9]]
	accounts.ProtocolFeeRecipientTokenAccount = p.allAccountKeys[inner.Accounts[10]]
	accounts.CoinCreatorVaultAta = p.allAccountKeys[inner.Accounts[17]]
	accounts.CoinCreatorVaultAuthority = p.allAccountKeys[inner.Accounts[18]]
	return &accounts

}

func (p *Parser) getPumpAmmEvent() *PumpAmmEvent { // anchor Self CPI Log

	if p.txMeta == nil || p.txMeta.InnerInstructions == nil {
		return nil
	}
	for _, inner := range p.txInfo.Message.Instructions {

		if p.allAccountKeys[inner.ProgramIDIndex].Equals(PUMP_AMM_PROGRAM_ID) && len(inner.Accounts) == 1 {
			pumpAmmEvent, err := parsePumpAmmEventInstruction(inner)
			if err != nil {
				continue
			}
			if pumpAmmEvent != nil {
				return pumpAmmEvent
			}
		}

	}

	for _, inner := range p.txMeta.InnerInstructions {
		for _, inst := range inner.Instructions {
			if p.allAccountKeys[inst.ProgramIDIndex].Equals(PUMP_AMM_PROGRAM_ID) && len(inst.Accounts) == 1 {
				pumpAmmEvent, err := parsePumpAmmEventInstruction(inst)
				if err != nil {
					continue
				}
				if pumpAmmEvent != nil {
					return pumpAmmEvent
				}
			}
		}
	}
	return nil
}

func parsePumpAmmEventInstruction(instruction solana.CompiledInstruction) (*PumpAmmEvent, error) {
	decodedBytes, err := base58.Decode(instruction.Data.String())
	if err != nil {
		return nil, fmt.Errorf("error decoding instruction data: %s", err)
	}
	decoder := ag_binary.NewBorshDecoder(decodedBytes[16:])

	return handlePumpAmmEvent(decoder)
}
func handlePumpAmmEvent(decoder *ag_binary.Decoder) (*PumpAmmEvent, error) {
	var create PumpAmmEvent
	if err := decoder.Decode(&create); err != nil {
		return nil, fmt.Errorf("error unmarshaling PumpAmmEvent: %s", err)
	}

	return &create, nil
}
