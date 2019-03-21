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
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type CreateCrossChainTxParam struct {
	OngxFee         uint64
	Address         common.Address
	ChainID         uint64
	ContractAddress common.Address
	FunctionName    string
	Args            []byte
}

func (this *CreateCrossChainTxParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.OngxFee)
	utils.EncodeAddress(sink, this.Address)
	utils.EncodeVarUint(sink, this.ChainID)
	utils.EncodeAddress(sink, this.ContractAddress)
	utils.EncodeString(sink, this.FunctionName)
	utils.EncodeVarBytes(sink, this.Args)
}

func (this *CreateCrossChainTxParam) Deserialization(source *common.ZeroCopySource) error {
	ongxFee, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize chainID error:%s", err)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize address error:%s", err)
	}
	chainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize chainID error:%s", err)
	}
	contractAddress, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize contractAddress error:%s", err)
	}
	functionName, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize functionName error:%s", err)
	}
	args, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize args error:%s", err)
	}
	this.OngxFee = ongxFee
	this.Address = address
	this.ChainID = chainID
	this.ContractAddress = contractAddress
	this.FunctionName = functionName
	this.Args = args
	return nil
}

type ProcessCrossChainTxParam struct {
	Address common.Address
	ChainID uint64
	ID      uint64
	Height  uint32
	Proof   [][]byte
	Value   []byte
}

func (this *ProcessCrossChainTxParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Address)
	utils.EncodeVarUint(sink, this.ChainID)
	utils.EncodeVarUint(sink, this.ID)
	utils.EncodeVarUint(sink, uint64(this.Height))
	utils.EncodeVarUint(sink, uint64(len(this.Proof)))
	for _, v := range this.Proof {
		utils.EncodeVarBytes(sink, v)
	}
	utils.EncodeVarBytes(sink, this.Value)
}

func (this *ProcessCrossChainTxParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("CreateCrossChainTxParam deserialize address error:%s", err)
	}
	chainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngUnlockParam deserialize chainID error:%s", err)
	}
	id, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngUnlockParam deserialize id error:%s", err)
	}
	height, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngUnlockParam deserialize height error:%s", err)
	}
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngUnlockParam deserialize proof count error:%s", err)
	}
	var proof [][]byte
	for i := 0; uint64(i) < n; i++ {
		v, err := utils.DecodeVarBytes(source)
		if err != nil {
			return fmt.Errorf("OngUnlockParam deserialize proof error:%s", err)
		}
		proof = append(proof, v)
	}
	value, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("OngUnlockParam deserialize value error:%s", err)
	}
	this.Address = address
	this.ChainID = chainID
	this.ID = id
	this.Height = uint32(height)
	this.Proof = proof
	this.Value = value
	return nil
}
