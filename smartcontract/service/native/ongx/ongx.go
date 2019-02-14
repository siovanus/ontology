/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package ongx

import (
	"fmt"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/vm/neovm/types"
)

const (
	//method
	TRANSFER_NAME     = "transfer"
	APPROVE_NAME      = "approve"
	TRANSFERFROM_NAME = "transferFrom"
	NAME_NAME         = "name"
	SYMBOL_NAME       = "symbol"
	DECIMALS_NAME     = "decimals"
	TOTALSUPPLY_NAME  = "totalSupply"
	BALANCEOF_NAME    = "balanceOf"
	ALLOWANCE_NAME    = "allowance"
	ONGX_UNLOCK       = "ongxUnlock"
	ONGX_LOCK         = "ongxLock"

	//prefix
	TOTAL_SUPPLY_NAME = "totalSupply"
	REQUEST_ID        = "requestID"

	TRANSFER_FLAG byte = 1
	APPROVE_FLAG  byte = 2
)

func InitOngx() {
	native.Contracts[utils.OngContractAddress] = RegisterOngContract
}

func RegisterOngContract(native *native.NativeService) {
	native.Register(TRANSFER_NAME, OngxTransfer)
	native.Register(APPROVE_NAME, OngxApprove)
	native.Register(TRANSFERFROM_NAME, OngxTransferFrom)
	native.Register(NAME_NAME, OngxName)
	native.Register(SYMBOL_NAME, OngxSymbol)
	native.Register(DECIMALS_NAME, OngxDecimals)
	native.Register(TOTALSUPPLY_NAME, OngxTotalSupply)
	native.Register(BALANCEOF_NAME, OngxBalanceOf)
	native.Register(ALLOWANCE_NAME, OngxAllowance)
	native.Register(ONGX_UNLOCK, OngxUnlock)
	native.Register(ONGX_LOCK, OngxLock)
}

func OngxTransfer(native *native.NativeService) ([]byte, error) {
	var transfers Transfers
	source := common.NewZeroCopySource(native.Input)
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngTransfer] Transfers deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	for _, v := range transfers.States {
		if v.Value == 0 {
			continue
		}
		if v.Value > constants.ONGX_TOTAL_SUPPLY {
			return utils.BYTE_FALSE, fmt.Errorf("transfer ong amount:%d over totalSupply:%d", v.Value, constants.ONGX_TOTAL_SUPPLY)
		}
		if _, _, err := Transfer(native, contract, &v); err != nil {
			return utils.BYTE_FALSE, err
		}
		AddTransferNotifications(native, contract, &v)
	}
	return utils.BYTE_TRUE, nil
}

func OngxApprove(native *native.NativeService) ([]byte, error) {
	var state State
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.ONGX_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve ong amount:%d over totalSupply:%d", state.Value, constants.ONGX_TOTAL_SUPPLY)
	}
	if native.ContextRef.CheckWitness(state.From) == false {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	native.CacheDB.Put(GenApproveKey(contract, state.From, state.To), utils.GenUInt64StorageItem(state.Value).ToArray())
	return utils.BYTE_TRUE, nil
}

func OngxTransferFrom(native *native.NativeService) ([]byte, error) {
	var state TransferFrom
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.ONGX_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve ong amount:%d over totalSupply:%d", state.Value, constants.ONGX_TOTAL_SUPPLY)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	if _, _, err := TransferedFrom(native, contract, &state); err != nil {
		return utils.BYTE_FALSE, err
	}
	AddTransferNotifications(native, contract, &State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}

func OngxName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONGX_NAME), nil
}

func OngxDecimals(native *native.NativeService) ([]byte, error) {
	return big.NewInt(int64(constants.ONGX_DECIMALS)).Bytes(), nil
}

func OngxSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONGX_SYMBOL), nil
}

func OngxTotalSupply(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := utils.GetStorageUInt64(native, GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTotalSupply] get totalSupply error!")
	}
	return types.BigIntToBytes(big.NewInt(int64(amount))), nil
}

func OngxBalanceOf(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, TRANSFER_FLAG)
}

func OngxAllowance(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, APPROVE_FLAG)
}

func OngxUnlock(native *native.NativeService) ([]byte, error) {
	context := native.ContextRef.CurrentContext().ContractAddress
	source := common.NewZeroCopySource(native.Input)
	var param OngxUnlockParam
	if err := param.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngUnlock] error:%s", err)
	}
	totalSupplyKey := GenTotalSupplyKey(context)
	amount, err := utils.GetStorageUInt64(native, totalSupplyKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngUnlock] error:%s", err)
	}

	//TODO: auth check
	key := append(context[:], param.Addr[:]...)
	balance, err := utils.GetStorageUInt64(native, key)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngSwap] error:%s", err)
	}
	native.CacheDB.Put(key, utils.GenUInt64StorageItem(balance+param.Value).ToArray())
	var ok bool
	amount, ok = common.SafeAdd(amount, param.Value)
	if ok {
		return utils.BYTE_FALSE, fmt.Errorf("[OngSwap] total supply is more than MAX_UINT64")
	}
	if amount > constants.ONGX_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("[OngSwap] total supply is more than constants.ONGX_TOTAL_SUPPLY")
	}
	AddOngxUnlockNotifications(native, context, &State{To: param.Addr, Value: param.Value})

	native.CacheDB.Put(totalSupplyKey, utils.GenUInt64StorageItem(amount).ToArray())
	return utils.BYTE_TRUE, nil
}

func OngxLock(native *native.NativeService) ([]byte, error) {
	context := native.ContextRef.CurrentContext().ContractAddress
	source := common.NewZeroCopySource(native.Input)
	var param OngxLockParam
	if err := param.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxLock] error:%s", err)
	}
	if !native.ContextRef.CheckWitness(param.Addr) {
		return utils.BYTE_FALSE, errors.NewErr("[OngxLock] authentication failed!")
	}
	key := append(context[:], param.Addr[:]...)
	balance, err := utils.GetStorageUInt64(native, key)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxLock] error:%s", err)
	}
	if param.Value > balance {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxLock] swap ongx balance insufficient! have %d, want %d", balance, param.Value)
	} else if param.Value == balance {
		native.CacheDB.Delete(key)
	} else {
		native.CacheDB.Put(key, utils.GenUInt64StorageItem(balance-param.Value).ToArray())
	}

	//record ongx lock amount
	requestID, err := getRequestID(native, context)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxLock] getRequestID error:%s", err)
	}
	newID := requestID + 1
	err = putRequest(native, context, newID, native.Input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxLock] putRequest error:%s", err)
	}
	err = putRequestID(native, context, newID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxLock] putRequestID error:%s", err)
	}

	//update total supply
	totalSupplyKey := GenTotalSupplyKey(context)
	amount, err := utils.GetStorageUInt64(native, totalSupplyKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxLock] error:%s", err)
	}
	native.CacheDB.Put(totalSupplyKey, utils.GenUInt64StorageItem(amount-param.Value).ToArray())

	AddOngxLockNotifications(native, context, newID, native.Height, param.Addr, param.Value)
	return utils.BYTE_TRUE, nil
}
