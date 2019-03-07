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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"
)

type InitGenesisHeaderParam struct {
	GenesisHeader []byte
}

func (this *InitGenesisHeaderParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarBytes(sink, this.GenesisHeader)
	return nil
}

func (this *InitGenesisHeaderParam) Deserialization(source *common.ZeroCopySource) error {
	genesisHeader, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize genesisHeader count error:%s", err)
	}
	this.GenesisHeader = genesisHeader
	return nil
}

type SyncBlockHeaderParam struct {
	Headers [][]byte
}

func (this *SyncBlockHeaderParam) Serialize(w io.Writer) error {
	err := utils.WriteVarUint(w, uint64(len(this.Headers)))
	if err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize header count error:%s", err)
	}
	for _, v := range this.Headers {
		if err := serialization.WriteVarBytes(w, v); err != nil {
			return fmt.Errorf("serialization.WriteVarBytes, serialize header error: %v", err)
		}
	}
	return nil
}

func (this *SyncBlockHeaderParam) Deserialize(r io.Reader) error {
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize header count error:%s", err)
	}
	var headers [][]byte
	for i := 0; uint64(i) < n; i++ {
		header, err := serialization.ReadVarBytes(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadVarBytes, deserialize header error: %v", err)
		}
		headers = append(headers, header)
	}
	this.Headers = headers
	return nil
}
