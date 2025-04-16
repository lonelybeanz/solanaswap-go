package solanaswapgo

import (
	"bytes"

	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
	"github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

// isTransfer checks if the instruction is a token transfer (Raydium, Orca)
func (p *Parser) isTransfer(instr *proto.InnerInstruction) bool {
	progID := p.allAccountKeys[instr.GetProgramIdIndex()]

	if !progID.Equals(solana.TokenProgramID) {
		return false
	}

	if len(instr.Accounts) < 3 || len(instr.Data) < 9 {
		return false
	}

	if instr.Data[0] != 3 {
		return false
	}

	for i := 0; i < 3; i++ {
		if int(instr.Accounts[i]) >= len(p.allAccountKeys) {
			return false
		}
	}

	return true
}

// isTransferCheck checks if the instruction is a token transfer check (Meteora)
func (p *Parser) isTransferCheck(instr *proto.InnerInstruction) bool {
	progID := p.allAccountKeys[instr.GetProgramIdIndex()]

	if !progID.Equals(solana.TokenProgramID) && !progID.Equals(solana.Token2022ProgramID) {
		return false
	}

	if len(instr.Accounts) < 4 || len(instr.Data) < 9 {
		return false
	}

	if instr.Data[0] != 12 {
		return false
	}

	for i := 0; i < 4; i++ {
		if int(instr.Accounts[i]) >= len(p.allAccountKeys) {
			return false
		}
	}

	return true
}

func (p *Parser) isPumpFunTradeEventInstruction(inst *proto.InnerInstruction) bool {
	if !p.allAccountKeys[inst.GetProgramIdIndex()].Equals(PUMP_FUN_PROGRAM_ID) || len(inst.Data) < 16 {
		return false
	}
	decodedBytes, err := base58.Decode(string(inst.GetData()))
	if err != nil {
		return false
	}
	return bytes.Equal(decodedBytes[:16], PumpfunTradeEventDiscriminator[:])
}

func (p *Parser) isJupiterRouteEventInstruction(inst *proto.InnerInstruction) bool {
	if !p.allAccountKeys[inst.GetProgramIdIndex()].Equals(JUPITER_PROGRAM_ID) || len(inst.Data) < 16 {
		return false
	}
	decodedBytes, err := base58.Decode(string(inst.GetData()))
	if err != nil {
		return false
	}
	return bytes.Equal(decodedBytes[:16], JupiterRouteEventDiscriminator[:])
}
