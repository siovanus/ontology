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

func GetMainChain(native *native.NativeService) (uint64, error) {
	contract := utils.ChainManagerContractAddress
	mainChainStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(MAIN_CHAIN)))
	if err != nil {
		return 0, fmt.Errorf("get mainChainStore error: %v", err)
	}
	if mainChainStore == nil {
		return 0, fmt.Errorf("GetMainChain, can not find any record")
	}
	mainChainBytes, err := cstates.GetValueFromRawStorageItem(mainChainStore)
	if err != nil {
		return 0, fmt.Errorf("GetMainChain, deserialize from raw storage item err:%v", err)
	}
	mainChainID, err := utils.GetBytesUint64(mainChainBytes)
	if err != nil {
		return 0, fmt.Errorf("GetMainChain, utils.GetBytesUint64 err:%v", err)
	}
	return mainChainID, nil
}

func putMainChain(native *native.NativeService, chainID uint64) error {
	contract := utils.ChainManagerContractAddress
	mainChainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("getUint64Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(MAIN_CHAIN)), cstates.GenRawStorageItem(mainChainIDBytes))
	return nil
}

func GetSideChain(native *native.NativeService, chainID uint64) (*SideChain, error) {
	contract := utils.ChainManagerContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return nil, fmt.Errorf("getUint64Bytes error: %v", err)
	}
	sideChainStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIDE_CHAIN), chainIDBytes))
	if err != nil {
		return nil, fmt.Errorf("get sideChainStore error: %v", err)
	}
	sideChain := new(SideChain)
	if sideChainStore == nil {
		return nil, fmt.Errorf("getSideChain, can not find any record")
	}
	sideChainBytes, err := cstates.GetValueFromRawStorageItem(sideChainStore)
	if err != nil {
		return nil, fmt.Errorf("getSideChain, deserialize from raw storage item err:%v", err)
	}
	if err := sideChain.Deserialize(common.NewZeroCopySource(sideChainBytes)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize sideChain error: %v", err)
	}
	return sideChain, nil
}

func putSideChain(native *native.NativeService, sideChain *SideChain) error {
	contract := utils.ChainManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	sideChain.Serialize(sink)
	chainIDBytes, err := utils.GetUint64Bytes(sideChain.ChainID)
	if err != nil {
		return fmt.Errorf("getUint64Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIDE_CHAIN), chainIDBytes),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func deleteSideChain(native *native.NativeService, chainID uint64) error {
	contract := utils.ChainManagerContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("getUint64Bytes error: %v", err)
	}
	native.CacheDB.Delete(utils.ConcatKey(contract, []byte(SIDE_CHAIN), chainIDBytes))
	return nil
}

func getInflationInfo(native *native.NativeService, chainID uint64) (*InflationParam, error) {
	contract := utils.ChainManagerContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return nil, fmt.Errorf("getUint64Bytes error: %v", err)
	}
	inflationInfoStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(INFLATION_INFO), chainIDBytes))
	if err != nil {
		return nil, fmt.Errorf("get inflationInfoStore error: %v", err)
	}
	inflationInfo := new(InflationParam)
	if inflationInfoStore == nil {
		return nil, fmt.Errorf("getInflationInfo, can not find any record")
	}
	inflationInfoBytes, err := cstates.GetValueFromRawStorageItem(inflationInfoStore)
	if err != nil {
		return nil, fmt.Errorf("getInflationInfo, deserialize from raw storage item err:%v", err)
	}
	if err := inflationInfo.Deserialization(common.NewZeroCopySource(inflationInfoBytes)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize inflationInfo error: %v", err)
	}
	return inflationInfo, nil
}

func putInflationInfo(native *native.NativeService, inflationInfo *InflationParam) error {
	contract := utils.ChainManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	if err := inflationInfo.Serialization(sink); err != nil {
		return fmt.Errorf("serialize, serialize inflationInfo error: %v", err)
	}
	chainIDBytes, err := utils.GetUint64Bytes(inflationInfo.ChainID)
	if err != nil {
		return fmt.Errorf("getUint64Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(INFLATION_INFO), chainIDBytes),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getSideChainNodeInfo(native *native.NativeService, chainID uint64) (*SideChainNodeInfo, error) {
	contract := utils.ChainManagerContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return nil, fmt.Errorf("getUint64Bytes error: %v", err)
	}
	sideChainNodeInfoStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIDE_CHAIN_NODE_INFO), chainIDBytes))
	if err != nil {
		return nil, fmt.Errorf("get sideChainNodeInfoStore error: %v", err)
	}
	sideChainNodeInfo := &SideChainNodeInfo{
		ChainID:     chainID,
		NodeInfoMap: make(map[string]*NodeToSideChainParams),
	}
	if sideChainNodeInfoStore != nil {
		sideChainNodeInfoBytes, err := cstates.GetValueFromRawStorageItem(sideChainNodeInfoStore)
		if err != nil {
			return nil, fmt.Errorf("getSideChainNodeInfo, deserialize from raw storage item err:%v", err)
		}
		if err := sideChainNodeInfo.Deserialization(common.NewZeroCopySource(sideChainNodeInfoBytes)); err != nil {
			return nil, fmt.Errorf("deserialize, deserialize sideChainNodeInfo error: %v", err)
		}
	}
	return sideChainNodeInfo, nil
}

func putSideChainNodeInfo(native *native.NativeService, sideChainNodeInfo *SideChainNodeInfo) error {
	contract := utils.ChainManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	if err := sideChainNodeInfo.Serialization(sink); err != nil {
		return fmt.Errorf("serialize, serialize sideChainNodeInfo error: %v", err)
	}
	chainIDBytes, err := utils.GetUint64Bytes(sideChainNodeInfo.ChainID)
	if err != nil {
		return fmt.Errorf("getUint64Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIDE_CHAIN_NODE_INFO), chainIDBytes),
		cstates.GenRawStorageItem(sink.Bytes()))
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
