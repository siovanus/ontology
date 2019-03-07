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

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/smartcontract/service/native/header_sync"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	//status
	RegisterSideChainStatus Status = iota
	SideChainStatus
	QuitingStatus
)

const (
	//function name
	INIT_CONFIG                 = "initConfig"
	REGISTER_SIDE_CHAIN         = "registerSideChain"
	APPROVE_SIDE_CHAIN          = "approveSideChain"
	REJECT_SIDE_CHAIN           = "rejectSideChain"
	QUIT_SIDE_CHAIN             = "quitSideChain"
	APPROVE_QUIT_SIDE_CHAIN     = "approveQuitSideChain"
	BLACK_SIDE_CHAIN            = "blackSideChain"
	INFLATION                   = "inflation"
	APPROVE_INFLATION           = "approveInflation"
	REJECT_INFLATION            = "rejectInflation"
	REGISTER_NODE_TO_SIDE_CHAIN = "registerNodeToSideChain"
	QUIT_NODE_TO_SIDE_CHAIN     = "quitNodeToSideChain"

	//key prefix
	SIDE_CHAIN           = "sideChain"
	INFLATION_INFO       = "inflationInfo"
	SIDE_CHAIN_NODE_INFO = "sideChainNodeInfo"
)

//Init governance contract address
func InitChainManager() {
	native.Contracts[utils.ChainManagerContractAddress] = RegisterChainManagerContract
}

//Register methods of governance contract
func RegisterChainManagerContract(native *native.NativeService) {
	native.Register(INIT_CONFIG, InitConfig)
	native.Register(REGISTER_SIDE_CHAIN, RegisterSideChain)
	native.Register(APPROVE_SIDE_CHAIN, ApproveSideChain)
	native.Register(REJECT_SIDE_CHAIN, RejectSideChain)
	native.Register(QUIT_SIDE_CHAIN, QuitSideChain)
	native.Register(APPROVE_QUIT_SIDE_CHAIN, ApproveQuitSideChain)
	native.Register(BLACK_SIDE_CHAIN, BlackSideChain)
	native.Register(INFLATION, Inflation)
	native.Register(APPROVE_INFLATION, ApproveInflation)
	native.Register(REJECT_INFLATION, RejectInflation)
	native.Register(REGISTER_NODE_TO_SIDE_CHAIN, RegisterNodeToSideChain)
	native.Register(QUIT_NODE_TO_SIDE_CHAIN, QuitNodeToSideChain)
}

func InitConfig(native *native.NativeService) ([]byte, error) {
	configuration := new(config.VBFTConfig)
	buf, err := serialization.ReadVarBytes(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitConfig, serialization.ReadVarBytes error: %v", err)
	}
	if err := configuration.Deserialize(bytes.NewBuffer(buf)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitConfig, configuration.Deserialize error: %v", err)
	}

	//init admin OntID
	err = appCallInitContractAdmin(native, []byte(configuration.AdminOntID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("InitConfig, appCallInitContractAdmin error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func RegisterSideChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSideChainParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, contract params deserialize error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check auth of OntID
	err := appCallVerifyToken(native, contract, params.Caller, REGISTER_SIDE_CHAIN, uint64(params.KeyNo))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, verifyToken failed: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, checkWitness error: %v", err)
	}

	//check if side chain id is correct
	header, err := types.HeaderFromRawBytes(params.GenesisBlockHeader)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, deserialize header err: %v", err)
	}

	//check if side chain exist
	sideChainIDBytes, err := utils.GetUint32Bytes(header.SideChainID)
	if err != nil {
		return nil, fmt.Errorf("RegisterSideChain, getUint32Bytes error: %v", err)
	}
	sideChainBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIDE_CHAIN), sideChainIDBytes))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, get sideChainBytes error: %v", err)
	}
	if sideChainBytes != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, side chain is already registered")
	}

	//side chain storage
	sideChain := &SideChain{
		SideChainID:        header.SideChainID,
		Address:            params.Address,
		Ratio:              uint64(params.Ratio),
		Deposit:            uint64(params.Deposit),
		OngNum:             0,
		OngPool:            uint64(params.OngPool),
		Status:             RegisterSideChainStatus,
		GenesisBlockHeader: params.GenesisBlockHeader,
	}
	err = putSideChain(native, contract, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, put sideChain error: %v", err)
	}

	//ong transfer
	err = appCallTransferOng(native, params.Address, utils.ChainManagerContractAddress, uint64(params.Deposit))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, ong transfer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveSideChain(native *native.NativeService) ([]byte, error) {
	params := new(SideChainIDParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//change side chain status
	sideChain, err := GetSideChain(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, get sideChain error: %v", err)
	}
	if sideChain.Status != RegisterSideChainStatus {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, side chain is not register side chain status")
	}
	sideChain.Status = SideChainStatus

	err = putSideChain(native, contract, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, put sideChain error: %v", err)
	}

	header, err := types.HeaderFromRawBytes(sideChain.GenesisBlockHeader)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, deserialize header err: %v", err)
	}
	//block header storage
	err = header_sync.PutBlockHeader(native, header)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, put blockHeader error: %v", err)
	}

	//consensus node pk storage
	err = header_sync.UpdateConsensusPeer(native, header)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveSideChain, update ConsensusPeer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func RejectSideChain(native *native.NativeService) ([]byte, error) {
	params := new(SideChainIDParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//change side chain status
	sideChain, err := GetSideChain(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, get sideChain error: %v", err)
	}
	if sideChain.Status != RegisterSideChainStatus {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, side chain is not register side chain status")
	}
	err = deleteSideChain(native, contract, params.SideChainID)
	if err != nil {
		return nil, fmt.Errorf("RejectSideChain, deleteSideChain error: %v", err)
	}

	//ong transfer
	err = appCallTransferOng(native, utils.ChainManagerContractAddress, sideChain.Address, sideChain.Deposit)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RejectSideChain, ong transfer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func QuitSideChain(native *native.NativeService) ([]byte, error) {
	params := new(QuitSideChainParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, contract params deserialize error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, checkWitness error: %v", err)
	}

	//get side chain
	sideChain, err := GetSideChain(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, get sideChain error: %v", err)
	}
	if sideChain.SideChainID != params.SideChainID {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, side chain is not registered")
	}
	if sideChain.Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, address is not side chain admin")
	}
	if sideChain.Status != SideChainStatus {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, side chain is not side chain status")
	}
	sideChain.Status = QuitingStatus

	err = putSideChain(native, contract, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("QuitSideChain, put sideChain error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveQuitSideChain(native *native.NativeService) ([]byte, error) {
	params := new(SideChainIDParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get side chain
	sideChain, err := GetSideChain(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, get sideChain error: %v", err)
	}
	if sideChain.SideChainID != params.SideChainID {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, side chain is not registered")
	}
	if sideChain.Status != QuitingStatus {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, side chain is not quiting status")
	}
	err = deleteSideChain(native, contract, params.SideChainID)
	if err != nil {
		return nil, fmt.Errorf("ApproveQuitSideChain, deleteSideChain error: %v", err)
	}

	//ong transfer
	err = appCallTransferOng(native, utils.ChainManagerContractAddress, sideChain.Address, sideChain.Deposit)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveQuitSideChain, ong transfer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func BlackSideChain(native *native.NativeService) ([]byte, error) {
	params := new(BlackSideChainParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackSideChain, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackSideChain, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackSideChain, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get side chain
	sideChain, err := GetSideChain(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackSideChain, get sideChain error: %v", err)
	}

	amount := sideChain.OngNum + sideChain.Deposit
	//ong transfer
	err = appCallTransferOng(native, utils.ChainManagerContractAddress, params.Address, amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BlackSideChain, ong transfer error: %v", err)
	}

	err = deleteSideChain(native, contract, params.SideChainID)
	if err != nil {
		return nil, fmt.Errorf("BlackSideChain, deleteSideChain error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func Inflation(native *native.NativeService) ([]byte, error) {
	params := new(InflationParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, validateOwner error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get side chain
	sideChain, err := GetSideChain(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, get sideChain error: %v", err)
	}
	if sideChain.Status != SideChainStatus {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, side chain status is not normal status")
	}
	if sideChain.Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, address is not side chain admin")
	}

	err = putInflationInfo(native, contract, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, put inflationInfo error: %v", err)
	}

	//ong transfer
	err = appCallTransferOng(native, params.Address, utils.ChainManagerContractAddress, params.DepositAdd)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("inflation, ong transfer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveInflation(native *native.NativeService) ([]byte, error) {
	params := new(SideChainIDParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get inflation info
	inflationInfo, err := getInflationInfo(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, get inflationInfo error: %v", err)
	}

	//get side chain
	sideChain, err := GetSideChain(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, get sideChain error: %v", err)
	}

	sideChain.Deposit = sideChain.Deposit + inflationInfo.DepositAdd
	sideChain.OngPool = sideChain.OngPool + inflationInfo.OngPoolAdd
	err = putSideChain(native, contract, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("approveInflation, put sideChain error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func RejectInflation(native *native.NativeService) ([]byte, error) {
	params := new(SideChainIDParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectInflation, contract params deserialize error: %v", err)
	}

	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectInflation, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectInflation, checkWitness error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get inflation info
	inflationInfo, err := getInflationInfo(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectInflation, get inflationInfo error: %v", err)
	}
	sideChainIDBytes, err := utils.GetUint32Bytes(params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectInflation, getUint32Bytes error: %v", err)
	}
	native.CacheDB.Delete(utils.ConcatKey(contract, []byte(INFLATION_INFO), sideChainIDBytes))

	//ong transfer
	err = appCallTransferOng(native, utils.ChainManagerContractAddress, inflationInfo.Address, inflationInfo.DepositAdd)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("rejectInflation, ong transfer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func RegisterNodeToSideChain(native *native.NativeService) ([]byte, error) {
	params := new(NodeToSideChainParams)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("registerNodeToSideChain, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("registerNodeToSideChain, validateOwner error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//check if side chain exist
	_, err = GetSideChain(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("registerNodeToSideChain, get sideChain error: %v", err)
	}

	//get current view
	view, err := governance.GetView(native, utils.GovernanceContractAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("registerNodeToSideChain, get view error: %v", err)
	}

	//get peerPoolMap
	peerPoolMap, err := governance.GetPeerPoolMap(native, utils.GovernanceContractAddress, view)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("registerNodeToSideChain, get peerPoolMap error: %v", err)
	}

	peerPoolItem, ok := peerPoolMap.PeerPoolMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("registerNodeToSideChain, node is not registered in peer pool map")
	}
	if peerPoolItem.Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("registerNodeToSideChain, address is not node owner")
	}

	//get side chain node info
	sideChainNodeInfo, err := getSideChainNodeInfo(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("registerNodeToSideChain, get sideChainNodeInfo error: %v", err)
	}
	sideChainNodeInfo.NodeInfoMap[params.PeerPubkey] = params

	err = putSideChainNodeInfo(native, contract, sideChainNodeInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("registerNodeToSideChain, put sideChainNodeInfo error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func QuitNodeToSideChain(native *native.NativeService) ([]byte, error) {
	params := new(NodeToSideChainParams)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("quitNodeToSideChain, contract params deserialize error: %v", err)
	}

	//check witness
	err := utils.ValidateOwner(native, params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("quitNodeToSideChain, validateOwner error: %v", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress

	//get side chain node info
	sideChainNodeInfo, err := getSideChainNodeInfo(native, contract, params.SideChainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("quitNodeToSideChain, get sideChainNodeInfo error: %v", err)
	}
	_, ok := sideChainNodeInfo.NodeInfoMap[params.PeerPubkey]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("quitNodeToSideChain, node is not registered")
	}
	if sideChainNodeInfo.NodeInfoMap[params.PeerPubkey].Address != params.Address {
		return utils.BYTE_FALSE, fmt.Errorf("quitNodeToSideChain, address is not node owner")
	}
	delete(sideChainNodeInfo.NodeInfoMap, params.PeerPubkey)

	err = putSideChainNodeInfo(native, contract, sideChainNodeInfo)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("quitNodeToSideChain, put sideChainNodeInfo error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}
