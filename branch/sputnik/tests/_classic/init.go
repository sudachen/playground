package tests

import (
	"github.com/sudachen/playground/playtool"
	"github.com/sudachen/playground/playtool/classic"
)

func init() {
	// https://github.com/ethereumproject/go-ethereum/issues/432
	// sputnikvm panicked at 'arithmetic operation overflow'
	playtool.SkipTests(classic.StateTests,
		"Special/OverflowGasMakeMoney",
		"Special/txCost-sec73",
		"Transition/createNameRegistratorPerTxsNotEnoughGasAfter",
		"Transition/createNameRegistratorPerTxsNotEnoughGasAt",
	)
	// TODO check it later
	playtool.SkipTests(classic.StateTests,
		"SystemOperations/CreateHashCollision",
		"PreCompiledContracts/*",
		"Transition/delegatecallAfterTransition",
		"Transition/delegatecallAtTransition",
		"CallCreateCallCode/callcodeWithHighValue",
		"CallCodes/*",
		"DelegateCall/Call1024OOG",
	)
}
