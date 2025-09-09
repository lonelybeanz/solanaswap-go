package solanaswapgo

import (
	"encoding/binary"

	"github.com/gagliardetto/solana-go"
)

type TransferInfo struct {
	Amount      uint64 `json:"amount"`
	Authority   string `json:"authority"`
	Destination string `json:"destination"`
	Source      string `json:"source"`
}

type TransferData struct {
	Info     TransferInfo `json:"info"`
	Type     string       `json:"type"`
	Mint     string       `json:"mint"`
	Decimals uint8        `json:"decimals"`
}

type SystemTransfer struct {
	From   string
	To     string
	Amount uint64 // 单位：lamports（1 SOL = 1e9 lamports）
}

type TokenInfo struct {
	Mint     string
	Decimals uint8
}

func (p *Parser) processRaydSwaps(instructionIndex int) []SwapData {
	var swaps []SwapData
	for _, innerInstructionSet := range p.txMeta.InnerInstructions {
		if innerInstructionSet.Index == uint16(instructionIndex) {
			for _, innerInstruction := range innerInstructionSet.Instructions {
				switch {
				case p.isTokenTransfer(p.convertRPCToSolanaInstruction(innerInstruction)):
					transfer := p.processTokenTransfer(p.convertRPCToSolanaInstruction(innerInstruction))
					if transfer != nil {
						swaps = append(swaps, SwapData{Type: RAYDIUM, Data: transfer})
					}
				case p.isTransferCheck(p.convertRPCToSolanaInstruction(innerInstruction)):
					transfer := p.processTransferCheck(p.convertRPCToSolanaInstruction(innerInstruction))
					if transfer != nil {
						swaps = append(swaps, SwapData{Type: RAYDIUM, Data: transfer})
					}
				}
			}
		}
	}
	return swaps
}

func (p *Parser) processOrcaSwaps(instructionIndex int) []SwapData {
	var swaps []SwapData
	for _, innerInstructionSet := range p.txMeta.InnerInstructions {
		if innerInstructionSet.Index == uint16(instructionIndex) {
			for _, innerInstruction := range innerInstructionSet.Instructions {
				if p.isTokenTransfer(p.convertRPCToSolanaInstruction(innerInstruction)) {
					transfer := p.processTokenTransfer(p.convertRPCToSolanaInstruction(innerInstruction))
					if transfer != nil {
						swaps = append(swaps, SwapData{Type: ORCA, Data: transfer})
					}
				}
			}
		}
	}
	return swaps
}

func (p *Parser) processTokenTransfer(instr solana.CompiledInstruction) *TransferData {
	amount := binary.LittleEndian.Uint64(instr.Data[1:9])

	transferData := &TransferData{
		Info: TransferInfo{
			Amount:      amount,
			Source:      p.allAccountKeys[instr.Accounts[0]].String(),
			Destination: p.allAccountKeys[instr.Accounts[1]].String(),
			Authority:   p.allAccountKeys[instr.Accounts[2]].String(),
		},
		Type:     "transfer",
		Mint:     p.splTokenInfoMap[p.allAccountKeys[instr.Accounts[1]].String()].Mint,
		Decimals: p.splTokenInfoMap[p.allAccountKeys[instr.Accounts[1]].String()].Decimals,
	}

	if transferData.Mint == "" {
		transferData.Mint = "Unknown"
	}

	return transferData
}

func (p *Parser) processSystemTransfer(instr solana.CompiledInstruction) *SystemTransfer {
	progID := p.allAccountKeys[instr.ProgramIDIndex]

	// 判断是否是系统程序
	if !progID.Equals(solana.SystemProgramID) {
		return nil
	}

	// Data 至少为 9 字节，且首字节为 0（transfer 指令）
	if condition := len(instr.Data) < 9 || instr.Data[0] != 2; condition {
		return nil
	}

	// 判断账户数量是否足够（至少两个）
	if len(instr.Accounts) < 2 {
		return nil
	}

	// 获取 source 和 destination 的 pubkey
	fromIndex := instr.Accounts[0]
	toIndex := instr.Accounts[1]

	if int(fromIndex) >= len(p.allAccountKeys) || int(toIndex) >= len(p.allAccountKeys) {
		return nil
	}

	from := p.allAccountKeys[fromIndex]
	to := p.allAccountKeys[toIndex]

	// 提取 lamports（从第1字节开始的8个字节，little-endian）
	amount := binary.LittleEndian.Uint64(instr.Data[4:12])

	return &SystemTransfer{
		From:   from.String(),
		To:     to.String(),
		Amount: amount,
	}
}

func (p *Parser) extractSPLTokenInfo() error {
	splTokenAddresses := make(map[string]TokenInfo)

	for _, accountInfo := range p.txMeta.PostTokenBalances {
		if !accountInfo.Mint.IsZero() {
			accountKey := p.allAccountKeys[accountInfo.AccountIndex].String()
			splTokenAddresses[accountKey] = TokenInfo{
				Mint:     accountInfo.Mint.String(),
				Decimals: accountInfo.UiTokenAmount.Decimals,
			}
		}
	}

	processInstruction := func(instr solana.CompiledInstruction) {
		if !p.allAccountKeys[instr.ProgramIDIndex].Equals(solana.TokenProgramID) {
			return
		}

		if len(instr.Data) == 0 || (instr.Data[0] != 3 && instr.Data[0] != 12) {
			return
		}

		if len(instr.Accounts) < 3 {
			return
		}

		source := p.allAccountKeys[instr.Accounts[0]].String()
		destination := p.allAccountKeys[instr.Accounts[1]].String()

		if _, exists := splTokenAddresses[source]; !exists {
			splTokenAddresses[source] = TokenInfo{Mint: "", Decimals: 0}
		}
		if _, exists := splTokenAddresses[destination]; !exists {
			splTokenAddresses[destination] = TokenInfo{Mint: "", Decimals: 0}
		}
	}

	for _, instr := range p.txInfo.Message.Instructions {
		processInstruction(instr)
	}
	for _, innerSet := range p.txMeta.InnerInstructions {
		for _, instr := range innerSet.Instructions {
			processInstruction(p.convertRPCToSolanaInstruction(instr))
		}
	}

	for account, info := range splTokenAddresses {
		if info.Mint == "" {
			splTokenAddresses[account] = TokenInfo{
				Mint:     NATIVE_SOL_MINT_PROGRAM_ID.String(),
				Decimals: 9, // Native SOL has 9 decimal places
			}
		}
	}

	p.splTokenInfoMap = splTokenAddresses

	return nil
}
