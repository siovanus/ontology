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

package cross_chain

import (
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func putRequestID(native *native.NativeService, contract common.Address, requestID uint64, sideChainID uint32) error {
	requestIDBytes, err := utils.GetUint64Bytes(requestID)
	if err != nil {
		return fmt.Errorf("putRequestID, get requestIDBytes error: %v", err)
	}
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return fmt.Errorf("putRequestID, get sideChainIDBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(REQUEST_ID), sideChainIDBytes), cstates.GenRawStorageItem(requestIDBytes))
	return nil
}

func getRequestID(native *native.NativeService, contract common.Address, sideChainID uint32) (uint64, error) {
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return 0, fmt.Errorf("getRequestID, get sideChainIDBytes error: %v", err)
	}
	var requestID uint64 = 0
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(REQUEST_ID), sideChainIDBytes))
	if err != nil {
		return 0, fmt.Errorf("getRequestID, get requestID value error: %v", err)
	}
	if value != nil {
		requestIDBytes, err := cstates.GetValueFromRawStorageItem(value)
		if err != nil {
			return 0, fmt.Errorf("getRequestID, deserialize from raw storage item err:%v", err)
		}
		requestID, err = utils.GetBytesUint64(requestIDBytes)
		if err != nil {
			return 0, fmt.Errorf("getRequestID, get requestID error: %v", err)
		}
	}
	return requestID, nil
}

func putRequest(native *native.NativeService, contract common.Address, requestID uint64, request []byte, sideChainID uint32) error {
	prefix, err := utils.GetUint64Bytes(requestID)
	if err != nil {
		return fmt.Errorf("putRequest, GetUint64Bytes error:%s", err)
	}
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return fmt.Errorf("putRequest, get sideChainIDBytes error: %v", err)
	}
	utils.PutBytes(native, utils.ConcatKey(contract, []byte(REQUEST), sideChainIDBytes, prefix), request)
	return nil
}

//must be called before putCurrentID
func putRemainedIDs(native *native.NativeService, contract common.Address, requestID, currentID uint64, sideChainID uint32) error {
	for i := currentID + 1; i < requestID; i++ {
		requestIDBytes, err := utils.GetUint64Bytes(i)
		if err != nil {
			return fmt.Errorf("putRemainedID, get requestIDBytes error: %v", err)
		}
		sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
		if err != nil {
			return fmt.Errorf("putRemainedID, get sideChainIDBytes error: %v", err)
		}
		native.CacheDB.Put(utils.ConcatKey(contract, []byte(REMAINED_ID), sideChainIDBytes, requestIDBytes), cstates.GenRawStorageItem(requestIDBytes))
	}
	return nil
}

func checkIfRemained(native *native.NativeService, contract common.Address, requestID uint64, sideChainID uint32) (bool, error) {
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return false, fmt.Errorf("checkIfRemained, get sideChainIDBytes error: %v", err)
	}
	requestIDBytes, err := utils.GetUint64Bytes(requestID)
	if err != nil {
		return false, fmt.Errorf("checkIfRemained, get requestIDBytes error: %v", err)
	}
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(REMAINED_ID), sideChainIDBytes, requestIDBytes))
	if err != nil {
		return false, fmt.Errorf("checkIfRemained, get value error: %v", err)
	}
	if value == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func removeRemained(native *native.NativeService, contract common.Address, requestID uint64, sideChainID uint32) error {
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return fmt.Errorf("removeRemained, get sideChainIDBytes error: %v", err)
	}
	requestIDBytes, err := utils.GetUint64Bytes(requestID)
	if err != nil {
		return fmt.Errorf("removeRemained, get requestIDBytes error: %v", err)
	}
	native.CacheDB.Delete(utils.ConcatKey(contract, []byte(REMAINED_ID), sideChainIDBytes, requestIDBytes))
	return nil
}

func putCurrentID(native *native.NativeService, contract common.Address, currentID uint64, sideChainID uint32) error {
	currentIDBytes, err := utils.GetUint64Bytes(currentID)
	if err != nil {
		return fmt.Errorf("putCurrentID, get currentIDBytes error: %v", err)
	}
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return fmt.Errorf("putRequestID, get sideChainIDBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CURRENT_ID), sideChainIDBytes), cstates.GenRawStorageItem(currentIDBytes))
	return nil
}

func getCurrentID(native *native.NativeService, contract common.Address, sideChainID uint32) (uint64, error) {
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return 0, fmt.Errorf("getCurrentID, get sideChainIDBytes error: %v", err)
	}
	var currentID uint64 = 0
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(CURRENT_ID), sideChainIDBytes))
	if err != nil {
		return 0, fmt.Errorf("getCurrentID, get currentID value error: %v", err)
	}
	if value != nil {
		currentIDBytes, err := cstates.GetValueFromRawStorageItem(value)
		if err != nil {
			return 0, fmt.Errorf("getCurrentID, deserialize from raw storage item err:%v", err)
		}
		currentID, err = utils.GetBytesUint64(currentIDBytes)
		if err != nil {
			return 0, fmt.Errorf("getCurrentID, get currentID error: %v", err)
		}
	}
	return currentID, nil
}

func notifyCreateCrossChainTx(native *native.NativeService, contract common.Address, sideChainID uint32, requestID uint64, height uint32) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{CREATE_CROSS_CHAIN_TX, sideChainID, requestID, height},
		})
}

func notifyProcessCrossChainTx(native *native.NativeService, contract common.Address, sideChainID uint32, requestID uint64, height uint32) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{PROCESS_CROSS_CHAIN_TX, sideChainID, requestID, height},
		})
}
