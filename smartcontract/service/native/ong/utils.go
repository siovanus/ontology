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

package ong

import (
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/chain_manager"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func putSideChain(native *native.NativeService, sideChain *chain_manager.SideChain) error {
	contract := utils.ChainManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	sideChain.Serialize(sink)
	chainIDBytes, err := utils.GetUint64Bytes(sideChain.ChainID)
	if err != nil {
		return fmt.Errorf("getUint64Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(chain_manager.SIDE_CHAIN), chainIDBytes),
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

func notifyOngLock(native *native.NativeService, contract common.Address, chainID uint64,
	address common.Address, ongxAmount uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{ONG_LOCK, chainID, address.ToBase58(), ongxAmount},
		})
}

func notifyOngUnlock(native *native.NativeService, contract common.Address, chainID uint64,
	address common.Address, ongAmount uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{ONG_UNLOCK, chainID, address.ToBase58(), ongAmount},
		})
}
