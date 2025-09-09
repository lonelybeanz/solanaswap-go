package solanaswapgo

import (
	"bytes"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
)

func (p *Parser) isSystemTransfer(instr solana.CompiledInstruction) bool {
	progID := p.allAccountKeys[instr.ProgramIDIndex]

	if !progID.Equals(solana.SystemProgramID) {
		return false
	}

	if len(instr.Accounts) < 2 {
		return false
	}

	// Transfer 指令 opcode 是 2，data 至少 9 字节（1 + 8）
	if len(instr.Data) < 9 || instr.Data[0] != 2 {
		return false
	}

	for _, idx := range instr.Accounts {
		if int(idx) >= len(p.allAccountKeys) {
			return false
		}
	}

	return true
}

// isTransfer checks if the instruction is a token transfer (Raydium, Orca)
func (p *Parser) isTokenTransfer(instr solana.CompiledInstruction) bool {
	progID := p.allAccountKeys[instr.ProgramIDIndex]

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
func (p *Parser) isTransferCheck(instr solana.CompiledInstruction) bool {
	progID := p.allAccountKeys[instr.ProgramIDIndex]

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

func (p *Parser) isPumpFunInstruction(inst solana.CompiledInstruction) bool {
	if !p.allAccountKeys[inst.ProgramIDIndex].Equals(PUMP_FUN_PROGRAM_ID) || len(inst.Data) < 16 {
		return false
	}
	decodedBytes, err := base58.Decode(inst.Data.String())
	if err != nil {
		return false
	}
	return bytes.Equal(decodedBytes[:16], PumpfunTradeEventDiscriminator[:])
}

func (p *Parser) isPumpAmmInstruction(inst solana.CompiledInstruction) bool {
	return p.allAccountKeys[inst.ProgramIDIndex].Equals(PUMP_AMM_PROGRAM_ID) && len(inst.Accounts) == 21
}

func (p *Parser) isPumpfunCreateEventInstruction(inst rpc.CompiledInstruction) bool {
	if !p.allAccountKeys[inst.ProgramIDIndex].Equals(PUMP_FUN_PROGRAM_ID) || len(inst.Data) < 16 {
		return false
	}
	decodedBytes, err := base58.Decode(inst.Data.String())
	if err != nil {
		return false
	}
	return bytes.Equal(decodedBytes[:16], PumpfunCreateEventDiscriminator[:])
}

func (p *Parser) isPumpFunTradeEventInstruction(inst rpc.CompiledInstruction) bool {
	if !p.allAccountKeys[inst.ProgramIDIndex].Equals(PUMP_FUN_PROGRAM_ID) || len(inst.Data) < 16 {
		return false
	}
	decodedBytes, err := base58.Decode(inst.Data.String())
	if err != nil {
		return false
	}
	return bytes.Equal(decodedBytes[:16], PumpfunTradeEventDiscriminator[:])
}

func (p *Parser) isJupiterRouteEventInstruction(inst rpc.CompiledInstruction) bool {
	if !p.allAccountKeys[inst.ProgramIDIndex].Equals(JUPITER_PROGRAM_ID) || len(inst.Data) < 16 {
		return false
	}
	decodedBytes, err := base58.Decode(inst.Data.String())
	if err != nil {
		return false
	}
	return bytes.Equal(decodedBytes[:16], JupiterRouteEventDiscriminator[:])
}

func (p *Parser) ContainsProgram(programID solana.PublicKey) bool {
	for _, account := range p.allAccountKeys {
		if account.Equals(programID) {
			return true
		}
	}
	return false
}
