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
	"github.com/ontio/ontology/rlp"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/header_sync"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/trie"
)

const (
	CREATE_CROSS_CHAIN_TX  = "createCrossChainTx"
	PROCESS_CROSS_CHAIN_TX = "processCrossChainTx"

	//key prefix
	REQUEST_ID  = "requestID"
	REQUEST     = "request"
	CURRENT_ID  = "currentID"
	REMAINED_ID = "remainedID"
)

//Init governance contract address
func InitCrossChain() {
	native.Contracts[utils.CrossChainContractAddress] = RegisterCrossChianContract
}

//Register methods of governance contract
func RegisterCrossChianContract(native *native.NativeService) {
	native.Register(CREATE_CROSS_CHAIN_TX, CreateCrossChainTx)
	native.Register(PROCESS_CROSS_CHAIN_TX, ProcessCrossChainTx)
}

func CreateCrossChainTx(native *native.NativeService) ([]byte, error) {
	params := new(CreateCrossChainTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, contract params deserialize error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//record cross chain tx
	requestID, err := getRequestID(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, getRequestID error:%s", err)
	}
	newID := requestID + 1
	err = putRequest(native, contract, newID, native.Input, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, putRequest error:%s", err)
	}
	err = putRequestID(native, contract, newID, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateCrossChainTx, putRequestID error:%s", err)
	}
	notifyCreateCrossChainTx(native, contract, params.SideChainID, newID, native.Height)
	return utils.BYTE_TRUE, nil
}

func ProcessCrossChainTx(native *native.NativeService) ([]byte, error) {
	params := new(ProcessCrossChainTxParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, contract params deserialize error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//record done cross chain tx
	oldCurrentID, err := getCurrentID(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, getCurrentID error: %v", err)
	}
	if params.ID > oldCurrentID {
		err = putRemainedIDs(native, contract, params.ID, oldCurrentID, params.SideChainID)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, putRemainedIDs error: %v", err)
		}
		err = putCurrentID(native, contract, params.ID, params.SideChainID)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, putCurrentID error: %v", err)
		}
	} else {
		ok, err := checkIfRemained(native, contract, params.ID, params.SideChainID)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, checkIfRemained error: %v", err)
		}
		if !ok {
			return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, tx already done")
		} else {
			err = removeRemained(native, contract, params.ID, params.SideChainID)
			if err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, removeRemained error: %v", err)
			}
		}
	}

	//get block header
	header, err := header_sync.GetHeaderByHeight(native, params.SideChainID, params.Height)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetHeaderByHeight, get header by height error: %v", err)
	}

	prefix, err := utils.GetUint64Bytes(params.ID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, GetUint64Bytes error:%s", err)
	}
	sideChainIDBytes, err := utils.GetUint32Bytes(config.DefConfig.Genesis.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ProcessCrossChainTx, GetUint32Bytes error:%s", err)
	}
	//verify mpt
	proof := make([]rlp.RawValue, 0, len(params.Proof))
	for _, v := range params.Proof {
		proof = append(proof, v)
	}
	key := utils.ConcatKey(utils.CrossChainContractAddress, []byte(REQUEST), sideChainIDBytes, prefix)
	value, err := trie.VerifyProof(header.StatesRoot, key, proof)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("trie.VerifyProof, verify mpt proof error: %v", err)
	}
	s := common.NewZeroCopySource(value)
	crossChainParam := new(CreateCrossChainTxParam)
	if err := crossChainParam.Deserialization(s); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("crossChainParam.Deserialization, deserialize CreateCrossChainTxParam error:%s", err)
	}

	//call cross chain function
	destContractAddr := crossChainParam.ContractAddress
	functionName := crossChainParam.FunctionName
	args := crossChainParam.Args
	if destContractAddr == utils.OngContractAddress {
		if _, err := native.NativeCall(destContractAddr, functionName, args); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("native.NativeCall, appCall error: %v", err)
		}
	}
	notifyProcessCrossChainTx(native, contract, params.SideChainID, params.ID, params.Height)
	return utils.BYTE_TRUE, nil
}
