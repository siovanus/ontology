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
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"sort"
)

type Status uint8

func (this *Status) Serialize(sink *common.ZeroCopySink) {
	sink.WriteUint8(uint8(*this))
}

func (this *Status) Deserialize(source *common.ZeroCopySource) error {
	status, eof := source.NextUint8()
	if eof {
		return io.ErrUnexpectedEOF
	}
	*this = Status(status)
	return nil
}

type SideChain struct {
	SideChainID  uint32         //side chain id
	Address      common.Address //side chain admin
	Ratio        uint64         //side chain ong ratio(ong:ongx)
	Deposit      uint64         //side chain deposit
	OngNum       uint64         //side chain ong num
	OngPool      uint64         //side chain ong pool limit
	Status       Status         //side chain status
	GenesisBlock []byte         //side chain genesis block
}

func (this *SideChain) Serialize(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.SideChainID)
	sink.WriteAddress(this.Address)
	sink.WriteUint64(this.Ratio)
	sink.WriteUint64(this.Deposit)
	sink.WriteUint64(this.OngNum)
	sink.WriteUint64(this.OngPool)
	this.Status.Serialize(sink)
	sink.WriteVarBytes(this.GenesisBlock)
}

func (this *SideChain) Deserialize(source *common.ZeroCopySource) error {
	sideChainID, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextString, deserialize sideChainID error: %v", io.ErrUnexpectedEOF)
	}
	address, eof := source.NextAddress()
	if eof {
		return fmt.Errorf("source.NextAddress, deserialize address error: %v", io.ErrUnexpectedEOF)
	}
	ratio, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize ratio error: %v", io.ErrUnexpectedEOF)
	}
	deposit, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize deposit error: %v", io.ErrUnexpectedEOF)
	}
	ongNum, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize ongNum error: %v", io.ErrUnexpectedEOF)
	}
	ongPool, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize ongPool error: %v", io.ErrUnexpectedEOF)
	}
	status := new(Status)
	err := status.Deserialize(source)
	if err != nil {
		return fmt.Errorf("status.Deserialize. deserialize status error: %v", err)
	}
	genesisBlock, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return fmt.Errorf("source.NextVarBytes, deserialize genesisBlock error: %v", common.ErrIrregularData)
	}
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize genesisBlock error: %v", io.ErrUnexpectedEOF)
	}
	this.SideChainID = sideChainID
	this.Address = address
	this.Ratio = ratio
	this.Deposit = deposit
	this.OngNum = ongNum
	this.OngPool = ongPool
	this.Status = *status
	this.GenesisBlock = genesisBlock
	return nil
}

type SideChainNodeInfo struct {
	SideChainID uint32
	NodeInfoMap map[string]*NodeToSideChainParams
}

func (this *SideChainNodeInfo) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.SideChainID); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize sideChainID error: %v", err)
	}
	if err := serialization.WriteUint32(w, uint32(len(this.NodeInfoMap))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize PeerPoolMap length error: %v", err)
	}
	var nodeInfoMapList []*NodeToSideChainParams
	for _, v := range this.NodeInfoMap {
		nodeInfoMapList = append(nodeInfoMapList, v)
	}
	sort.SliceStable(nodeInfoMapList, func(i, j int) bool {
		return nodeInfoMapList[i].PeerPubkey > nodeInfoMapList[j].PeerPubkey
	})
	for _, v := range nodeInfoMapList {
		if err := v.Serialize(w); err != nil {
			return fmt.Errorf("serialize peerPool error: %v", err)
		}
	}
	return nil
}

func (this *SideChainNodeInfo) Deserialize(r io.Reader) error {
	sideChainID, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize sideChainID error: %v", err)
	}
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize PeerPoolMap length error: %v", err)
	}
	nodeInfoMap := make(map[string]*NodeToSideChainParams)
	for i := 0; uint32(i) < n; i++ {
		nodeInfo := new(NodeToSideChainParams)
		if err := nodeInfo.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize peerPool error: %v", err)
		}
		nodeInfoMap[nodeInfo.PeerPubkey] = nodeInfo
	}
	this.SideChainID = sideChainID
	this.NodeInfoMap = nodeInfoMap
	return nil
}
