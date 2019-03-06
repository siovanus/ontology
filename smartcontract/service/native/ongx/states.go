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

package ongx

import (
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type OngxLockParam struct {
	Addr  common.Address
	Value uint64
}

func (this *OngxLockParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Addr)
	utils.EncodeVarUint(sink, this.Value)
}

func (this *OngxLockParam) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Addr, err = utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("swap deserialize to error:%s", err)
	}
	this.Value, err = utils.DecodeVarUint(source)
	if err != nil {
		fmt.Errorf("swap deserialize value error:%s", err)
	}
	return nil
}

type OngxUnlockParam struct {
	Addr  common.Address
	Value uint64
	Proof [][]byte
}

func (this *OngxUnlockParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Addr)
	utils.EncodeVarUint(sink, this.Value)
	utils.EncodeVarUint(sink, uint64(len(this.Proof)))
	for _, v := range this.Proof {
		sink.WriteVarBytes(v)
	}
}

func (this *OngxUnlockParam) Deserialization(source *common.ZeroCopySource) error {
	addr, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("OngUnlockParam deserialize addr error:%s", err)
	}
	value, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngUnlockParam deserialize value error:%s", err)
	}
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("OngUnlockParam deserialize proof count error:%s", err)
	}
	var proof [][]byte
	for i := 0; uint64(i) < n; i++ {
		v, _, irregular, eof := source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		proof = append(proof, v)
	}
	this.Addr = addr
	this.Value = value
	this.Proof = proof
	return nil
}
