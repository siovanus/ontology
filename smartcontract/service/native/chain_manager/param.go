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
	"math"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type RegisterSideChainParam struct {
	Address            common.Address
	Ratio              uint32
	Deposit            uint64
	OngPool            uint64
	GenesisBlockHeader []byte
	Caller             []byte
	KeyNo              uint32
}

func (this *RegisterSideChainParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Address); err != nil {
		return fmt.Errorf("utils.WriteAddress, serialize address error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Ratio)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize ratio error: %v", err)
	}
	if err := utils.WriteVarUint(w, this.Deposit); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize deposit error: %v", err)
	}
	if err := utils.WriteVarUint(w, this.OngPool); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize ongPool error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Caller); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize caller error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.GenesisBlockHeader); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize genesisBlockHeader error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.KeyNo)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize keyNo error: %v", err)
	}
	return nil
}

func (this *RegisterSideChainParam) Deserialize(r io.Reader) error {
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	ratio, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize ratio error: %v", err)
	}
	if ratio > math.MaxUint32 {
		return fmt.Errorf("ratio larger than max of uint32")
	}
	deposit, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize deposit error: %v", err)
	}
	ongPool, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize ongPool error: %v", err)
	}
	genesisBlockHeader, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes, deserialize genesisBlockHeader error: %v", err)
	}
	caller, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes, deserialize caller error: %v", err)
	}
	keyNo, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize keyNo error: %v", err)
	}
	if keyNo > math.MaxUint32 {
		return fmt.Errorf("initPos larger than max of uint32")
	}
	this.Address = address
	this.Ratio = uint32(ratio)
	this.Deposit = deposit
	this.OngPool = ongPool
	this.GenesisBlockHeader = genesisBlockHeader
	this.Caller = caller
	this.KeyNo = uint32(keyNo)
	return nil
}

type SideChainIDParam struct {
	SideChainID uint32
}

func (this *SideChainIDParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.SideChainID)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize sideChainID error: %v", err)
	}
	return nil
}

func (this *SideChainIDParam) Deserialize(r io.Reader) error {
	sideChainID, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize sideChainID error: %v", err)
	}
	this.SideChainID = uint32(sideChainID)
	return nil
}

type QuitSideChainParam struct {
	SideChainID uint32
	Address     common.Address
}

func (this *QuitSideChainParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.SideChainID)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize sideChainID error: %v", err)
	}
	if err := utils.WriteAddress(w, this.Address); err != nil {
		return fmt.Errorf("utils.WriteVarBytes, serialize address error: %v", err)
	}
	return nil
}

func (this *QuitSideChainParam) Deserialize(r io.Reader) error {
	sideChainID, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize sideChainID error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.SideChainID = uint32(sideChainID)
	this.Address = address
	return nil
}

type InflationParam struct {
	SideChainID uint32
	Address     common.Address
	DepositAdd  uint64
	OngPoolAdd  uint64
}

func (this *InflationParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.SideChainID)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize sideChainID error: %v", err)
	}
	if err := utils.WriteAddress(w, this.Address); err != nil {
		return fmt.Errorf("utils.WriteAddress, serialize address error: %v", err)
	}
	if err := utils.WriteVarUint(w, this.DepositAdd); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize deposit error: %v", err)
	}
	if err := utils.WriteVarUint(w, this.OngPoolAdd); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize ongPool error: %v", err)
	}
	return nil
}

func (this *InflationParam) Deserialize(r io.Reader) error {
	sideChainID, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize sideChainID error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	depositAdd, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize depositAdd error: %v", err)
	}
	ongPoolAdd, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize ongPoolAdd error: %v", err)
	}
	this.SideChainID = uint32(sideChainID)
	this.Address = address
	this.DepositAdd = depositAdd
	this.OngPoolAdd = ongPoolAdd
	return nil
}

type NodeToSideChainParams struct {
	PeerPubkey  string
	Address     common.Address
	SideChainID uint32
}

func (this *NodeToSideChainParams) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := utils.WriteAddress(w, this.Address); err != nil {
		return fmt.Errorf("utils.WriteAddress, serialize address error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.SideChainID)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize sideChainID error: %v", err)
	}
	return nil
}

func (this *NodeToSideChainParams) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	sideChainID, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize sideChainID error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.SideChainID = uint32(sideChainID)
	return nil
}

type BlackSideChainParam struct {
	SideChainID uint32
	Address     common.Address
}

func (this *BlackSideChainParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.SideChainID)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize sideChainID error: %v", err)
	}
	if err := utils.WriteAddress(w, this.Address); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize address error: %v", err)
	}
	return nil
}

func (this *BlackSideChainParam) Deserialize(r io.Reader) error {
	sideChainID, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize sideChainID error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.SideChainID = uint32(sideChainID)
	this.Address = address
	return nil
}

type CommitDposParam struct {
	GovernanceView    *governance.GovernanceView
	PeerPoolMap       *governance.PeerPoolMap
	SideChainNodeInfo *SideChainNodeInfo
	Configuration     *governance.Configuration
	GlobalParam       *governance.GlobalParam
	GlobalParam2      *governance.GlobalParam2
	SplitCurve        *governance.SplitCurve
}

func (this *CommitDposParam) Serialize(w io.Writer) error {
	if err := this.GovernanceView.Serialize(w); err != nil {
		return fmt.Errorf("this.GovernanceView.Serialize, serialize GovernanceView error: %v", err)
	}
	if err := this.PeerPoolMap.Serialize(w); err != nil {
		return fmt.Errorf("this.PeerPoolMap.Serialize, serialize PeerPoolMap error: %v", err)
	}
	if err := this.SideChainNodeInfo.Serialize(w); err != nil {
		return fmt.Errorf("this.SideChainNodeInfo.Serialize, serialize SideChainNodeInfo error: %v", err)
	}
	if err := this.Configuration.Serialize(w); err != nil {
		return fmt.Errorf("this.Configuration.Serialize, serialize Configuration error: %v", err)
	}
	if err := this.GlobalParam.Serialize(w); err != nil {
		return fmt.Errorf("this.GlobalParam.Serialize, serialize GlobalParam error: %v", err)
	}
	if err := this.GlobalParam2.Serialize(w); err != nil {
		return fmt.Errorf("this.GlobalParam2.Serialize, serialize GlobalParam2 error: %v", err)
	}
	if err := this.SplitCurve.Serialize(w); err != nil {
		return fmt.Errorf("this.SplitCurve.Serialize, serialize SplitCurve error: %v", err)
	}
	return nil
}

func (this *CommitDposParam) Deserialize(r io.Reader) error {
	governanceView := new(governance.GovernanceView)
	err := governanceView.Deserialize(r)
	if err != nil {
		return fmt.Errorf("governanceView.Deserialize, deserialize governanceView error: %v", err)
	}
	peerPoolMap := new(governance.PeerPoolMap)
	err = peerPoolMap.Deserialize(r)
	if err != nil {
		return fmt.Errorf("peerPoolMap.Deserialize, deserialize peerPoolMap error: %v", err)
	}
	sideChainNodeInfo := new(SideChainNodeInfo)
	err = sideChainNodeInfo.Deserialize(r)
	if err != nil {
		return fmt.Errorf("sideChainNodeInfo.Deserialize, deserialize sideChainNodeInfo error: %v", err)
	}
	configuration := new(governance.Configuration)
	err = configuration.Deserialize(r)
	if err != nil {
		return fmt.Errorf("configuration.Deserialize, deserialize configuration error: %v", err)
	}
	globalParam := new(governance.GlobalParam)
	err = globalParam.Deserialize(r)
	if err != nil {
		return fmt.Errorf("globalParam.Deserialize, deserialize globalParam error: %v", err)
	}
	globalParam2 := new(governance.GlobalParam2)
	err = globalParam2.Deserialize(r)
	if err != nil {
		return fmt.Errorf("globalParam2.Deserialize, deserialize globalParam2 error: %v", err)
	}
	splitCurve := new(governance.SplitCurve)
	err = splitCurve.Deserialize(r)
	if err != nil {
		return fmt.Errorf("splitCurve.Deserialize, deserialize splitCurve error: %v", err)
	}
	this.GovernanceView = governanceView
	this.PeerPoolMap = peerPoolMap
	this.SideChainNodeInfo = sideChainNodeInfo
	this.Configuration = configuration
	this.GlobalParam = globalParam
	this.GlobalParam2 = globalParam2
	this.SplitCurve = splitCurve
	return nil
}
