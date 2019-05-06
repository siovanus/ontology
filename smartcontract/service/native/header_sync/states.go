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
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"sort"
)

type Peer struct {
	Index      uint32
	PeerPubkey string
}

func (this *Peer) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.Index); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize index error: %v", err)
	}
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	return nil
}

func (this *Peer) Deserialize(r io.Reader) error {
	index, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize index error: %v", err)
	}
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	this.Index = index
	this.PeerPubkey = peerPubkey
	return nil
}

type KeyHeights struct {
	HeightList []uint32
}

func (this *KeyHeights) Serialization(sink *common.ZeroCopySink) {
	//first sort the list  (small -> big)
	sort.SliceStable(this.HeightList, func(i, j int) bool {
		return this.HeightList[i] > this.HeightList[j]
	})
	utils.EncodeVarUint(sink, uint64(len(this.HeightList)))
	for _, v := range this.HeightList {
		utils.EncodeVarUint(sink, uint64(v))
	}
}

func (this *KeyHeights) Deserialization(source *common.ZeroCopySource) error {
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize HeightList length error: %v", err)
	}
	heightList := make([]uint32, 0)
	for i := 0; uint64(i) < n; i++ {
		height, err := utils.DecodeVarUint(source)
		if err != nil {
			return fmt.Errorf("utils.DecodeVarUint, deserialize height error: %v", err)
		}
		heightList = append(heightList, uint32(height))
	}
	this.HeightList = heightList
	return nil
}

type ConsensusPeers struct {
	PeerMap map[string]*Peer
}

func (this *ConsensusPeers) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.PeerMap))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize PeerMap length error: %v", err)
	}
	var peerList []*Peer
	for _, v := range this.PeerMap {
		peerList = append(peerList, v)
	}
	sort.SliceStable(peerList, func(i, j int) bool {
		return peerList[i].PeerPubkey > peerList[j].PeerPubkey
	})
	for _, v := range peerList {
		if err := v.Serialize(w); err != nil {
			return fmt.Errorf("serialize peer error: %v", err)
		}
	}
	return nil
}

func (this *ConsensusPeers) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize PeerMap length error: %v", err)
	}
	peerMap := make(map[string]*Peer)
	for i := 0; uint32(i) < n; i++ {
		peer := new(Peer)
		if err := peer.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize peer error: %v", err)
		}
		peerMap[peer.PeerPubkey] = peer
	}
	this.PeerMap = peerMap
	return nil
}
