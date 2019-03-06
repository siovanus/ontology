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

package chain_manager

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/auth"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func appCallInitContractAdmin(native *native.NativeService, adminOntID []byte) error {
	bf := new(bytes.Buffer)
	params := &auth.InitContractAdminParam{
		AdminOntID: adminOntID,
	}
	err := params.Serialize(bf)
	if err != nil {
		return fmt.Errorf("appCallInitContractAdmin, param serialize error: %v", err)
	}

	if _, err := native.NativeCall(utils.AuthContractAddress, "initContractAdmin", bf.Bytes()); err != nil {
		return fmt.Errorf("appCallInitContractAdmin, appCall error: %v", err)
	}
	return nil
}

func appCallVerifyToken(native *native.NativeService, contract common.Address, caller []byte, fn string, keyNo uint64) error {
	bf := new(bytes.Buffer)
	params := &auth.VerifyTokenParam{
		ContractAddr: contract,
		Caller:       caller,
		Fn:           fn,
		KeyNo:        keyNo,
	}
	err := params.Serialize(bf)
	if err != nil {
		return fmt.Errorf("appCallVerifyToken, param serialize error: %v", err)
	}

	ok, err := native.NativeCall(utils.AuthContractAddress, "verifyToken", bf.Bytes())
	if err != nil {
		return fmt.Errorf("appCallVerifyToken, appCall error: %v", err)
	}
	if !bytes.Equal(ok.([]byte), utils.BYTE_TRUE) {
		return fmt.Errorf("appCallVerifyToken, verifyToken failed")
	}
	return nil
}

func GetSideChain(native *native.NativeService, contract common.Address, sideChainID uint32) (*SideChain, error) {
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return nil, fmt.Errorf("getUint32Bytes error: %v", err)
	}
	sideChainBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIDE_CHAIN), sideChainIDBytes))
	if err != nil {
		return nil, fmt.Errorf("get sideChainBytes error: %v", err)
	}
	sideChain := new(SideChain)
	if sideChainBytes == nil {
		return nil, fmt.Errorf("getSideChain, can not find any record")
	}
	sideChainStore, err := cstates.GetValueFromRawStorageItem(sideChainBytes)
	if err != nil {
		return nil, fmt.Errorf("getSideChain, deserialize from raw storage item err:%v", err)
	}
	if err := sideChain.Deserialize(common.NewZeroCopySource(sideChainStore)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize sideChain error: %v", err)
	}
	return sideChain, nil
}

func putSideChain(native *native.NativeService, contract common.Address, sideChain *SideChain) error {
	sink := common.NewZeroCopySink(nil)
	sideChain.Serialize(sink)
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChain.SideChainID)
	if err != nil {
		return fmt.Errorf("getUint32Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIDE_CHAIN), sideChainIDBytes),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func deleteSideChain(native *native.NativeService, contract common.Address, sideChainID uint32) error {
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return fmt.Errorf("getUint32Bytes error: %v", err)
	}
	native.CacheDB.Delete(utils.ConcatKey(contract, []byte(SIDE_CHAIN), sideChainIDBytes))
	return nil
}

func getInflationInfo(native *native.NativeService, contract common.Address, sideChainID uint32) (*InflationParam, error) {
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return nil, fmt.Errorf("getUint32Bytes error: %v", err)
	}
	inflationInfoBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(INFLATION_INFO), sideChainIDBytes))
	if err != nil {
		return nil, fmt.Errorf("get inflationInfoBytes error: %v", err)
	}
	inflationInfo := new(InflationParam)
	if inflationInfoBytes == nil {
		return nil, fmt.Errorf("getInflationInfo, can not find any record")
	}
	inflationInfoStore, err := cstates.GetValueFromRawStorageItem(inflationInfoBytes)
	if err != nil {
		return nil, fmt.Errorf("getInflationInfo, deserialize from raw storage item err:%v", err)
	}
	if err := inflationInfo.Deserialize(bytes.NewBuffer(inflationInfoStore)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize inflationInfo error: %v", err)
	}
	return inflationInfo, nil
}

func putInflationInfo(native *native.NativeService, contract common.Address, inflationInfo *InflationParam) error {
	bf := new(bytes.Buffer)
	if err := inflationInfo.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize inflationInfo error: %v", err)
	}
	sideChainIDBytes, err := utils.GetUint32Bytes(inflationInfo.SideChainID)
	if err != nil {
		return fmt.Errorf("getUint32Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(INFLATION_INFO), sideChainIDBytes),
		cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

func getSideChainNodeInfo(native *native.NativeService, contract common.Address, sideChainID uint32) (*SideChainNodeInfo, error) {
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return nil, fmt.Errorf("getUint32Bytes error: %v", err)
	}
	sideChainNodeInfoBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIDE_CHAIN_NODE_INFO), sideChainIDBytes))
	if err != nil {
		return nil, fmt.Errorf("get sideChainNodeInfoBytes error: %v", err)
	}
	sideChainNodeInfo := &SideChainNodeInfo{
		SideChainID: sideChainID,
		NodeInfoMap: make(map[string]*NodeToSideChainParams),
	}
	if sideChainNodeInfoBytes != nil {
		sideChainNodeInfoStore, err := cstates.GetValueFromRawStorageItem(sideChainNodeInfoBytes)
		if err != nil {
			return nil, fmt.Errorf("getSideChainNodeInfo, deserialize from raw storage item err:%v", err)
		}
		if err := sideChainNodeInfo.Deserialize(bytes.NewBuffer(sideChainNodeInfoStore)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize sideChainNodeInfo error: %v", err)
		}
	}
	return sideChainNodeInfo, nil
}

func putSideChainNodeInfo(native *native.NativeService, contract common.Address, sideChainNodeInfo *SideChainNodeInfo) error {
	bf := new(bytes.Buffer)
	if err := sideChainNodeInfo.Serialize(bf); err != nil {
		return fmt.Errorf("serialize, serialize sideChainNodeInfo error: %v", err)
	}
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainNodeInfo.SideChainID)
	if err != nil {
		return fmt.Errorf("getUint32Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIDE_CHAIN_NODE_INFO), sideChainIDBytes),
		cstates.GenRawStorageItem(bf.Bytes()))
	return nil
}

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
