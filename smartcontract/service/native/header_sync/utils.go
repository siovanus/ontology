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

package header_sync

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/signature"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func PutBlockHeader(native *native.NativeService, blockHeader *types.Header) error {
	contract := utils.HeaderSyncContractAddress
	sink := common.NewZeroCopySink(nil)
	blockHeader.Serialization(sink)
	sideChainIDBytes, err := utils.GetUint32Bytes(blockHeader.SideChainID)
	if err != nil {
		return fmt.Errorf("sideChainIDBytes, getUint32Bytes error: %v", err)
	}
	heightBytes, err := utils.GetUint32Bytes(blockHeader.Height)
	if err != nil {
		return fmt.Errorf("heightBytes, getUint32Bytes error: %v", err)
	}
	blockHash := blockHeader.Hash()
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(BLOCK_HEADER), sideChainIDBytes, blockHash.ToArray()),
		cstates.GenRawStorageItem(sink.Bytes()))
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(HEADER_INDEX), sideChainIDBytes, heightBytes),
		cstates.GenRawStorageItem(blockHash.ToArray()))
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CURRENT_HEIGHT), sideChainIDBytes), cstates.GenRawStorageItem(heightBytes))
	return nil
}

func GetHeaderByHeight(native *native.NativeService, sideChainID uint32, height uint32) (*types.Header, error) {
	contract := utils.HeaderSyncContractAddress
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, getUint32Bytes error: %v", err)
	}
	heightBytes, err := utils.GetUint32Bytes(height)
	if err != nil {
		return nil, fmt.Errorf("heightBytes, getUint32Bytes error: %v", err)
	}
	blockHashBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(HEADER_INDEX), sideChainIDBytes, heightBytes))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, get headerBytes error: %v", err)
	}
	blockHashStore, err := cstates.GetValueFromRawStorageItem(blockHashBytes)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize from raw storage item err:%v", err)
	}
	header := new(types.Header)
	headerBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(BLOCK_HEADER), sideChainIDBytes, blockHashStore))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, get headerBytes error: %v", err)
	}
	if headerBytes == nil {
		return nil, fmt.Errorf("GetHeaderByHash, can not find any records")
	}
	headerStore, err := cstates.GetValueFromRawStorageItem(headerBytes)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize from raw storage item err:%v", err)
	}
	if err := header.Deserialize(bytes.NewBuffer(headerStore)); err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize header error: %v", err)
	}
	return header, nil
}

func GetHeaderByHash(native *native.NativeService, sideChainID uint32, hash common.Uint256) (*types.Header, error) {
	contract := utils.HeaderSyncContractAddress
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, getUint32Bytes error: %v", err)
	}
	header := new(types.Header)
	headerBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(BLOCK_HEADER), sideChainIDBytes, hash.ToArray()))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, get headerBytes error: %v", err)
	}
	if headerBytes == nil {
		return nil, fmt.Errorf("GetHeaderByHash, can not find any records")
	}
	headerStore, err := cstates.GetValueFromRawStorageItem(headerBytes)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize from raw storage item err:%v", err)
	}
	if err := header.Deserialize(bytes.NewBuffer(headerStore)); err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize header error: %v", err)
	}
	return header, nil
}

func verifyHeader(native *native.NativeService, header *types.Header) error {
	height := header.Height
	//search consensus peer
	keyHeight, err := findKeyHeight(native, height, header.SideChainID)
	if err != nil {
		return fmt.Errorf("verifyHeader, findKeyHeight error:%v", err)
	}
	consensusPeer, err := getConsensusPeer(native, header.SideChainID, keyHeight)
	if err != nil {
		return fmt.Errorf("verifyHeader, get ConsensusPeer error:%v", err)
	}
	if len(header.Bookkeepers)*3 < len(consensusPeer.PeerMap)*2 {
		return fmt.Errorf("verifyHeader, header Bookkeepers num %d must more than 2/3 consensus node num %d", len(header.Bookkeepers), len(consensusPeer.PeerMap))
	}
	for _, bookkeeper := range header.Bookkeepers {
		pubkey := vconfig.PubkeyID(bookkeeper)
		_, present := consensusPeer.PeerMap[pubkey]
		if !present {
			return fmt.Errorf("verifyHeader, invalid pubkey error:%v", pubkey)
		}
	}
	hash := header.Hash()
	err = signature.VerifyMultiSignature(hash[:], header.Bookkeepers, len(header.Bookkeepers), header.SigData)
	if err != nil {
		return fmt.Errorf("verifyHeader, VerifyMultiSignature error:%s, heigh:%d", err, header.Height)
	}
	return nil
}

func GetKeyHeights(native *native.NativeService, sideChainID uint32) (*KeyHeights, error) {
	contract := utils.HeaderSyncContractAddress
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return nil, fmt.Errorf("sideChainIDBytes, getUint32Bytes error: %v", err)
	}
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_HEIGHTS), sideChainIDBytes))
	if err != nil {
		return nil, fmt.Errorf("GetKeyHeights, get keyHeights value error: %v", err)
	}
	keyHeights := &KeyHeights{
		HeightList: make([]uint32, 0),
	}
	if value != nil {
		keyHeightsStore, err := cstates.GetValueFromRawStorageItem(value)
		if err != nil {
			return nil, fmt.Errorf("GetKeyHeights, deserialize from raw storage item err:%v", err)
		}
		err = keyHeights.Deserialization(common.NewZeroCopySource(keyHeightsStore))
		if err != nil {
			return nil, fmt.Errorf("GetKeyHeights, deserialize keyHeights err:%v", err)
		}
	}
	return keyHeights, nil
}

func putKeyHeights(native *native.NativeService, sideChainID uint32, keyHeights *KeyHeights) error {
	contract := utils.HeaderSyncContractAddress
	sink := common.NewZeroCopySink(nil)
	keyHeights.Serialization(sink)
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return fmt.Errorf("getUint32Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_HEIGHTS), sideChainIDBytes), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getConsensusPeer(native *native.NativeService, sideChainID uint32, height uint32) (*ConsensusPeer, error) {
	contract := utils.HeaderSyncContractAddress
	heightBytes, err := utils.GetUint32Bytes(height)
	if err != nil {
		return nil, fmt.Errorf("putConsensusPeer, getUint32Bytes error: %v", err)
	}
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return nil, fmt.Errorf("getUint32Bytes error: %v", err)
	}
	consensusPeerBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(CONSENSUS_PEER), sideChainIDBytes, heightBytes))
	if err != nil {
		return nil, fmt.Errorf("get consensusPeerBytes error: %v", err)
	}
	consensusPeer := new(ConsensusPeer)
	if consensusPeerBytes == nil {
		return nil, fmt.Errorf("getConsensusPeer, can not find any record")
	}
	consensusPeerStore, err := cstates.GetValueFromRawStorageItem(consensusPeerBytes)
	if err != nil {
		return nil, fmt.Errorf("getConsensusPeer, deserialize from raw storage item err:%v", err)
	}
	if err := consensusPeer.Deserialize(bytes.NewBuffer(consensusPeerStore)); err != nil {
		return nil, fmt.Errorf("getConsensusPeer, deserialize consensusPeer error: %v", err)
	}
	return consensusPeer, nil
}

func putConsensusPeer(native *native.NativeService, sideChainID, height uint32, consensusPeer *ConsensusPeer) error {
	contract := utils.HeaderSyncContractAddress
	bf := new(bytes.Buffer)
	if err := consensusPeer.Serialize(bf); err != nil {
		return fmt.Errorf("putConsensusPeer, serialize consensusPeer error: %v", err)
	}
	sideChainIDBytes, err := utils.GetUint32Bytes(sideChainID)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, getUint32Bytes error: %v", err)
	}
	heightBytes, err := utils.GetUint32Bytes(height)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, getUint32Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CONSENSUS_PEER), sideChainIDBytes, heightBytes), cstates.GenRawStorageItem(bf.Bytes()))

	//update key heights
	keyHeights, err := GetKeyHeights(native, sideChainID)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, GetKeyHeights error: %v", err)
	}
	keyHeights.HeightList = append(keyHeights.HeightList, height)
	err = putKeyHeights(native, sideChainID, keyHeights)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, putKeyHeights error: %v", err)
	}
	return nil
}

func UpdateConsensusPeer(native *native.NativeService, header *types.Header) error {
	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(header.ConsensusPayload, blkInfo); err != nil {
		return fmt.Errorf("updateConsensusPeer, unmarshal blockInfo error: %s", err)
	}
	if blkInfo.NewChainConfig != nil {
		consensusPeer := &ConsensusPeer{
			PeerMap: make(map[string]*Peer),
		}
		for _, p := range blkInfo.NewChainConfig.Peers {
			consensusPeer.PeerMap[p.ID] = &Peer{Index: p.Index, PeerPubkey: p.ID}
		}
		err := putConsensusPeer(native, header.SideChainID, header.Height, consensusPeer)
		if err != nil {
			return fmt.Errorf("updateConsensusPeer, put ConsensusPeer eerror: %s", err)
		}
	}
	return nil
}

func findKeyHeight(native *native.NativeService, height uint32, sideChainID uint32) (uint32, error) {
	keyHeights, err := GetKeyHeights(native, sideChainID)
	if err != nil {
		return 0, fmt.Errorf("findKeyHeight, GetKeyHeights error: %v", err)
	}
	for _, v := range keyHeights.HeightList {
		if (height - v) > 0 {
			return v, nil
		}
	}
	return 0, fmt.Errorf("findKeyHeight, can not find key height with height %d", height)
}
