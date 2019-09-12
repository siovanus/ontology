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

package types

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/ontio/ontology/common/serialization"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
)

type Header struct {
	Version          uint32
	ShardID          uint64
	ParentHeight     uint32
	PrevBlockHash    common.Uint256
	TransactionsRoot common.Uint256
	CrossStatesRoot  common.Uint256
	BlockRoot        common.Uint256
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	ConsensusPayload []byte
	NextBookkeeper   common.Address

	//Program *program.Program
	Bookkeepers []keypair.PublicKey
	SigData     [][]byte

	hash *common.Uint256
}

func (bd *Header) Serialization(sink *common.ZeroCopySink) error {
	bd.serializationUnsigned(sink)
	sink.WriteVarUint(uint64(len(bd.Bookkeepers)))

	for _, pubkey := range bd.Bookkeepers {
		sink.WriteVarBytes(keypair.SerializePublicKey(pubkey))
	}

	sink.WriteVarUint(uint64(len(bd.SigData)))
	for _, sig := range bd.SigData {
		sink.WriteVarBytes(sig)
	}

	return nil
}

//Serialize the blockheader data without program
func (bd *Header) serializationUnsigned(sink *common.ZeroCopySink) {
	sink.WriteUint32(bd.Version)
	if bd.Version > CURR_HEADER_VERSION {
		panic(fmt.Errorf("invalid header version:%d", bd.Version))
	}
	if bd.Version == VERSION_SUPPORT_SHARD {
		sink.WriteUint64(bd.ShardID)
		sink.WriteUint32(bd.ParentHeight)
		sink.WriteHash(bd.CrossStatesRoot)
	}
	sink.WriteBytes(bd.PrevBlockHash[:])
	sink.WriteBytes(bd.TransactionsRoot[:])
	sink.WriteBytes(bd.BlockRoot[:])
	sink.WriteUint32(bd.Timestamp)
	sink.WriteUint32(bd.Height)
	sink.WriteUint64(bd.ConsensusData)
	sink.WriteVarBytes(bd.ConsensusPayload)
	sink.WriteBytes(bd.NextBookkeeper[:])
}

func (bd *Header) Serialize(w io.Writer) error {
	if err := bd.serializeUnsigned(w); err != nil {
		return err
	}
	if err := serialization.WriteVarUint(w, uint64(len(bd.Bookkeepers))); err != nil {
		return err
	}

	for _, pubkey := range bd.Bookkeepers {
		if err := serialization.WriteVarBytes(w, keypair.SerializePublicKey(pubkey)); err != nil {
			return err
		}
	}

	if err := serialization.WriteVarUint(w, uint64(len(bd.SigData))); err != nil {
		return err
	}
	for _, sig := range bd.SigData {
		if err := serialization.WriteVarBytes(w, sig); err != nil {
			return err
		}
	}
	return nil
}

func (bd *Header) serializeUnsigned(w io.Writer) error {
	if bd.Version > CURR_HEADER_VERSION {
		panic(fmt.Errorf("invalid header %d over max version:%d", bd.Version, CURR_HEADER_VERSION))
	}
	if err := serialization.WriteUint32(w, bd.Version); err != nil {
		return err
	}
	if bd.Version == VERSION_SUPPORT_SHARD {
		if err := serialization.WriteUint64(w, bd.ShardID); err != nil {
			return err
		}
		if err := serialization.WriteUint32(w, bd.ParentHeight); err != nil {
			return err
		}
		if err := serialization.WriteVarBytes(w, bd.CrossStatesRoot[:]); err != nil {
			return err
		}
	}
	if err := serialization.WriteVarBytes(w, bd.PrevBlockHash[:]); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, bd.TransactionsRoot[:]); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, bd.BlockRoot[:]); err != nil {
		return err
	}
	if err := serialization.WriteUint32(w, bd.Timestamp); err != nil {
		return err
	}
	if err := serialization.WriteUint32(w, bd.Height); err != nil {
		return err
	}
	if err := serialization.WriteUint64(w, bd.ConsensusData); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, bd.ConsensusPayload); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, bd.NextBookkeeper[:]); err != nil {
		return err
	}
	return nil
}

func HeaderFromRawBytes(raw []byte) (*Header, error) {
	source := common.NewZeroCopySource(raw)
	header := &Header{}
	err := header.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return header, nil

}
func (bd *Header) Deserialization(source *common.ZeroCopySource) error {
	err := bd.deserializationUnsigned(source)
	if err != nil {
		return err
	}

	n, _, irregular, eof := source.NextVarUint()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}

	for i := 0; i < int(n); i++ {
		buf, _, irregular, eof := source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		pubkey, err := keypair.DeserializePublicKey(buf)
		if err != nil {
			return err
		}
		bd.Bookkeepers = append(bd.Bookkeepers, pubkey)
	}

	m, _, irregular, eof := source.NextVarUint()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}

	for i := 0; i < int(m); i++ {
		sig, _, irregular, eof := source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		bd.SigData = append(bd.SigData, sig)
	}

	return nil
}

func (bd *Header) deserializationUnsigned(source *common.ZeroCopySource) error {
	var irregular, eof bool

	bd.Version, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if bd.Version > CURR_HEADER_VERSION {
		return common.ErrIrregularData
	}
	if bd.Version == VERSION_SUPPORT_SHARD {
		bd.ShardID, eof = source.NextUint64()
		bd.ParentHeight, eof = source.NextUint32()
		bd.CrossStatesRoot, eof = source.NextHash()
	}
	bd.PrevBlockHash, eof = source.NextHash()
	bd.TransactionsRoot, eof = source.NextHash()
	bd.BlockRoot, eof = source.NextHash()
	bd.Timestamp, eof = source.NextUint32()
	bd.Height, eof = source.NextUint32()
	bd.ConsensusData, eof = source.NextUint64()

	bd.ConsensusPayload, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}

	bd.NextBookkeeper, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (bd *Header) Deserialize(w io.Reader) error {
	err := bd.deserializeUnsigned(w)
	if err != nil {
		return err
	}

	n, err := serialization.ReadVarUint(w, 0)
	if err != nil {
		return errors.New("[Header] deserialize bookkeepers length error")
	}

	for i := 0; i < int(n); i++ {
		buf, err := serialization.ReadVarBytes(w)
		if err != nil {
			return errors.New("[Header] deserialize bookkeepers public key error")
		}
		pubkey, err := keypair.DeserializePublicKey(buf)
		if err != nil {
			return err
		}
		bd.Bookkeepers = append(bd.Bookkeepers, pubkey)
	}

	m, err := serialization.ReadVarUint(w, 0)
	if err != nil {
		return errors.New("[Header] deserialize sigData length error")
	}

	for i := 0; i < int(m); i++ {
		sig, err := serialization.ReadVarBytes(w)
		if err != nil {
			return errors.New("[Header] deserialize sigData error")
		}
		bd.SigData = append(bd.SigData, sig)
	}
	return nil
}

func (bd *Header) deserializeUnsigned(w io.Reader) error {
	var err error
	bd.Version, err = serialization.ReadUint32(w)
	if err != nil {
		return errors.New("[Header] read version error")
	}
	if bd.Version > CURR_HEADER_VERSION {
		return fmt.Errorf("[Header] header version %d over max version %d", bd.Version, CURR_HEADER_VERSION)
	}
	if bd.Version == VERSION_SUPPORT_SHARD {
		bd.ShardID, err = serialization.ReadUint64(w)
		if err != nil {
			return errors.New("[Header] read shardID error")
		}
		bd.ParentHeight, err = serialization.ReadUint32(w)
		if err != nil {
			return errors.New("[Header] read parentHeight error")
		}
		bd.CrossStatesRoot, err = serialization.ReadHash(w)
		if err != nil {
			return errors.New("[Header] read crossStatesRoot error")
		}
	}

	bd.PrevBlockHash, err = serialization.ReadHash(w)
	if err != nil {
		return errors.New("[Header] read prevBlockHash error")
	}
	bd.TransactionsRoot, err = serialization.ReadHash(w)
	if err != nil {
		return errors.New("[Header] read transactionsRoot error")
	}
	bd.CrossStatesRoot, err = serialization.ReadHash(w)
	if err != nil {
		return errors.New("[Header] read crossStatesRoot error")
	}
	bd.BlockRoot, err = serialization.ReadHash(w)
	if err != nil {
		return errors.New("[Header] read blockRoot error")
	}
	bd.Timestamp, err = serialization.ReadUint32(w)
	if err != nil {
		return errors.New("[Header] read timestamp error")
	}
	bd.Height, err = serialization.ReadUint32(w)
	if err != nil {
		return errors.New("[Header] read height error")
	}
	bd.ConsensusData, err = serialization.ReadUint64(w)
	if err != nil {
		return errors.New("[Header] read consensusData error")
	}
	bd.ConsensusPayload, err = serialization.ReadVarBytes(w)
	if err != nil {
		return errors.New("[Header] read consensusPayload error")
	}
	bd.NextBookkeeper, err = serialization.ReadAddress(w)
	if err != nil {
		return errors.New("[Header] read nextBookkeeper error")
	}
	return nil
}

func (bd *Header) Hash() common.Uint256 {
	if bd.hash != nil {
		return *bd.hash
	}
	sink := common.NewZeroCopySink(nil)
	bd.serializationUnsigned(sink)
	temp := sha256.Sum256(sink.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))

	bd.hash = &hash
	return hash
}

func (bd *Header) GetMessage() []byte {
	sink := common.NewZeroCopySink(nil)
	bd.serializationUnsigned(sink)
	return sink.Bytes()
}

func (bd *Header) ToArray() []byte {
	sink := common.NewZeroCopySink(nil)
	bd.Serialization(sink)
	return sink.Bytes()
}
