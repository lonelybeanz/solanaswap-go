package solanaswapgo

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

type RaydiumLaunchpadPool struct {
	Authority      solana.PublicKey
	GlobalConfig   solana.PublicKey
	PlatformConfig solana.PublicKey
	PoolState      solana.PublicKey
	BaseVault      solana.PublicKey
	QuoteVault     solana.PublicKey
	BaseMint       solana.PublicKey
	QuoteMint      solana.PublicKey
	EventAuthority solana.PublicKey

	VirtualBase     uint64
	VirtualQuote    uint64
	RealBaseBefore  uint64
	RealQuoteBefore uint64
}

// 先定义结构体（和 Anchor 中顺序、类型一致）
type RaydiumLaunchpadEvent struct {
	PoolState       solana.PublicKey // 32字节公钥
	TotalBaseSell   uint64
	VirtualBase     uint64
	VirtualQuote    uint64
	RealBaseBefore  uint64
	RealQuoteBefore uint64
	RealBaseAfter   uint64
	RealQuoteAfter  uint64
	AmountIn        uint64
	AmountOut       uint64
	ProtocolFee     uint64
	PlatformFee     uint64
	ShareFee        uint64
}

type RaydiumCPMMPool struct {
	Authority              solana.PublicKey
	AmmConfig              solana.PublicKey
	PoolState              solana.PublicKey
	InputVault             solana.PublicKey
	OutputVault            solana.PublicKey
	InputTokenMint         solana.PublicKey
	OutputTokenMint        solana.PublicKey
	ObservationState       solana.PublicKey
	PoolBaseTokenReserves  uint64
	PoolQuoteTokenReserves uint64
}

func (p *Parser) getRaydiumLaunchpadPool() *RaydiumLaunchpadPool {

	if p.txMeta == nil || p.txMeta.InnerInstructions == nil {
		return nil
	}
	for _, inner := range p.txInfo.Message.Instructions {

		if p.allAccountKeys[inner.ProgramIDIndex].Equals(RAYDIUM_Launchpad_PROGRAM_ID) && len(inner.Accounts) >= 18 {
			pumpAmmPool := p.processRaydiumLaunchpadAccounts(inner)
			if pumpAmmPool != nil {
				return pumpAmmPool
			}
		}

	}

	for _, inner := range p.txMeta.InnerInstructions {
		for _, inst := range inner.Instructions {
			if p.allAccountKeys[inst.ProgramIDIndex].Equals(RAYDIUM_Launchpad_PROGRAM_ID) && len(inst.Accounts) >= 18 {
				pumpAmmPool := p.processRaydiumLaunchpadAccounts(p.convertRPCToSolanaInstruction(inst))
				if pumpAmmPool != nil {
					return pumpAmmPool
				}
			}
		}
	}
	return nil
}

func (p *Parser) processRaydiumLaunchpadAccounts(inner solana.CompiledInstruction) *RaydiumLaunchpadPool {
	var accounts RaydiumLaunchpadPool
	accounts.Authority = p.allAccountKeys[inner.Accounts[1]]
	accounts.GlobalConfig = p.allAccountKeys[inner.Accounts[2]]
	accounts.PlatformConfig = p.allAccountKeys[inner.Accounts[3]]
	accounts.PoolState = p.allAccountKeys[inner.Accounts[4]]
	accounts.BaseVault = p.allAccountKeys[inner.Accounts[7]]
	accounts.QuoteVault = p.allAccountKeys[inner.Accounts[8]]
	accounts.BaseMint = p.allAccountKeys[inner.Accounts[9]]
	accounts.QuoteMint = p.allAccountKeys[inner.Accounts[10]]
	accounts.EventAuthority = p.allAccountKeys[inner.Accounts[13]]

	return &accounts

}

func (p *Parser) getRaydiumLaunchpadEvent() *RaydiumLaunchpadEvent { // anchor Self CPI Log
	var events []*RaydiumLaunchpadEvent
	if p.txMeta == nil || p.txMeta.InnerInstructions == nil {
		return nil
	}
	for _, inner := range p.txInfo.Message.Instructions {

		if p.allAccountKeys[inner.ProgramIDIndex].Equals(RAYDIUM_Launchpad_PROGRAM_ID) && len(inner.Accounts) == 1 {
			pumpAmmEvent, err := parseRaydiumLaunchpadEventInstruction(inner)
			if err != nil {
				continue
			}
			if pumpAmmEvent != nil {
				events = append(events, pumpAmmEvent)
			}
		}

	}

	for _, inner := range p.txMeta.InnerInstructions {
		for _, inst := range inner.Instructions {
			if p.allAccountKeys[inst.ProgramIDIndex].Equals(RAYDIUM_Launchpad_PROGRAM_ID) && len(inst.Accounts) == 1 {
				pumpAmmEvent, err := parseRaydiumLaunchpadEventInstruction(p.convertRPCToSolanaInstruction(inst))
				if err != nil {
					continue
				}
				if pumpAmmEvent != nil {
					events = append(events, pumpAmmEvent)
				}
			}
		}
	}
	if len(events) < 1 {
		return nil
	}
	if len(events) > 0 {
		return events[len(events)-1]
	}

	return events[0]
}

func parseRaydiumLaunchpadEventInstruction(instruction solana.CompiledInstruction) (*RaydiumLaunchpadEvent, error) {
	decodedBytes, err := base58.Decode(instruction.Data.String())
	if err != nil {
		return nil, fmt.Errorf("error decoding instruction data: %s", err)
	}
	decoder := ag_binary.NewBorshDecoder(decodedBytes[16:])

	return handleRaydiumLaunchpadEvent(decoder)
}
func handleRaydiumLaunchpadEvent(decoder *ag_binary.Decoder) (*RaydiumLaunchpadEvent, error) {
	var event RaydiumLaunchpadEvent
	if err := decoder.Decode(&event); err != nil {
		return nil, fmt.Errorf("error unmarshaling RaydiumLaunchpadEvent: %s", err)
	}

	return &event, nil
}
