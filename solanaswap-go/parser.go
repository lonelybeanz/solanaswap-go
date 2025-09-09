package solanaswapgo

import (
	"fmt"
	"strconv"
	"time"

	pb "solana-bot/internal/pb/yellowstone-grpc"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

const (
	PROTOCOL_RAYDIUM  = "raydium"
	PROTOCOL_ORCA     = "orca"
	PROTOCOL_METEORA  = "meteora"
	PROTOCOL_PUMPFUN  = "pumpfun"
	PROTOCOL_PUMPSWAP = "pumpswap"
)

type TokenTransfer struct {
	user     string
	mint     string
	amount   uint64
	decimals uint8
}

type Parser struct {
	txMeta *rpc.TransactionMeta
	txInfo *solana.Transaction
	// allAccountMetas solana.AccountMetaSlice
	allAccountKeys  solana.PublicKeySlice
	splTokenInfoMap map[string]TokenInfo
	splDecimalsMap  map[string]uint8
	SwapType        SwapType
	Log             *logrus.Logger
}

func NewTransactionParser(tx *rpc.GetTransactionResult) (*Parser, error) {
	txInfo, err := tx.Transaction.GetTransaction()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return NewTransactionParserFromTransaction(txInfo, tx.Meta)
}

func NewPbTransactionParserFromTransaction(pbtx *pb.Transaction, pbtxMeta *pb.TransactionStatusMeta) (*Parser, error) {

	convertToUint16 := func(data []byte) []uint16 {
		result := make([]uint16, len(data))
		for i, b := range data {
			result[i] = uint16(b)
		}
		return result
	}

	tx := &solana.Transaction{
		Message: solana.Message{
			Instructions: make([]solana.CompiledInstruction, len(pbtx.Message.Instructions)),
			AccountKeys:  make(solana.PublicKeySlice, len(pbtx.Message.AccountKeys)),
		},
		Signatures: make([]solana.Signature, len(pbtx.Signatures)),
	}
	for i, sig := range pbtx.Signatures {
		tx.Signatures[i] = solana.SignatureFromBytes(sig)
	}
	for i, key := range pbtx.Message.AccountKeys {
		tx.Message.AccountKeys[i] = solana.PublicKeyFromBytes(key)
	}
	for i, inst := range pbtx.Message.Instructions {
		tx.Message.Instructions[i] = solana.CompiledInstruction{
			ProgramIDIndex: uint16(inst.GetProgramIdIndex()),
			Accounts:       convertToUint16(inst.Accounts),
			Data:           inst.Data,
		}
	}

	// 将 [][]byte 转换为 solana.PublicKeySlice
	var readOnlypublicKeySlice = make(solana.PublicKeySlice, len(pbtxMeta.LoadedReadonlyAddresses))
	for i, key := range pbtxMeta.LoadedReadonlyAddresses {
		publicKey := solana.PublicKeyFromBytes(key)
		readOnlypublicKeySlice[i] = publicKey
		// readOnlypublicKeySlice = append(readOnlypublicKeySlice, publicKey)
	}

	var writablepublicKeySlice = make(solana.PublicKeySlice, len(pbtxMeta.LoadedWritableAddresses))

	for i, key := range pbtxMeta.LoadedWritableAddresses {
		publicKey := solana.PublicKeyFromBytes(key)
		writablepublicKeySlice[i] = publicKey
		// writablepublicKeySlice = append(writablepublicKeySlice, publicKey)
	}

	txMeta := &rpc.TransactionMeta{
		PostBalances:      pbtxMeta.PostBalances,
		PreBalances:       pbtxMeta.PreBalances,
		PostTokenBalances: make([]rpc.TokenBalance, len(pbtxMeta.PostTokenBalances)),
		PreTokenBalances:  make([]rpc.TokenBalance, len(pbtxMeta.PreTokenBalances)),
		LoadedAddresses: rpc.LoadedAddresses{
			ReadOnly: readOnlypublicKeySlice,
			Writable: writablepublicKeySlice,
		},
		InnerInstructions: make([]rpc.InnerInstruction, len(pbtxMeta.InnerInstructions)),
		LogMessages:       pbtxMeta.LogMessages,
	}
	for i, inner := range pbtxMeta.InnerInstructions {
		txMeta.InnerInstructions[i] = rpc.InnerInstruction{
			Index:        uint16(inner.Index),
			Instructions: make([]solana.CompiledInstruction, len(inner.Instructions)),
		}
		for j, instr := range inner.Instructions {
			txMeta.InnerInstructions[i].Instructions[j] = solana.CompiledInstruction{
				ProgramIDIndex: uint16(instr.GetProgramIdIndex()),
				Accounts:       convertToUint16(instr.Accounts),
				Data:           instr.Data,
			}
		}
	}
	for i, tokenBalance := range pbtxMeta.PostTokenBalances {
		txMeta.PostTokenBalances[i] = rpc.TokenBalance{
			AccountIndex: uint16(tokenBalance.GetAccountIndex()),
			Mint:         solana.MustPublicKeyFromBase58(tokenBalance.GetMint()),
			UiTokenAmount: &rpc.UiTokenAmount{
				Amount:   tokenBalance.GetUiTokenAmount().GetAmount(),
				Decimals: uint8(tokenBalance.GetUiTokenAmount().GetDecimals()),
			},
		}
	}

	for i, tokenBalance := range pbtxMeta.PreTokenBalances {
		txMeta.PreTokenBalances[i] = rpc.TokenBalance{
			AccountIndex: uint16(tokenBalance.GetAccountIndex()),
			Mint:         solana.MustPublicKeyFromBase58(tokenBalance.GetMint()),
			UiTokenAmount: &rpc.UiTokenAmount{
				Amount:   tokenBalance.GetUiTokenAmount().GetAmount(),
				Decimals: uint8(tokenBalance.GetUiTokenAmount().GetDecimals()),
			},
		}
	}

	return NewTransactionParserFromTransaction(tx, txMeta)

}

func BuildAddressTablesFromMeta(msg *solana.Message, meta *rpc.TransactionMeta) map[solana.PublicKey]solana.PublicKeySlice {
	result := make(map[solana.PublicKey]solana.PublicKeySlice)
	writableIndex := 0
	readonlyIndex := 0

	for _, lookup := range msg.AddressTableLookups {
		maxIndex := uint8(0)
		for _, i := range lookup.WritableIndexes {
			if i > maxIndex {
				maxIndex = i
			}
		}
		for _, i := range lookup.ReadonlyIndexes {
			if i > maxIndex {
				maxIndex = i
			}
		}

		table := make(solana.PublicKeySlice, maxIndex+1)
		for i := uint8(0); i <= maxIndex; i++ {
			if contains(lookup.WritableIndexes, i) {
				table[i] = meta.LoadedAddresses.Writable[writableIndex]
				writableIndex++
			} else if contains(lookup.ReadonlyIndexes, i) {
				table[i] = meta.LoadedAddresses.ReadOnly[readonlyIndex]
				readonlyIndex++
			}
		}
		result[lookup.AccountKey] = table
	}

	return result
}

func contains(arr []uint8, x uint8) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}

func NewTransactionParserFromTransaction(tx *solana.Transaction, txMeta *rpc.TransactionMeta) (*Parser, error) {
	allAccountKeys := append(tx.Message.AccountKeys, txMeta.LoadedAddresses.Writable...)
	allAccountKeys = append(allAccountKeys, txMeta.LoadedAddresses.ReadOnly...)

	// tables := BuildAddressTablesFromMeta(&tx.Message, txMeta)

	// spew.Dump(allAccountKeys)

	// // Step 2: 设置 Address Tables
	// msg := tx.Message
	// err := msg.SetAddressTables(tables)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to set address tables: %w", err)
	// }

	// // Step 3: Resolve ALT
	// if err := msg.ResolveLookups(); err != nil {
	// 	return nil, fmt.Errorf("resolve lookup failed: %w", err)
	// }

	// // Step 4: 获取 AccountMetaList（用于后续解析）
	// metas, err := msg.AccountMetaList()
	// if err != nil {
	// 	return nil, fmt.Errorf("account meta list failed: %w", err)
	// }

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	// spew.Dump(allAccountKeys)

	parser := &Parser{
		txMeta:         txMeta,
		txInfo:         tx,
		allAccountKeys: allAccountKeys,
		// allAccountMetas: metas,
		Log: log,
	}

	for _, v := range allAccountKeys {
		if v.Equals(PUMP_FUN_PROGRAM_ID) {
			parser.SwapType = PUMP_FUN
		} else if v.Equals(PUMP_AMM_PROGRAM_ID) {
			parser.SwapType = PUMP_SWAP
		} else if v.Equals(METEORA_DBC_PROGRAM_ID) {
			parser.SwapType = METEORA_DBC
		} else if v.Equals(RAYDIUM_Launchpad_PROGRAM_ID) {
			parser.SwapType = RAYDIUM_Launchpad
		}
	}

	if err := parser.extractSPLTokenInfo(); err != nil {
		return nil, fmt.Errorf("failed to extract SPL Token Addresses: %w", err)
	}

	if err := parser.extractSPLDecimals(); err != nil {
		return nil, fmt.Errorf("failed to extract SPL decimals: %w", err)
	}

	return parser, nil
}

func (p *Parser) ParseTransactionForMint() (*PumpfunCreateEvent, error) {

	for i, outerInstruction := range p.txInfo.Message.Instructions {
		progID := p.allAccountKeys[outerInstruction.ProgramIDIndex]
		if progID.Equals(PUMP_FUN_PROGRAM_ID) && len(outerInstruction.Accounts) == 14 {

			innerInstructions := p.getInnerInstructions(i)
			for _, inner := range innerInstructions {
				if p.isPumpfunCreateEventInstruction(inner) {
					createEvent, err := p.parsePumpfunCreateEventInstruction(inner)
					if err != nil {
						return nil, fmt.Errorf("error processing Pumpfun create event: %s", err)
					}
					if createEvent != nil {
						return createEvent, nil
					}
				}
			}

		}
	}
	return nil, fmt.Errorf("no valid mint data found")
}

type MigrateInfo struct {
	BaseMint  string
	QuoteMint string
	PoolData  interface{}
}

func (p *Parser) ParseTransactionForMigrate() (*MigrateInfo, error) {
	if p.allAccountKeys.Contains(RAYDIUM_Launchpad_Migration_PROGRAM_ID) {
		p.SwapType = "RAYDIUM_CPMM"
	}

	return &MigrateInfo{
		BaseMint:  p.allAccountKeys[0].String(),
		QuoteMint: p.allAccountKeys[0].String(),
		PoolData:  p.allAccountKeys[0].String(),
	}, nil
}

type SwapData struct {
	Type SwapType
	Data interface{}
}

func (p *Parser) ParseTransactionForSwap() ([]SwapData, error) {
	var parsedSwaps []SwapData

	skip := false
	for i, outerInstruction := range p.txInfo.Message.Instructions {
		progID := p.allAccountKeys[outerInstruction.ProgramIDIndex]
		switch {
		case progID.Equals(JUPITER_PROGRAM_ID):
			skip = true
			parsedSwaps = append(parsedSwaps, p.processJupiterSwaps(i)...)
		case progID.Equals(MOONSHOT_PROGRAM_ID):
			skip = true
			parsedSwaps = append(parsedSwaps, p.processMoonshotSwaps()...)
		case progID.Equals(BANANA_GUN_PROGRAM_ID) ||
			progID.Equals(MINTECH_PROGRAM_ID) ||
			progID.Equals(BLOOM_PROGRAM_ID) ||
			progID.Equals(NOVA_PROGRAM_ID) ||
			progID.Equals(MAESTRO_PROGRAM_ID):
			if innerSwaps := p.processRouterSwaps(i); len(innerSwaps) > 0 {
				parsedSwaps = append(parsedSwaps, innerSwaps...)
			}
		case progID.Equals(OKX_DEX_ROUTER_PROGRAM_ID):
			skip = true
			parsedSwaps = append(parsedSwaps, p.processOKXSwaps(i)...)
		case progID.Equals(AXIOM_PROGRAM_ID) || progID.Equals(AXIOM_PROGRAM_ID2):
			skip = true
			parsedSwaps = append(parsedSwaps, p.processAxionSwaps(i)...)
		case progID.Equals(METEORA_PROGRAM_ID) || progID.Equals(METEORA_POOLS_PROGRAM_ID) || progID.Equals(METEORA_DBC_PROGRAM_ID):
			skip = true
			parsedSwaps = append(parsedSwaps, p.processMeteoraSwaps(i)...)
		}
	}
	if skip {
		return parsedSwaps, nil
	}

	for i, outerInstruction := range p.txInfo.Message.Instructions {
		progID := p.allAccountKeys[outerInstruction.ProgramIDIndex]
		switch {
		case progID.Equals(RAYDIUM_V4_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_CPMM_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_AMM_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_CONCENTRATED_LIQUIDITY_PROGRAM_ID) ||
			p.ContainsProgram(RAYDIUM_Launchpad_PROGRAM_ID) ||
			progID.Equals(solana.MustPublicKeyFromBase58("AP51WLiiqTdbZfgyRMs35PsZpdmLuPDdHYmrB23pEtMU")):
			parsedSwaps = append(parsedSwaps, p.processRaydSwaps(i)...)
		case progID.Equals(ORCA_PROGRAM_ID):
			parsedSwaps = append(parsedSwaps, p.processOrcaSwaps(i)...)
		case progID.Equals(METEORA_PROGRAM_ID) || progID.Equals(METEORA_POOLS_PROGRAM_ID) || progID.Equals(METEORA_DBC_PROGRAM_ID):
			parsedSwaps = append(parsedSwaps, p.processMeteoraSwaps(i)...)
		case progID.Equals(PUMP_FUN_PROGRAM_ID):
			parsedSwaps = append(parsedSwaps, p.processPumpfunSwaps(i)...)
		case progID.Equals(PUMP_AMM_PROGRAM_ID):
			parsedSwaps = append(parsedSwaps, p.processPumpAmmSwaps(i)...) // New handler for PumpSwap
		default:
			// progID.Equals(solana.MustPublicKeyFromBase58("HgoHJy31rnpmm99CaoKn72g1QDLf6A8vzqEKAXCyBFv5")) ||
			// progID.Equals(solana.MustPublicKeyFromBase58("9RR5ZCvUU6rSEtE6iE4xQE4NeP9NMkbsfSsiEHCupj4M")) ||
			// progID.Equals(solana.MustPublicKeyFromBase58("b1oomGGqPKGD6errbyfbVMBuzSC8WtAAYo8MwNafWW1")) ||
			// progID.Equals(solana.MustPublicKeyFromBase58("AxiomQpD1TrYEHNYLts8h3ko1NHdtxfgNgHryj2hJJx4")) ||
			// progID.Equals(solana.MustPublicKeyFromBase58("AzcZqCRUQgKEg5FTAgY7JacATABEYCEfMbjXEzspLYFB"))
			parsedSwaps = append(parsedSwaps, p.processPumpAmmSwaps(i)...)
		}
	}

	return parsedSwaps, nil
}

type PoolData struct {
	PoolType string
	Data     interface{}
}

type SwapInfo struct {
	Signers          []solana.PublicKey
	Signatures       []solana.Signature
	AMMs             []string
	Timestamp        time.Time
	SwapType         string
	PoolData         *PoolData
	TokenInMint      solana.PublicKey
	TokenInAmount    uint64
	TokenInDecimals  uint8
	TokenOutMint     solana.PublicKey
	TokenOutAmount   uint64
	TokenOutDecimals uint8
}

func (p *Parser) ProcessSwapData(swapDatas []SwapData) (*SwapInfo, error) {
	if len(swapDatas) == 0 {
		return nil, fmt.Errorf("no swap data provided")
	}

	swapInfo := &SwapInfo{
		Signatures: p.txInfo.Signatures,
		SwapType:   string(p.SwapType),
	}

	switch p.SwapType {
	case RAYDIUM_Launchpad:
		rlPool := p.getRaydiumLaunchpadPool()
		if rlPool != nil {
			event := p.getRaydiumLaunchpadEvent()
			if event != nil {
				rlPool.RealBaseBefore = event.RealBaseAfter
				rlPool.RealQuoteBefore = event.RealQuoteAfter
				rlPool.VirtualBase = event.VirtualBase
				rlPool.VirtualQuote = event.VirtualQuote
				swapInfo.PoolData = &PoolData{
					PoolType: string(RAYDIUM_Launchpad),
					Data:     rlPool,
				}
			}
		}
	case METEORA_DBC:
		meteoraDbcPoll := p.getMeteoraDbcPool()
		if meteoraDbcPoll != nil {
			event := p.getMeteoraDbcEvent()
			if event != nil {
				meteoraDbcPoll.NextSqrtPrice = event.SwapResult.NextSqrtPrice
			}
			swapInfo.PoolData = &PoolData{
				Data:     meteoraDbcPoll,
				PoolType: string(METEORA_DBC),
			}
		}
	case PUMP_FUN:
		pumpFunPool := p.getPumpFunPool()
		if pumpFunPool != nil {
			event := p.getPumpFunEvent()
			if event != nil {
				pumpFunPool.VirtualSolReserves = event.VirtualSolReserves
				pumpFunPool.VirtualTokenReserves = event.VirtualTokenReserves
				pumpFunPool.RealSOLReserves = event.RealSOLReserves
				pumpFunPool.RealTokenReserves = event.RealTokenReserves
				swapInfo.PoolData = &PoolData{
					PoolType: string(PUMP_FUN),
					Data:     pumpFunPool,
				}
			}
		}
	case PUMP_SWAP:
		pumpAmmPool := p.getPumpAmmPool()
		if pumpAmmPool != nil {
			event := p.getPumpAmmEvent()
			if event != nil {
				pumpAmmPool.PoolBaseTokenReserves = event.PoolBaseTokenReserves
				pumpAmmPool.PoolQuoteTokenReserves = event.PoolQuoteTokenReserves

				swapInfo.PoolData = &PoolData{
					PoolType: string(PUMP_SWAP),
					Data:     pumpAmmPool,
				}
			}

		}
	}

	if p.containsDCAProgram() {
		swapInfo.Signers = []solana.PublicKey{p.allAccountKeys[2]}
	} else {
		swapInfo.Signers = []solana.PublicKey{p.allAccountKeys[0]}
	}

	jupiterSwaps := make([]SwapData, 0)
	pumpfunSwaps := make([]SwapData, 0)
	pumpAmmSwaps := make([]SwapData, 0) // newly added
	otherSwaps := make([]SwapData, 0)

	for _, swapData := range swapDatas {
		switch swapData.Type {
		case JUPITER:
			jupiterSwaps = append(jupiterSwaps, swapData)
		case PUMP_FUN:
			pumpfunSwaps = append(pumpfunSwaps, swapData)
		case PUMP_SWAP:
			pumpAmmSwaps = append(pumpAmmSwaps, swapData)
		default:
			otherSwaps = append(otherSwaps, swapData)
		}
	}

	if len(jupiterSwaps) > 0 {
		jupiterInfo, err := parseJupiterEvents(jupiterSwaps)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Jupiter events: %w", err)
		}

		swapInfo.TokenInMint = jupiterInfo.TokenInMint
		swapInfo.TokenInAmount = jupiterInfo.TokenInAmount
		swapInfo.TokenInDecimals = jupiterInfo.TokenInDecimals
		swapInfo.TokenOutMint = jupiterInfo.TokenOutMint
		swapInfo.TokenOutAmount = jupiterInfo.TokenOutAmount
		swapInfo.TokenOutDecimals = jupiterInfo.TokenOutDecimals
		swapInfo.AMMs = jupiterInfo.AMMs

		return swapInfo, nil
	}

	if len(pumpfunSwaps) > 0 {
		event := pumpfunSwaps[0].Data.(*PumpfunTradeEvent)
		if event.IsBuy {
			swapInfo.TokenInMint = NATIVE_SOL_MINT_PROGRAM_ID
			swapInfo.TokenInAmount = event.SolAmount
			swapInfo.TokenInDecimals = 9
			swapInfo.TokenOutMint = event.Mint
			swapInfo.TokenOutAmount = event.TokenAmount
			swapInfo.TokenOutDecimals = p.splDecimalsMap[event.Mint.String()]
		} else {
			swapInfo.TokenInMint = event.Mint
			swapInfo.TokenInAmount = event.TokenAmount
			swapInfo.TokenInDecimals = p.splDecimalsMap[event.Mint.String()]
			swapInfo.TokenOutMint = NATIVE_SOL_MINT_PROGRAM_ID
			swapInfo.TokenOutAmount = event.SolAmount
			swapInfo.TokenOutDecimals = 9
		}
		swapInfo.AMMs = append(swapInfo.AMMs, string(pumpfunSwaps[0].Type))
		swapInfo.Timestamp = time.Unix(int64(event.Timestamp), 0)
		return swapInfo, nil
	}

	//newly added
	if len(pumpAmmSwaps) > 0 {
		inputAmounts := make(map[string]uint64)
		inputDecimals := make(map[string]uint8)
		outputAmounts := make(map[string]uint64)
		outputDecimals := make(map[string]uint8)

		for _, swap := range pumpAmmSwaps {
			switch data := swap.Data.(type) {
			case *InputTransfer:
				mint := data.Mint
				inputAmounts[mint] += data.Info.Amount
				inputDecimals[mint] = data.Decimals
			case *OutputTransfer:
				mint := data.Mint
				outputAmounts[mint] += data.Info.Amount
				outputDecimals[mint] = data.Decimals
			}
		}

		if len(inputAmounts) > 1 {
			delete(inputAmounts, NATIVE_SOL_MINT_PROGRAM_ID.String())
		}

		if len(inputAmounts) == 1 && len(outputAmounts) == 1 {
			for mint, amount := range inputAmounts {
				swapInfo.TokenInMint = solana.MustPublicKeyFromBase58(mint)
				swapInfo.TokenInAmount = amount
				swapInfo.TokenInDecimals = inputDecimals[mint]
			}
			for mint, amount := range outputAmounts {
				swapInfo.TokenOutMint = solana.MustPublicKeyFromBase58(mint)
				swapInfo.TokenOutAmount = amount
				swapInfo.TokenOutDecimals = outputDecimals[mint]
			}
			swapInfo.AMMs = append(swapInfo.AMMs, string(PUMP_SWAP))
			swapInfo.Timestamp = time.Now() // Update if logs provide timestamp
			return swapInfo, nil
		}
		if len(otherSwaps) == 0 {
			return nil, fmt.Errorf("PumpSwap swap has %d input mints and %d output mints; expected 1 each", len(inputAmounts), len(outputAmounts))
		}
	}

	if len(otherSwaps) > 0 {
		var uniqueTokens []TokenTransfer
		seenTokens := make(map[string]bool)

		for _, swapData := range otherSwaps {
			transfer := getTransferFromSwapData(swapData)
			if transfer != nil && !seenTokens[transfer.mint] {
				uniqueTokens = append(uniqueTokens, *transfer)
				seenTokens[transfer.mint] = true
			}
		}

		if len(uniqueTokens) >= 2 {
			inputTransfer := uniqueTokens[0]
			outputTransfer := uniqueTokens[len(uniqueTokens)-1]

			seenInputs := make(map[string]bool)
			seenOutputs := make(map[string]bool)
			var totalInputAmount uint64 = 0
			var totalOutputAmount uint64 = 0

			swapChange := false
			for _, swapData := range otherSwaps {
				transfer := getTransferFromSwapData(swapData)
				if transfer == nil {
					continue
				}

				amountStr := fmt.Sprintf("%d-%s", transfer.amount, transfer.mint)
				if transfer.mint == inputTransfer.mint && !seenInputs[amountStr] {
					totalInputAmount += transfer.amount
					seenInputs[amountStr] = true
				}
				if transfer.mint == outputTransfer.mint && !seenOutputs[amountStr] {
					totalOutputAmount += transfer.amount
					seenOutputs[amountStr] = true
				}

				if swapData.Type == AXION {
					swapChange = true
				}
			}

			//inputTransfer和outputTransfer 置换
			if swapChange {
				inputTransfer, outputTransfer = outputTransfer, inputTransfer
				totalInputAmount, totalOutputAmount = totalOutputAmount, totalInputAmount
			}

			swapInfo.TokenInMint = solana.MustPublicKeyFromBase58(inputTransfer.mint)
			swapInfo.TokenInAmount = totalInputAmount
			swapInfo.TokenInDecimals = inputTransfer.decimals
			swapInfo.TokenOutMint = solana.MustPublicKeyFromBase58(outputTransfer.mint)
			swapInfo.TokenOutAmount = totalOutputAmount
			swapInfo.TokenOutDecimals = outputTransfer.decimals

			seenAMMs := make(map[string]bool)
			for _, swapData := range otherSwaps {
				if !seenAMMs[string(swapData.Type)] {
					swapInfo.AMMs = append(swapInfo.AMMs, string(swapData.Type))
					seenAMMs[string(swapData.Type)] = true
				}
			}

			swapInfo.Timestamp = time.Now()
			return swapInfo, nil
		}
	}

	return nil, fmt.Errorf("no valid swaps found")
}

func getTransferFromSwapData(swapData SwapData) *TokenTransfer {
	switch data := swapData.Data.(type) {
	case *SystemTransfer:
		return &TokenTransfer{
			user:     data.From,
			mint:     NATIVE_SOL_MINT_PROGRAM_ID.String(),
			amount:   data.Amount,
			decimals: 9,
		}
	case *TransferData:
		return &TokenTransfer{
			user:     data.Info.Source,
			mint:     data.Mint,
			amount:   data.Info.Amount,
			decimals: data.Decimals,
		}
	case *TransferCheck:
		amt, err := strconv.ParseUint(data.Info.TokenAmount.Amount, 10, 64)
		if err != nil {
			return nil
		}
		return &TokenTransfer{
			user:     data.Info.Source,
			mint:     data.Info.Mint,
			amount:   amt,
			decimals: data.Info.TokenAmount.Decimals,
		}
	}
	return nil
}

func (p *Parser) processRouterSwaps(instructionIndex int) []SwapData {
	var swaps []SwapData

	innerInstructions := p.getInnerInstructions(instructionIndex)
	if len(innerInstructions) == 0 {
		return swaps
	}

	processedProtocols := make(map[string]bool)

	for _, inner := range innerInstructions {
		progID := p.allAccountKeys[inner.ProgramIDIndex]

		switch {
		case (progID.Equals(RAYDIUM_V4_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_CPMM_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_AMM_PROGRAM_ID) ||
			progID.Equals(RAYDIUM_CONCENTRATED_LIQUIDITY_PROGRAM_ID)) && !processedProtocols[PROTOCOL_RAYDIUM]:
			processedProtocols[PROTOCOL_RAYDIUM] = true
			if raydSwaps := p.processRaydSwaps(instructionIndex); len(raydSwaps) > 0 {
				swaps = append(swaps, raydSwaps...)
			}

		case progID.Equals(ORCA_PROGRAM_ID) && !processedProtocols[PROTOCOL_ORCA]:
			processedProtocols[PROTOCOL_ORCA] = true
			if orcaSwaps := p.processOrcaSwaps(instructionIndex); len(orcaSwaps) > 0 {
				swaps = append(swaps, orcaSwaps...)
			}

		case (progID.Equals(METEORA_PROGRAM_ID) ||
			progID.Equals(METEORA_POOLS_PROGRAM_ID) || progID.Equals(METEORA_DBC_PROGRAM_ID)) && !processedProtocols[PROTOCOL_METEORA]:
			processedProtocols[PROTOCOL_METEORA] = true
			if meteoraSwaps := p.processMeteoraSwaps(instructionIndex); len(meteoraSwaps) > 0 {
				swaps = append(swaps, meteoraSwaps...)
			}

		case (progID.Equals(PUMP_FUN_PROGRAM_ID) ||
			progID.Equals(solana.MustPublicKeyFromBase58("BSfD6SHZigAfDWSjzD5Q41jw8LmKwtmjskPH9XW1mrRW"))) && !processedProtocols[PROTOCOL_PUMPFUN]:
			processedProtocols[PROTOCOL_PUMPFUN] = true
			if pumpfunSwaps := p.processPumpfunSwaps(instructionIndex); len(pumpfunSwaps) > 0 {
				swaps = append(swaps, pumpfunSwaps...)
			}
		}
	}

	return swaps
}

func (p *Parser) getInnerInstructions(index int) []solana.CompiledInstruction {
	if p.txMeta == nil || p.txMeta.InnerInstructions == nil {
		return nil
	}

	for _, inner := range p.txMeta.InnerInstructions {
		if inner.Index == uint16(index) {
			return inner.Instructions
		}
	}

	return nil
}
