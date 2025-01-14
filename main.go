package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"runtime"
	"sync"
	"time"
)

var TOTAL_TX = 1000000
var VerifiedTx = 0

func main() {
	testCase2()
}

//func testCase1() {
//	GenerateMemKey()
//
//	txPool := NewTxPool()
//
//	store := NewStore()
//	core := NewCore(txPool, store)
//	//不断产生新交易
//	net := NewNetwork(func(msg []byte) {
//		//网络收到消息后反序列化出Tx，验证签名通过，并放入TxPool
//		tx := &Transaction{}
//		tx.Unmarshal(msg)
//		if err := VerifyTx(tx); err != nil {
//			fmt.Println(err.Error())
//		}
//		txPool.AddTx(tx)
//	})
//	//客户端不断产生新交易并放入网络模块
//	go func() {
//		for i := 0; i < TOTAL_TX; i++ {
//			tx := GenerateTx(i)
//			txMsg, _ := tx.Marshal()
//			net.SendMessage(txMsg)
//		}
//	}()
//	//产块节点核心引擎不断产生新区块
//
//	core.GenerateBlock()
//
//}

//预先产生好所有的Tx并签名，然后以最快速度放入TxPool
func testCase2() {
	GenerateMemKey()

	txPool := NewTxPool()

	store := NewStore()
	core := NewCore(txPool, store)
	//不断产生新交易
	net := NewNetwork(func(msg []byte) {
		//网络收到消息后反序列化出Tx，验证签名通过，并放入TxPool
		tx := &Transaction{}
		tx.Unmarshal(msg)
		if err := VerifyTx(tx); err != nil {
			fmt.Println("verify tx fail:" + err.Error())
			return
		}
		txPool.AddTx(tx)
		VerifiedTx++
	})
	//客户端产生新交易并放入网络模块
	fmt.Println("Prepare tx...")
	start := time.Now()
	txs := GenerateTxs(TOTAL_TX)
	fmt.Printf("Generated %d tx, spend:%v\n", len(txs), time.Since(start))
	go func(txs []*Transaction) {
		for i := 0; i < len(txs); i++ {
			tx := txs[i]
			//tx.Payload = []byte{0xee}
			txMsg, _ := tx.Marshal()
			go net.SendMessage(txMsg)
		}
	}(txs)
	//产块节点核心引擎不断产生新区块

	core.GenerateBlock()

}

func GenerateMemKey() error {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privateKey = priv
	publicKey = &priv.PublicKey
	return nil
}
func GenerateKeyFile() error {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privateKey = priv
	publicKey = &priv.PublicKey
	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv)
	fmt.Printf("generate a new key:%x", privBytes)
	return ioutil.WriteFile("key.key", privBytes, fs.ModePerm)
}

func GenerateTx(i int) *Transaction {
	tx := &Transaction{
		Payload:   Uint32ToBytes(uint32(i)),
		Sender:    []byte{1},
		Signature: nil,
		TxHash:    nil,
	}
	txBytes, _ := tx.Marshal()
	tx.Signature, _ = SignData(txBytes)
	txBytes, _ = tx.Marshal()
	tx.TxHash = Hash(txBytes)
	return tx
}
func GenerateTxs(count int) []*Transaction {

	result := make([]*Transaction, 0)
	wg := sync.WaitGroup{}
	wg.Add(runtime.NumCPU())
	for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
		go func(c int) {
			defer wg.Done()
			for i := 0; i < count/runtime.NumCPU(); i++ {
				tx := GenerateTx(i)
				result = append(result, tx)
			}

		}(cpu)
	}
	wg.Wait()

	return result
}

func VerifyTx(tx *Transaction) error {
	tx2 := &Transaction{
		Payload:   tx.Payload,
		Sender:    tx.Sender,
		Signature: nil,
		TxHash:    nil,
	}
	txBytes, _ := tx2.Marshal()
	if !VerifySignature(txBytes, tx.Signature) {
		return errors.New("verify fail")
	}
	return nil
}

func VerifyTxs(txs []*Transaction) error {
	for _, tx := range txs {
		if err := VerifyTx(tx); err != nil {
			return err
		}
	}
	return nil
}
