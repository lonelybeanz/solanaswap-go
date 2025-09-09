package solanaswapgo

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

var (
	PumpfunTradeEventDiscriminator  = [16]byte{228, 69, 165, 46, 81, 203, 154, 29, 189, 219, 127, 211, 78, 230, 97, 238}
	PumpfunCreateEventDiscriminator = [16]byte{228, 69, 165, 46, 81, 203, 154, 29, 27, 114, 169, 77, 222, 235, 99, 118}
)

type PumpFunPool struct {
	Global                 solana.PublicKey
	FeeRecipient           solana.PublicKey
	Mint                   solana.PublicKey
	BondingCurve           solana.PublicKey
	AssociatedBondingCurve solana.PublicKey
	CreatorVault           solana.PublicKey
	EventAuthority         solana.PublicKey
	VirtualSolReserves     uint64
	VirtualTokenReserves   uint64
	RealSOLReserves        uint64
	RealTokenReserves      uint64
}

type PumpfunTradeEvent struct {
	Mint                 solana.PublicKey
	SolAmount            uint64
	TokenAmount          uint64
	IsBuy                bool
	User                 solana.PublicKey
	Timestamp            int64
	VirtualSolReserves   uint64
	VirtualTokenReserves uint64
	RealSOLReserves      uint64
	RealTokenReserves    uint64
}

type PumpfunCreateEvent struct {
	Name         string
	Symbol       string
	Uri          string
	Mint         solana.PublicKey
	BondingCurve solana.PublicKey
	User         solana.PublicKey
}

func (p *Parser) processPumpfunSwaps(instructionIndex int) []SwapData {
	var swaps []SwapData
	for _, innerInstructionSet := range p.txMeta.InnerInstructions {
		if innerInstructionSet.Index == uint16(instructionIndex) {
			for _, innerInstruction := range innerInstructionSet.Instructions {
				if p.isPumpFunTradeEventInstruction(innerInstruction) {
					eventData, err := p.parsePumpfunTradeEventInstruction(innerInstruction)
					if err != nil {
						p.Log.Errorf("error processing Pumpfun trade event: %s", err)
					}
					if eventData != nil {
						swaps = append(swaps, SwapData{Type: PUMP_FUN, Data: eventData})
					}
				}
			}
		}
	}
	return swaps
}

func (p *Parser) getPumpFunPool() *PumpFunPool {
	if p.txMeta == nil || p.txMeta.InnerInstructions == nil {
		return nil
	}
	for _, inner := range p.txInfo.Message.Instructions {

		if p.allAccountKeys[inner.ProgramIDIndex].Equals(PUMP_FUN_PROGRAM_ID) {
			pumpFunPool := p.processPumpFunAccounts(inner)
			if pumpFunPool != nil {
				return pumpFunPool
			}
		}

	}

	for _, inner := range p.txMeta.InnerInstructions {
		for _, inst := range inner.Instructions {
			if p.allAccountKeys[inst.ProgramIDIndex].Equals(PUMP_FUN_PROGRAM_ID) {
				pumpFunPool := p.processPumpFunAccounts(inst)
				if pumpFunPool != nil {
					return pumpFunPool
				}
			}
		}
	}
	return nil
}

func (p *Parser) processPumpFunAccounts(inner solana.CompiledInstruction) *PumpFunPool {
	if len(inner.Accounts) >= 14 {
		if !p.allAccountKeys[inner.Accounts[0]].Equals(solana.MustPublicKeyFromBase58("4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf")) {
			return nil
		}
		var accounts PumpFunPool
		accounts.Global = p.allAccountKeys[inner.Accounts[0]]
		accounts.FeeRecipient = p.allAccountKeys[inner.Accounts[1]]
		accounts.Mint = p.allAccountKeys[inner.Accounts[2]]
		accounts.BondingCurve = p.allAccountKeys[inner.Accounts[3]]
		accounts.AssociatedBondingCurve = p.allAccountKeys[inner.Accounts[4]]
		accounts.CreatorVault = p.allAccountKeys[inner.Accounts[9]]
		accounts.EventAuthority = p.allAccountKeys[inner.Accounts[10]]
		return &accounts
	}
	return nil
}

func (p *Parser) getPumpFunEvent() *PumpfunTradeEvent { // anchor Self CPI Log
	var events []*PumpfunTradeEvent

	if p.txMeta == nil || p.txMeta.InnerInstructions == nil {
		return nil
	}
	for _, inner := range p.txInfo.Message.Instructions {

		if p.allAccountKeys[inner.ProgramIDIndex].Equals(PUMP_FUN_PROGRAM_ID) && len(inner.Accounts) == 1 {
			pumpAmmEvent, err := p.parsePumpfunTradeEventInstruction(inner)
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
			if p.allAccountKeys[inst.ProgramIDIndex].Equals(PUMP_FUN_PROGRAM_ID) && len(inst.Accounts) == 1 {
				pumpAmmEvent, err := p.parsePumpfunTradeEventInstruction(inst)
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

func (p *Parser) parsePumpfunTradeEventInstruction(instruction solana.CompiledInstruction) (*PumpfunTradeEvent, error) {
	decodedBytes, err := base58.Decode(instruction.Data.String())
	if err != nil {
		return nil, fmt.Errorf("error decoding instruction data: %s", err)
	}
	decoder := ag_binary.NewBorshDecoder(decodedBytes[16:])

	return handlePumpfunTradeEvent(decoder)
}

func handlePumpfunTradeEvent(decoder *ag_binary.Decoder) (*PumpfunTradeEvent, error) {
	var trade PumpfunTradeEvent
	if err := decoder.Decode(&trade); err != nil {
		return nil, fmt.Errorf("error unmarshaling TradeEvent: %s", err)
	}

	return &trade, nil
}

func (p *Parser) parsePumpfunCreateEventInstruction(instruction solana.CompiledInstruction) (*PumpfunCreateEvent, error) {
	decodedBytes, err := base58.Decode(instruction.Data.String())
	if err != nil {
		return nil, fmt.Errorf("error decoding instruction data: %s", err)
	}
	decoder := ag_binary.NewBorshDecoder(decodedBytes[16:])

	return handlePumpfunCreateEvent(decoder)
}
func handlePumpfunCreateEvent(decoder *ag_binary.Decoder) (*PumpfunCreateEvent, error) {
	var create PumpfunCreateEvent
	if err := decoder.Decode(&create); err != nil {
		return nil, fmt.Errorf("error unmarshaling CreateEvent: %s", err)
	}

	return &create, nil
}
