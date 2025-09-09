package solanaswapgo

import "github.com/gagliardetto/solana-go"

var (
	JUPITER_PROGRAM_ID                        = solana.MustPublicKeyFromBase58("JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4")
	JUPITER_DCA_PROGRAM_ID                    = solana.MustPublicKeyFromBase58("DCAK36VfExkPdAkYUQg6ewgxyinvcEyPLyHjRbmveKFw")
	PUMP_FUN_PROGRAM_ID                       = solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P")
	PUMP_AMM_PROGRAM_ID                       = solana.MustPublicKeyFromBase58("pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA") //newly added
	PHOENIX_PROGRAM_ID                        = solana.MustPublicKeyFromBase58("PhoeNiXZ8ByJGLkxNfZRnkUfjvmuYqLR89jjFHGqdXY") // not supported yet
	BANANA_GUN_PROGRAM_ID                     = solana.MustPublicKeyFromBase58("BANANAjs7FJiPQqJTGFzkZJndT9o7UmKiYYGaJz6frGu")
	MINTECH_PROGRAM_ID                        = solana.MustPublicKeyFromBase58("minTcHYRLVPubRK8nt6sqe2ZpWrGDLQoNLipDJCGocY")
	BLOOM_PROGRAM_ID                          = solana.MustPublicKeyFromBase58("b1oomGGqPKGD6errbyfbVMBuzSC8WtAAYo8MwNafWW1")
	MAESTRO_PROGRAM_ID                        = solana.MustPublicKeyFromBase58("MaestroAAe9ge5HTc64VbBQZ6fP77pwvrhM8i1XWSAx")
	NOVA_PROGRAM_ID                           = solana.MustPublicKeyFromBase58("NoVA1TmDUqksaj2hB1nayFkPysjJbFiU76dT4qPw2wm")
	RAYDIUM_V4_PROGRAM_ID                     = solana.MustPublicKeyFromBase58("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8")
	RAYDIUM_AMM_PROGRAM_ID                    = solana.MustPublicKeyFromBase58("routeUGWgWzqBWFcrCfv8tritsqukccJPu3q5GPP3xS")
	RAYDIUM_CPMM_PROGRAM_ID                   = solana.MustPublicKeyFromBase58("CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C")
	RAYDIUM_Launchpad_PROGRAM_ID              = solana.MustPublicKeyFromBase58("LanMV9sAd7wArD4vJFi2qDdfnVhFxYSUg6eADduJ3uj")
	RAYDIUM_Launchpad_Migration_PROGRAM_ID    = solana.MustPublicKeyFromBase58("RAYpQbFNq9i3mu6cKpTKKRwwHFDeK5AuZz8xvxUrCgw")
	RAYDIUM_CONCENTRATED_LIQUIDITY_PROGRAM_ID = solana.MustPublicKeyFromBase58("CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK")
	METEORA_ALL                               = []solana.PublicKey{
		solana.MustPublicKeyFromBase58("LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"),
		solana.MustPublicKeyFromBase58("dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN"),
		solana.MustPublicKeyFromBase58("Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB"),
		solana.MustPublicKeyFromBase58("dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN"),
	}
	METEORA_PROGRAM_ID         = solana.MustPublicKeyFromBase58("LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo")
	METEORA_DBC_PROGRAM_ID     = solana.MustPublicKeyFromBase58("dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN") //Meteora Dynamic Bonding Curve Program
	METEORA_POOLS_PROGRAM_ID   = solana.MustPublicKeyFromBase58("Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB")
	MOONSHOT_PROGRAM_ID        = solana.MustPublicKeyFromBase58("MoonCVVNZFSYkqNXP6bxHLPL6QQJiMagDL3qcqUQTrG")
	ORCA_PROGRAM_ID            = solana.MustPublicKeyFromBase58("whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc")
	OKX_DEX_ROUTER_PROGRAM_ID  = solana.MustPublicKeyFromBase58("6m2CDdhRgxpH4WjvdzxAYbGxwdGUz5MziiL5jek2kBma")
	PHOTON_PROGRAM_ID          = solana.MustPublicKeyFromBase58("BSfD6SHZigAfDWSjzD5Q41jw8LmKwtmjskPH9XW1mrRW")
	AXIOM_PROGRAM_ID2          = solana.MustPublicKeyFromBase58("AxiomQpD1TrYEHNYLts8h3ko1NHdtxfgNgHryj2hJJx4")
	AXIOM_PROGRAM_ID           = solana.MustPublicKeyFromBase58("Axiom3a2w1UbMt2SMgqSvRiuJFTPusDhwKamNgPTeNQ9")
	NATIVE_SOL_MINT_PROGRAM_ID = solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
)

type SwapType string

const (
	PUMP_FUN          SwapType = "PumpFun"
	PUMP_SWAP         SwapType = "PumpAmm" //newly added
	JUPITER           SwapType = "Jupiter"
	RAYDIUM           SwapType = "Raydium"
	RAYDIUM_Launchpad SwapType = "RaydiumLaunchpad"
	OKX               SwapType = "OKX"
	ORCA              SwapType = "Orca"
	METEORA           SwapType = "Meteora"
	METEORA_DBC       SwapType = "MeteoraDbc"
	AXION             SwapType = "Axion"
	MOONSHOT          SwapType = "Moonshot"
	UNKNOWN           SwapType = "Unknown"
)
