package main

import (
	"fmt"
	"time"
)

type Core struct {
	txpool *TxPool
	store  *Store
}

func NewCore(txPool *TxPool, store *Store) *Core {
	return &Core{
		txpool: txPool,
		store:  store,
	}
}
func (core *Core) GenerateBlock() {
	var preHash []byte
	var preHeight uint64
	t := time.NewTicker(time.Second * 1)
	start := time.Now()
	firstBlockStart := time.Now()
	go func() {
		for {
			<-t.C

			//从交易池获取未打包交易
			txs := core.txpool.FetchTxs()
			if len(txs) == 0 {
				fmt.Printf("no tx in pool  cost: %v, TPS: %v\n", time.Since(firstBlockStart),
					float64(TOTAL_TX)/time.Since(firstBlockStart).Seconds())
				return
			}
			//产生新区块
			newBlock := GenerateBlock(preHeight+1, preHash, txs)
			//保存新区块
			core.store.SaveBlock(newBlock)
			//更新变量
			preHeight++
			preHash = newBlock.Header.BlockHash
			fmt.Printf("Generate new block[%d] tx count= %d, cost: %v, TPS: %v\n", newBlock.Header.BlockHeight,
				len(txs), time.Since(start), float64(len(txs))/time.Since(start).Seconds())
			start = time.Now()
		}
	}()
	for {
		time.Sleep(time.Second * 1)
		fmt.Printf("Verified tx count=%d\n", VerifiedTx)
	}
}

//GenerateBlock 根据一堆Txs和当前高度，上一区块Hash，生成新的区块
func GenerateBlock(height uint64, preBlockHash []byte, txs []*Transaction) *Block {
	header := &BlockHeader{
		BlockHeight:    height,
		BlockHash:      nil,
		PreBlockHash:   preBlockHash,
		TxRoot:         CalcTxRoot(txs),
		BlockTimestamp: time.Now().Unix(),
		Proposer:       []byte{1},
		Signature:      nil,
	}
	headerBytes, _ := header.Marshal()
	header.Signature, _ = SignData(headerBytes)
	headerBytes, _ = header.Marshal()
	header.BlockHash = Hash(headerBytes)
	return &Block{
		Header: header,
		Txs:    txs,
	}
}
