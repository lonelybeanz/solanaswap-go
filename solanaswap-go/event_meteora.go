package solanaswapgo

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

type MeteoraDbcPool struct {
	PoolAuthority        solana.PublicKey
	Config               solana.PublicKey
	Pool                 solana.PublicKey
	BaseVault            solana.PublicKey
	QuoteVault           solana.PublicKey
	BaseMint             solana.PublicKey
	QuoteMint            solana.PublicKey
	TokenBaseProgram     solana.PublicKey
	TokenQuoteProgram    solana.PublicKey
	ReferralTokenAccount solana.PublicKey
	EventAuthority       solana.PublicKey
	NextSqrtPrice        uint64
}

// 先定义结构体（和 Anchor 中顺序、类型一致）
type MeteoraDbcEvent struct {
	Pool             solana.PublicKey
	Config           solana.PublicKey
	TradeDirection   uint8
	HasReferral      bool
	Params           SwapParams
	SwapResult       SwapResult
	AmountIn         uint64
	CurrentTimestamp uint64
}

type SwapParams struct {
	AmountIn         uint64
	MinimumAmountOut uint64
}

type SwapResult struct {
	ActualInputAmount uint64
	OutputAmount      uint64
	NextSqrtPrice     uint64 // 可用 big.Int 或自定义类型
	TradingFee        uint64
	ProtocolFee       uint64
	ReferralFee       uint64
}

func (p *Parser) getMeteoraDbcPool() *MeteoraDbcPool {

	if p.txMeta == nil || p.txMeta.InnerInstructions == nil {
		return nil
	}
	for _, inner := range p.txInfo.Message.Instructions {

		if p.allAccountKeys[inner.ProgramIDIndex].Equals(METEORA_DBC_PROGRAM_ID) && len(inner.Accounts) == 15 {
			pumpAmmPool := p.processMeteoraDbcAccounts(inner)
			if pumpAmmPool != nil {
				return pumpAmmPool
			}
		}

	}

	for _, inner := range p.txMeta.InnerInstructions {
		for _, inst := range inner.Instructions {
			if p.allAccountKeys[inst.ProgramIDIndex].Equals(METEORA_DBC_PROGRAM_ID) && len(inst.Accounts) == 15 {
				pumpAmmPool := p.processMeteoraDbcAccounts(inst)
				if pumpAmmPool != nil {
					return pumpAmmPool
				}
			}
		}
	}
	return nil
}

func (p *Parser) processMeteoraDbcAccounts(inner solana.CompiledInstruction) *MeteoraDbcPool {
	var accounts MeteoraDbcPool
	accounts.PoolAuthority = p.allAccountKeys[inner.Accounts[0]]
	accounts.Config = p.allAccountKeys[inner.Accounts[1]]
	accounts.Pool = p.allAccountKeys[inner.Accounts[2]]
	accounts.BaseVault = p.allAccountKeys[inner.Accounts[5]]
	accounts.QuoteVault = p.allAccountKeys[inner.Accounts[6]]
	accounts.BaseMint = p.allAccountKeys[inner.Accounts[7]]
	accounts.QuoteMint = p.allAccountKeys[inner.Accounts[8]]
	accounts.TokenBaseProgram = p.allAccountKeys[inner.Accounts[10]]
	accounts.TokenQuoteProgram = p.allAccountKeys[inner.Accounts[11]]
	accounts.ReferralTokenAccount = p.allAccountKeys[inner.Accounts[12]]
	accounts.EventAuthority = p.allAccountKeys[inner.Accounts[13]]
	return &accounts

}

func (p *Parser) getMeteoraDbcEvent() *MeteoraDbcEvent { // anchor Self CPI Log

	if p.txMeta == nil || p.txMeta.InnerInstructions == nil {
		return nil
	}
	for _, inner := range p.txInfo.Message.Instructions {

		if p.allAccountKeys[inner.ProgramIDIndex].Equals(METEORA_DBC_PROGRAM_ID) && len(inner.Accounts) == 1 {
			pumpAmmEvent, err := parseMeteoraDbcEventInstruction(inner)
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
			if p.allAccountKeys[inst.ProgramIDIndex].Equals(METEORA_DBC_PROGRAM_ID) && len(inst.Accounts) == 1 {
				pumpAmmEvent, err := parseMeteoraDbcEventInstruction(inst)
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

func parseMeteoraDbcEventInstruction(instruction solana.CompiledInstruction) (*MeteoraDbcEvent, error) {
	decodedBytes, err := base58.Decode(instruction.Data.String())
	if err != nil {
		return nil, fmt.Errorf("error decoding instruction data: %s", err)
	}
	decoder := ag_binary.NewBorshDecoder(decodedBytes[16:])

	return handleMeteoraDbcEvent(decoder)
}
func handleMeteoraDbcEvent(decoder *ag_binary.Decoder) (*MeteoraDbcEvent, error) {
	var event MeteoraDbcEvent
	if err := decoder.Decode(&event); err != nil {
		return nil, fmt.Errorf("error unmarshaling MeteoraDbcEvent: %s", err)
	}

	return &event, nil
}
