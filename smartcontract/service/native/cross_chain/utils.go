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
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/merkle"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/header_sync"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func appCallTransferOng(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	err := appCallTransfer(native, utils.OngContractAddress, from, to, amount)
	if err != nil {
		return fmt.Errorf("appCallTransferOng, appCallTransfer error: %v", err)
	}
	return nil
}

func appCallTransfer(native *native.NativeService, contract common.Address, from common.Address, to common.Address, amount uint64) error {
	var sts []ont.State
	sts = append(sts, ont.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := ont.Transfers{
		States: sts,
	}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)

	if _, err := native.NativeCall(contract, "transfer", sink.Bytes()); err != nil {
		return fmt.Errorf("appCallTransfer, appCall error: %v", err)
	}
	return nil
}

func putDoneTx(native *native.NativeService, txHash common.Uint256, chainID uint64) error {
	contract := utils.CrossChainContractAddress
	prefix := txHash.ToArray()
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("putRequestID, get chainIDBytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, prefix), cstates.GenRawStorageItem(txHash.ToArray()))
	return nil
}

func checkDoneTx(native *native.NativeService, txHash common.Uint256, chainID uint64) error {
	contract := utils.CrossChainContractAddress
	prefix := txHash.ToArray()
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("checkDoneTx, get chainIDBytes error: %v", err)
	}
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(DONE_TX), chainIDBytes, prefix))
	if err != nil {
		return fmt.Errorf("checkDoneTx, native.CacheDB.Get error: %v", err)
	}
	if value != nil {
		return fmt.Errorf("checkDoneTx, tx already done")
	}
	return nil
}

func putRequest(native *native.NativeService, txHash common.Uint256, chainID uint64, request []byte) error {
	contract := utils.CrossChainContractAddress
	prefix := txHash.ToArray()
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("putRequest, get chainIDBytes error: %v", err)
	}
	utils.PutBytes(native, utils.ConcatKey(contract, []byte(REQUEST), chainIDBytes, prefix), request)
	return nil
}

func MakeFromOntProof(native *native.NativeService, params *CreateCrossChainTxParam) error {
	//record cross chain tx
	merkleValue := &FromMerkleValue{
		TxHash: native.Tx.Hash(),
		CreateCrossChainTxMerkle: &CreateCrossChainTxMerkle{
			FromChainID:         native.ShardID.ToUint64(),
			FromContractAddress: native.ContextRef.CallingContext().ContractAddress.ToHexString(),
			ToChainID:           params.ToChainID,
			Fee:                 params.Fee,
			ToAddress:           params.ToAddress,
			Amount:              params.Amount,
		},
	}
	sink := common.NewZeroCopySink(nil)
	merkleValue.Serialization(sink)
	err := putRequest(native, merkleValue.TxHash, params.ToChainID, sink.Bytes())
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, putRequest error:%s", err)
	}
	native.ContextRef.PutMerkleVal(sink.Bytes())
	prefix := merkleValue.TxHash.ToArray()
	chainIDBytes, err := utils.GetUint64Bytes(params.ToChainID)
	if err != nil {
		return fmt.Errorf("MakeFromOntProof, get chainIDBytes error: %v", err)
	}
	key := hex.EncodeToString(utils.ConcatKey(utils.CrossChainContractAddress, []byte(REQUEST), chainIDBytes, prefix))
	notifyMakeFromOntProof(native, merkleValue.TxHash.ToHexString(), params.ToChainID, key)
	return nil
}

func VerifyToOntTx(native *native.NativeService, proof []byte, fromChainid uint64, height uint32) (*ToMerkleValue, error) {
	//get block header
	header, err := header_sync.GetHeaderByHeight(native, fromChainid, height)
	if err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, get header by height %d from chain %d error: %v",
			height, fromChainid, err)
	}

	v := merkle.MerkleProve(proof, header.CrossStatesRoot)
	if v == nil {
		return nil, fmt.Errorf("VerifyToOntTx, merkle.MerkleProve verify merkle proof error")
	}

	s := common.NewZeroCopySource(v)
	merkleValue := new(ToMerkleValue)
	if err := merkleValue.Deserialization(s); err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, deserialize merkleValue error:%s", err)
	}

	//record done cross chain tx
	err = checkDoneTx(native, merkleValue.TxHash, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, checkDoneTx error:%s", err)
	}
	err = putDoneTx(native, merkleValue.TxHash, fromChainid)
	if err != nil {
		return nil, fmt.Errorf("VerifyToOntTx, putDoneTx error:%s", err)
	}

	notifyVerifyToOntProof(native, merkleValue.TxHash.ToHexString(), merkleValue.MakeTxParam.TxHash, fromChainid)
	return merkleValue, nil
}

func notifyMakeFromOntProof(native *native.NativeService, txHash string, toChainID uint64, key string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.OngContractAddress,
			States:          []interface{}{MAKE_FROM_ONT_PROOF, txHash, toChainID, native.Height, key},
		})
}

func notifyVerifyToOntProof(native *native.NativeService, txHash, rawTxHash string, fromChainID uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.OngContractAddress,
			States:          []interface{}{VERIFY_TO_ONT_PROOF, txHash, rawTxHash, fromChainID, native.Height},
		})
}
