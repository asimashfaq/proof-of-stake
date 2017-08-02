package miner

const (
	//Root Public Keys at initialization time. This is the only existing account at startup
	//All other accounts are created
	INITROOTKEY1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	INITROOTKEY2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"

	//Sha3-256 Hash of the ECC public key (needs to be part of the state)
	BENEFICIARY = "d280f5ea0e5b3d1c98ffb85a1e1acd89b2e7aed8f202517d6087b821c5a48520"

	//How many blocks can we verify dynamically (e.g., proper time check) until we're too far behind
	//that this dynamic check is not possible anymore
	DELAYED_BLOCKS = 10

	//After requesting a tx/block, timeout after this amount of seconds
	TXFETCH_TIMEOUT    = 5
	BLOCKFETCH_TIMEOUT = 40

	//Some prominent programming languages (e.g., Java) have not unsigned integer types
	//Neglecting MSB simplifies compatibility
	MAX_MONEY = 9223372036854775807 //(2^63)-1
)
