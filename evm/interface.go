// Copyright 2016 The go-ethereum Authors
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
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// StateDB is an EVM database for full state querying.
type StateDB interface {
	CreateEVMAccount(common.Address)

	SubEVMBalance(common.Address, *big.Int)
	AddEVMBalance(common.Address, *big.Int)
	GetEVMBalance(common.Address) *big.Int

	GetEVMNonce(common.Address) uint64
	SetEVMNonce(common.Address, uint64)

	GetEVMCodeHash(common.Address) common.Hash
	GetEVMCode(common.Address) []byte
	SetEVMCode(common.Address, []byte)
	GetEVMCodeSize(common.Address) int

	AddEVMRefund(uint64)
	SubEVMRefund(uint64)
	GetEVMRefund() uint64

	GetEVMCommittedState(common.Address, common.Hash) common.Hash
	GetEVMState(common.Address, common.Hash) common.Hash
	SetEVMState(common.Address, common.Hash, common.Hash)

	SuisideEVM(common.Address) bool
	HasSuisideEVM(common.Address) bool

	// Exist reports whether the given account exists in state.
	// Notably this should also return true for suicided accounts.
	ExistEVM(common.Address) bool
	// Empty returns whether the given account is empty. Empty
	// is defined according to EIP161 (balance = nonce = code = 0).
	EmptyEVM(common.Address) bool

	PrepareEVMAccessList(sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList)
	AddressInEVMAccessList(addr common.Address) bool
	SlotInEVMAceessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool)
	// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
	// even if the feature/fork is not active yet
	AddAddressToEVMAccessList(addr common.Address)
	// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
	// even if the feature/fork is not active yet
	AddSlotToEVMAccessList(addr common.Address, slot common.Hash)

	RevertToSnapshot(int)
	Snapshot() int

	AddEVMLog(*types.Log)
	AddEVMPreimage(common.Hash, []byte)

	GetBlockEVMHash(uint64) common.Hash
	PrepareEVM(common.Hash, int)
}

// CallContext provides a basic interface for the EVM calling conventions. The EVM
// depends on this context being implemented for doing subcalls and initialising new EVM contracts.
type CallContext interface {
	// Call another contract
	Call(env *EVM, me ContractRef, addr common.Address, data []byte, gas, value *big.Int) ([]byte, error)
	// Take another's contract code and execute within our own context
	CallCode(env *EVM, me ContractRef, addr common.Address, data []byte, gas, value *big.Int) ([]byte, error)
	// Same as CallCode except sender and value is propagated from parent to child scope
	DelegateCall(env *EVM, me ContractRef, addr common.Address, data []byte, gas *big.Int) ([]byte, error)
	// Create a new contract
	Create(env *EVM, me ContractRef, data []byte, gas, value *big.Int) ([]byte, common.Address, error)
}
