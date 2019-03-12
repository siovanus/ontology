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

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	alog "github.com/ontio/ontology-eventbus/log"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/consensus/vbft"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events"
	"github.com/urfave/cli"
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Ontology CLI"
	app.Action = startOntology
	app.Version = config.Version
	app.Copyright = "Copyright in 2018 The Ontology Authors"
	app.Commands = []cli.Command{
		cmd.AccountCommand,
		cmd.InfoCommand,
		cmd.AssetCommand,
		cmd.ContractCommand,
		cmd.ImportCommand,
		cmd.ExportCommand,
		cmd.TxCommond,
		cmd.SigTxCommand,
		cmd.MultiSigAddrCommand,
		cmd.MultiSigTxCommand,
		cmd.SendTxCommand,
		cmd.ShowTxCommand,
	}
	app.Flags = []cli.Flag{
		//common setting
		utils.ConfigFlag,
		utils.LogLevelFlag,
		utils.DisableEventLogFlag,
		utils.DataDirFlag,
		//account setting
		utils.WalletFileFlag,
		utils.AccountAddressFlag,
		utils.AccountPassFlag,
		//consensus setting
		utils.EnableConsensusFlag,
		utils.MaxTxInBlockFlag,
		//txpool setting
		utils.GasPriceFlag,
		utils.GasLimitFlag,
		utils.TxpoolPreExecDisableFlag,
		utils.DisableSyncVerifyTxFlag,
		utils.DisableBroadcastNetTxFlag,
		//p2p setting
		utils.ReservedPeersOnlyFlag,
		utils.ReservedPeersFileFlag,
		utils.NetworkIdFlag,
		utils.NodePortFlag,
		utils.ConsensusPortFlag,
		utils.DualPortSupportFlag,
		utils.MaxConnInBoundFlag,
		utils.MaxConnOutBoundFlag,
		utils.MaxConnInBoundForSingleIPFlag,
		//test mode setting
		utils.EnableTestModeFlag,
		utils.TestModeGenBlockTimeFlag,
		//rpc setting
		utils.RPCDisabledFlag,
		utils.RPCPortFlag,
		utils.RPCLocalEnableFlag,
		utils.RPCLocalProtFlag,
		//rest setting
		utils.RestfulEnableFlag,
		utils.RestfulPortFlag,
		utils.RestfulMaxConnsFlag,
		//ws setting
		utils.WsEnabledFlag,
		utils.WsPortFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func main() {
	if err := setupAPP().Run(os.Args); err != nil {
		cmd.PrintErrorMsg(err.Error())
		os.Exit(1)
	}
}

func startOntology(ctx *cli.Context) {
	initLog(ctx)

	log.Infof("ontology version %s", config.Version)

	cfg, err := initConfig(ctx)
	if err != nil {
		log.Errorf("initConfig error:%s", err)
		return
	}
	stateHashHeight := config.GetStateHashCheckHeight(cfg.P2PNode.NetworkId)
	ldg, err := initLedger(ctx, stateHashHeight)
	if err != nil {
		log.Errorf("initLedger error: %s", err)
		return
	}
	defer ldg.Close()
	log.Infof("current block height is :%d", ldg.GetCurrentBlockHeight())
	var singers []*account.Account
	paths := []string{
		"wallet1.dat",
		"wallet2.dat",
		"wallet3.dat",
		"wallet4.dat",
		"wallet5.dat",
		"wallet6.dat",
		"wallet7.dat",
	}
	for _, path := range paths {
		wallet, err := account.Open(path)
		if err != nil {
			log.Errorf("open wallet error:%s", err)
			return
		}
		pwd, err := password.GetPassword()
		if err != nil {
			log.Errorf("getPassword error:%s", err)
			return
		}
		acc, err := wallet.GetDefaultAccount(pwd)
		if err != nil {
			log.Errorf("wallet.GetDefaultAccount error:%s", err)
			return
		}
		singers = append(singers, acc)
	}
	for i := 0; i < 200000; i++ {
		if i % 10000 == 0 {
			log.Infof("current Height is :%d", i)
		}
		currentHeight := ldg.GetCurrentBlockHeight()
		preBlock, err := ldg.GetBlockByHeight(currentHeight)
		if err != nil {
			log.Errorf("ldg.GetBlockByHeight error: %s", err)
			return
		}
		block, err := buildEmptyBlock(preBlock, singers)
		if err != nil {
			log.Errorf("buildEmptyBlock error:%s", err)
			return
		}
		ldg.AddBlock(block, common.UINT256_EMPTY)
	}
	fmt.Println("Done")
}

func buildEmptyBlock(preBlock *types.Block, singers []*account.Account) (*types.Block, error) {
	sysTxs := make([]*types.Transaction, 0)
	consensusPayload, err := getconsensusPayload(preBlock)
	if err != nil {
		return nil, err
	}
	blocktimestamp := uint32(time.Now().Unix())
	if preBlock.Header.Timestamp >= blocktimestamp {
		blocktimestamp = preBlock.Header.Timestamp + 1
	}
	blk, err := constructBlock(singers, preBlock, blocktimestamp, sysTxs, consensusPayload)
	if err != nil {
		return nil, fmt.Errorf("constructBlock failed")
	}
	return blk, nil
}

//func buildCommitDposBlock(preBlock *types.Block, singers []*account.Account) (*types.Block, error) {
//	sysTxs := make([]*types.Transaction, 0)
//	tx, err := createGovernaceTransaction(preBlock.Header.Height + 1)
//	if err != nil {
//		return nil, err
//	}
//	sysTxs = append(sysTxs, tx)
//	consensusPayload, err := getconsensusPayload(preBlock)
//	if err != nil {
//		return nil, err
//	}
//	blocktimestamp := uint32(time.Now().Unix())
//	if preBlock.Header.Timestamp >= blocktimestamp {
//		blocktimestamp = preBlock.Header.Timestamp + 1
//	}
//	blk, err := constructBlock(singers, preBlock, blocktimestamp, sysTxs, consensusPayload)
//	if err != nil {
//		return nil, fmt.Errorf("constructBlock failed")
//	}
//	return blk, nil
//}
//
//func createGovernaceTransaction(blkNum uint32) (*types.Transaction, error) {
//	mutable := cutils.BuildNativeTransaction(nutils.GovernanceContractAddress, governance.COMMIT_DPOS, []byte{})
//	mutable.Nonce = blkNum
//	tx, err := mutable.IntoImmutable()
//	return tx, err
//}

func getconsensusPayload(blk *types.Block) ([]byte, error) {
	block, err := initVbftBlock(blk)
	if err != nil {
		return nil, err
	}
	lastConfigBlkNum := block.Info.LastConfigBlockNum
	if block.Info.NewChainConfig != nil {
		lastConfigBlkNum = block.Block.Header.Height
	}
	vbftBlkInfo := &vconfig.VbftBlockInfo{
		Proposer:           math.MaxUint32,
		LastConfigBlockNum: lastConfigBlkNum,
		NewChainConfig:     nil,
	}
	consensusPayload, err := json.Marshal(vbftBlkInfo)
	if err != nil {
		return nil, err
	}
	return consensusPayload, nil
}

func initVbftBlock(block *types.Block) (*vbft.Block, error) {
	if block == nil {
		return nil, fmt.Errorf("nil block in initVbftBlock")
	}

	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, blkInfo); err != nil {
		return nil, fmt.Errorf("unmarshal blockInfo: %s", err)
	}

	return &vbft.Block{
		Block: block,
		Info:  blkInfo,
	}, nil
}

func constructBlock(singers []*account.Account, preBlock *types.Block, blocktimestamp uint32, systxs []*types.Transaction, consensusPayload []byte) (*types.Block, error) {
	txHash := []common.Uint256{}
	for _, t := range systxs {
		txHash = append(txHash, t.Hash())
	}
	txRoot := common.ComputeMerkleRoot(txHash)
	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoots(preBlock.Header.Height, []common.Uint256{preBlock.Header.TransactionsRoot, txRoot})

	blkHeader := &types.Header{
		PrevBlockHash:    preBlock.Hash(),
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        blocktimestamp,
		Height:           uint32(preBlock.Header.Height + 1),
		ConsensusData:    common.GetNonce(),
		ConsensusPayload: consensusPayload,
	}
	blk := &types.Block{
		Header:       blkHeader,
		Transactions: systxs,
	}

	blkHash := blk.Hash()
	for _, singer := range singers {
		sig, err := signature.Sign(singer, blkHash[:])
		if err != nil {
			return nil, fmt.Errorf("sign block failed, block hash：%x, error: %s", blkHash, err)
		}
		blkHeader.Bookkeepers = append(blkHeader.Bookkeepers, singer.PublicKey)
		blkHeader.SigData = append(blkHeader.SigData, sig)
	}

	return blk, nil
}

func initLog(ctx *cli.Context) {
	//init log module
	logLevel := ctx.GlobalInt(utils.GetFlagName(utils.LogLevelFlag))
	alog.InitLog(log.PATH)
	log.InitLog(logLevel, log.PATH, log.Stdout)
}

func initConfig(ctx *cli.Context) (*config.OntologyConfig, error) {
	//init ontology config from cli
	cfg, err := cmd.SetOntologyConfig(ctx)
	if err != nil {
		return nil, err
	}
	log.Infof("Config init success")
	return cfg, nil
}

func initLedger(ctx *cli.Context, stateHashHeight uint32) (*ledger.Ledger, error) {
	events.Init() //Init event hub

	var err error
	dbDir := utils.GetStoreDirPath(config.DefConfig.Common.DataDir, config.DefConfig.P2PNode.NetworkName)
	ledger.DefLedger, err = ledger.NewLedger(dbDir, stateHashHeight)
	if err != nil {
		return nil, fmt.Errorf("NewLedger error:%s", err)
	}
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return nil, fmt.Errorf("GetBookkeepers error:%s", err)
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		return nil, fmt.Errorf("genesisBlock error %s", err)
	}
	err = ledger.DefLedger.Init(bookKeepers, genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("Init ledger error:%s", err)
	}

	log.Infof("Ledger init success")
	return ledger.DefLedger, nil
}
