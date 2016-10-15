// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

// These are the minimum amount the gas price can be.
// It is set to the current gas price on ethereum so miners can decide what these become.
var (
	MinGasQuickStep   = big.NewInt(2)
	MinGasFastestStep = big.NewInt(3)
	MinGasFastStep    = big.NewInt(5)
	MinGasMidStep     = big.NewInt(8)
	MinGasSlowStep    = big.NewInt(10)
	MinGasExtStep     = big.NewInt(20)
	
	// TODO: Replace these with current values.
	MinGasSload = big.NewInt(500)
	MinGasStore = big.NewInt(500)
	MinGasSha3 = big.NewInt(30)
	MinGasCreate = big.NewInt(500)
	MinGasCall = big.NewInt(500)
	MinGasJumpdest = big.NewInt(10)
	MinGasSuicide = big.NewInt(0)
	MinGasBalance = big.NewInt(20)
	MinGasExtcodesize = big.NewInt(20)
	MinGasExtcodecopy = big.NewInt(20)

	GasReturn = big.NewInt(0)
	GasStop   = big.NewInt(0)

	GasContractByte = big.NewInt(200)
)

// These are the params which are targetable.
type dynamicGas struct {
	// Step opcodes (affects multiple opcodes).
    quickStep *big.Int
    fastestStep *big.Int
	fastStep *big.Int
	midStep *big.Int
	slowStep *big.Int
	extStep *big.Int
	
	// Seperate opcodes.
	sload *big.Int
	sstore *big.Int
	sha3 *big.Int
	create *big.Int
	call *big.Int
	jumpdest *big.Int
	suicide *big.Int
	balance *big.Int
	extcodesize *big.Int
	extcodecopy *big.Int
}

// This should be used to check the opcode targets per block.
// Then every 64 blocks it should then use those to retarget the gas prices.
func checkGasPricing(*dynamicGas dynGas) error {
	
	// Check to make sure that we are not setting the gas price too low.
	// Preventing of reaching 0 gas for function (inf loop) and really cheap for attacks.
	if (dynGas.quickStep.Cmp(MinGasQuickStep) == -1 || 
		dynGas.fastestStep.Cmp(MinGasFastestStep) == -1 || 
		dynGas.fastStep.Cmp(MinGasFastStep) == -1 || 
		dynGas.midStep.Cmp(MinGasMidStep) == -1 || 
		dynGas.slowStep.Cmp(MinGasSlowStep) == -1 || 
		dynGas.extStep.Cmp(MinGasExtStep) == -1 || 
		// speific opcodes.
		dynGas.sload.Cmp(MinGasSload) == -1 || 
		dynGas.sstore.Cmp(MinGasStore) == -1 || 
		dynGas.sha3.Cmp(MinGasSha3) == -1 || 
		dynGas.create.Cmp(MinGasCreate) == -1 || 
		dynGas.call.Cmp(MinGasCall) == -1 || 
		dynGas.jumpdest.Cmp(MinGasJumpdest) == -1 || 
		dynGas.suicide.Cmp(MinGasSuicide) == -1 || 
		dynGas.balance.Cmp(MinGasBalance) == -1 || 
		dynGas.extcodesize.Cmp(MinGasExtcodesize) == -1 || 
		dynGas.extcodecopy.Cmp(MinGasExtcodecopy) == -1) {
			return fmt.Errorf("Block tried to set gas price of opcode too low.")
		}
		
	// Also make sure this block is not trying to retarget too high.
	// This is to prevent targeting 9999999999 to increase to an insane value even if the other 63 vote low.
	// So the max a block can vote is an 2x increase / decrease.
	// This does mean that if the target goes above 0 then it can never get back to zero and 1 would be the minimum.
	// Keep the divisor constant, don't create a new bigint each calc.
	divisor := big.NewInt(2)
	
	if (dynGas.quickStep.Cmp(big.NewInt(0).Mul(_baseCheck[GAS].gas, divisor)) == 1 ||
		dynGas.quickStep.Cmp(big.NewInt(0).Div(_baseCheck[GAS].gas, divisor)) == -1 ||
		
		dynGas.fastestStep.Cmp(big.NewInt(0).Mul(_baseCheck[ADD].gas, divisor)) == 1 ||
		dynGas.fastestStep.Cmp(big.NewInt(0).Div(_baseCheck[ADD].gas, divisor)) == -1 ||
		
		dynGas.fastStep.Cmp(big.NewInt(0).Mul(_baseCheck[MOD].gas, divisor)) == 1 ||
		dynGas.fastStep.Cmp(big.NewInt(0).Div(_baseCheck[MOD].gas, divisor)) == -1 ||
		
		dynGas.midStep.Cmp(big.NewInt(0).Mul(_baseCheck[JUMP].gas, divisor)) == 1 ||
		dynGas.midStep.Cmp(big.NewInt(0).Div(_baseCheck[JUMP].gas, divisor)) == -1 ||
		
		dynGas.slowStep.Cmp(big.NewInt(0).Mul(_baseCheck[EXP].gas, divisor)) == 1 ||
		dynGas.slowStep.Cmp(big.NewInt(0).Div(_baseCheck[EXP].gas, divisor)) == -1 ||
		
		dynGas.extStep.Cmp(big.NewInt(0).Mul(_baseCheck[BLOCKHASH].gas, divisor)) == 1 ||
		dynGas.extStep.Cmp(big.NewInt(0).Div(_baseCheck[BLOCKHASH].gas, divisor)) == -1 ||
		
		dynGas.sload.Cmp(big.NewInt(0).Mul(_baseCheck[SLOAD].gas, divisor)) == 1 ||
		dynGas.sload.Cmp(big.NewInt(0).Div(_baseCheck[SLOAD].gas, divisor)) == -1 ||
		
		dynGas.sstore.Cmp(big.NewInt(0).Mul(_baseCheck[SSTORE].gas, divisor)) == 1 ||
		dynGas.sstore.Cmp(big.NewInt(0).Div(_baseCheck[SSTORE].gas, divisor)) == -1// ||
		
		// Continue...
		) {
			return fmt.Errorf("Block tried to set gas price of opcode too high.")
		}

	return nil
}

// Update the gas pricing with the last subset of blocks.
// The gas is not stored and will need to be recalculated from the last 64 blocks.
func updateGasPricing(*dynamicGas dynGas) error {

	// Make sure that the gas target is valid.
	// Since each block should be checked this should never happen.
	gasCheck := checkGasPricing(dynGas)
	
	if (gasCheck != nil) {
		return gasCheck
	}

	_baseCheck = map[OpCode]req{
		// opcode  |  stack pop | gas price | stack push
		ADD:          {2, dynamicGas.fastestStep, 1},
		LT:           {2, dynamicGas.fastestStep, 1},
		GT:           {2, dynamicGas.fastestStep, 1},
		SLT:          {2, dynamicGas.fastestStep, 1},
		SGT:          {2, dynamicGas.fastestStep, 1},
		EQ:           {2, dynamicGas.fastestStep, 1},
		ISZERO:       {1, dynamicGas.fastestStep, 1},
		SUB:          {2, dynamicGas.fastestStep, 1},
		AND:          {2, dynamicGas.fastestStep, 1},
		OR:           {2, dynamicGas.fastestStep, 1},
		XOR:          {2, dynamicGas.fastestStep, 1},
		NOT:          {1, dynamicGas.fastestStep, 1},
		BYTE:         {2, dynamicGas.fastestStep, 1},
		CALLDATALOAD: {1, dynamicGas.fastestStep, 1},
		CALLDATACOPY: {3, dynamicGas.fastestStep, 1},
		MLOAD:        {1, dynamicGas.fastestStep, 1},
		MSTORE:       {2, dynamicGas.fastestStep, 0},
		MSTORE8:      {2, dynamicGas.fastestStep, 0},
		CODECOPY:     {3, dynamicGas.fastestStep, 0},
		MUL:          {2, dynamicGas.fastStep, 1},
		DIV:          {2, dynamicGas.fastStep, 1},
		SDIV:         {2, dynamicGas.fastStep, 1},
		MOD:          {2, dynamicGas.fastStep, 1},
		SMOD:         {2, dynamicGas.fastStep, 1},
		SIGNEXTEND:   {2, dynamicGas.fastStep, 1},
		ADDMOD:       {3, dynamicGas.midStep, 1},
		MULMOD:       {3, dynamicGas.midStep, 1},
		JUMP:         {1, dynamicGas.midStep, 0},
		JUMPI:        {2, dynamicGas.slowStep, 0},
		EXP:          {2, dynamicGas.slowStep, 1},
		ADDRESS:      {0, dynamicGas.quickStep, 1},
		ORIGIN:       {0, dynamicGas.quickStep, 1},
		CALLER:       {0, dynamicGas.quickStep, 1},
		CALLVALUE:    {0, dynamicGas.quickStep, 1},
		CODESIZE:     {0, dynamicGas.quickStep, 1},
		GASPRICE:     {0, dynamicGas.quickStep, 1},
		COINBASE:     {0, dynamicGas.quickStep, 1},
		TIMESTAMP:    {0, dynamicGas.quickStep, 1},
		NUMBER:       {0, dynamicGas.quickStep, 1},
		CALLDATASIZE: {0, dynamicGas.quickStep, 1},
		DIFFICULTY:   {0, dynamicGas.quickStep, 1},
		GASLIMIT:     {0, dynamicGas.quickStep, 1},
		POP:          {1, dynamicGas.quickStep, 0},
		PC:           {0, dynamicGas.quickStep, 1},
		MSIZE:        {0, dynamicGas.quickStep, 1},
		GAS:          {0, dynamicGas.quickStep, 1},
		BLOCKHASH:    {1, dynamicGas.extStep, 1},
		BALANCE:      {1, dynamicGas.balance, 1},
		EXTCODESIZE:  {1, dynamicGas.extcodesize, 1},
		EXTCODECOPY:  {4, dynamicGas.extcodecopy, 0},
		SLOAD:        {1, dynamicGas.sload, 1},
		SSTORE:       {2, dynamicGas.sstore, 0},
		SHA3:         {2, dynamicGas.sha3, 1},
		CREATE:       {3, dynamicGas.create, 1},
		CALL:         {7, dynamicGas.call, 1},
		CALLCODE:     {7, dynamicGas.call, 1},
		DELEGATECALL: {6, dynamicGas.call, 1},
		JUMPDEST:     {0, dynamicGas.jumpdest, 0},
		SUICIDE:      {1, dynamicGas.suicide, 0},
		RETURN:       {2, Zero, 0},
		PUSH1:        {0, dynamicGas.fastestStep, 1},
		DUP1:         {0, Zero, 1},
	}
	
	return nil
}

// baseCheck checks for any stack error underflows
func baseCheck(op OpCode, stack *stack, gas *big.Int) error {
	// PUSH and DUP are a bit special. They all cost the same but we do want to have checking on stack push limit
	// PUSH is also allowed to calculate the same price for all PUSHes
	// DUP requirements are handled elsewhere (except for the stack limit check)
	if op >= PUSH1 && op <= PUSH32 {
		op = PUSH1
	}
	if op >= DUP1 && op <= DUP16 {
		op = DUP1
	}

	if r, ok := _baseCheck[op]; ok {
		err := stack.require(r.stackPop)
		if err != nil {
			return err
		}

		if r.stackPush > 0 && stack.len()-r.stackPop+r.stackPush > int(params.StackLimit.Int64()) {
			return fmt.Errorf("stack limit reached %d (%d)", stack.len(), params.StackLimit.Int64())
		}

		gas.Add(gas, r.gas)
	}
	return nil
}

// casts a arbitrary number to the amount of words (sets of 32 bytes)
func toWordSize(size *big.Int) *big.Int {
	tmp := new(big.Int)
	tmp.Add(size, u256(31))
	tmp.Div(tmp, u256(32))
	return tmp
}

type req struct {
	stackPop  int
	gas       *big.Int
	stackPush int
}


var _baseCheck = map[OpCode]req{
	// opcode  |  stack pop | gas price | stack push
	ADD:          {2, MinGasFastestStep, 1},
	LT:           {2, MinGasFastestStep, 1},
	GT:           {2, MinGasFastestStep, 1},
	SLT:          {2, MinGasFastestStep, 1},
	SGT:          {2, MinGasFastestStep, 1},
	EQ:           {2, MinGasFastestStep, 1},
	ISZERO:       {1, MinGasFastestStep, 1},
	SUB:          {2, MinGasFastestStep, 1},
	AND:          {2, MinGasFastestStep, 1},
	OR:           {2, MinGasFastestStep, 1},
	XOR:          {2, MinGasFastestStep, 1},
	NOT:          {1, MinGasFastestStep, 1},
	BYTE:         {2, MinGasFastestStep, 1},
	CALLDATALOAD: {1, MinGasFastestStep, 1},
	CALLDATACOPY: {3, MinGasFastestStep, 1},
	MLOAD:        {1, MinGasFastestStep, 1},
	MSTORE:       {2, MinGasFastestStep, 0},
	MSTORE8:      {2, MinGasFastestStep, 0},
	CODECOPY:     {3, MinGasFastestStep, 0},
	MUL:          {2, MinGasFastStep, 1},
	DIV:          {2, MinGasFastStep, 1},
	SDIV:         {2, MinGasFastStep, 1},
	MOD:          {2, MinGasFastStep, 1},
	SMOD:         {2, MinGasFastStep, 1},
	SIGNEXTEND:   {2, MinGasFastStep, 1},
	ADDMOD:       {3, MinGasMidStep, 1},
	MULMOD:       {3, MinGasMidStep, 1},
	JUMP:         {1, MinGasMidStep, 0},
	JUMPI:        {2, MinGasSlowStep, 0},
	EXP:          {2, MinGasSlowStep, 1},
	ADDRESS:      {0, MinGasQuickStep, 1},
	ORIGIN:       {0, MinGasQuickStep, 1},
	CALLER:       {0, MinGasQuickStep, 1},
	CALLVALUE:    {0, MinGasQuickStep, 1},
	CODESIZE:     {0, MinGasQuickStep, 1},
	GASPRICE:     {0, MinGasQuickStep, 1},
	COINBASE:     {0, MinGasQuickStep, 1},
	TIMESTAMP:    {0, MinGasQuickStep, 1},
	NUMBER:       {0, MinGasQuickStep, 1},
	CALLDATASIZE: {0, MinGasQuickStep, 1},
	DIFFICULTY:   {0, MinGasQuickStep, 1},
	GASLIMIT:     {0, MinGasQuickStep, 1},
	POP:          {1, MinGasQuickStep, 0},
	PC:           {0, MinGasQuickStep, 1},
	MSIZE:        {0, MinGasQuickStep, 1},
	GAS:          {0, MinGasQuickStep, 1},
	BLOCKHASH:    {1, MinGasExtStep, 1},
	BALANCE:      {1, MinGasExtStep, 1},
	EXTCODESIZE:  {1, MinGasExtStep, 1},
	EXTCODECOPY:  {4, MinGasExtStep, 0},
	SLOAD:        {1, params.SloadGas, 1},
	SSTORE:       {2, Zero, 0},
	SHA3:         {2, params.Sha3Gas, 1},
	CREATE:       {3, params.CreateGas, 1},
	CALL:         {7, params.CallGas, 1},
	CALLCODE:     {7, params.CallGas, 1},
	DELEGATECALL: {6, params.CallGas, 1},
	JUMPDEST:     {0, params.JumpdestGas, 0},
	SUICIDE:      {1, Zero, 0},
	RETURN:       {2, Zero, 0},
	PUSH1:        {0, GasFastestStep, 1},
	DUP1:         {0, Zero, 1},
}