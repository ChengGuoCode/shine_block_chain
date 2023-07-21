package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"github.com/syndtr/goleveldb/leveldb"
	"runtime"
	"shineBlockChain/constcoe"
	"shineBlockChain/transaction"
	"shineBlockChain/utils"
)

type BlockChain struct {
	LastHash []byte
	Database *leveldb.DB
}

type ChainIterator struct {
	CurrentHash []byte
	Database    *leveldb.DB
}

func (bc *BlockChain) Iterator() *ChainIterator {
	iterator := ChainIterator{bc.LastHash, bc.Database}
	return &iterator
}

func InitBlockChain(address []byte) *BlockChain {
	db, err := leveldb.OpenFile(constcoe.BCPath, nil)
	utils.HandleErr(err)

	genesis := GenesisBlockAddr(address)

	err = db.Put(genesis.Hash, genesis.Serialize(), nil)
	utils.HandleErr(err)

	err = db.Put([]byte("1h"), genesis.Hash, nil)
	utils.HandleErr(err)

	err = db.Put([]byte("gbHash"), genesis.PrevHash, nil)
	utils.HandleErr(err)

	blockChain := BlockChain{genesis.Hash, db}
	return &blockChain
}

func ContinueBlockChain() *BlockChain {
	db, err := leveldb.OpenFile(constcoe.BCPath, nil)
	utils.HandleErr(err)

	lastHash, err := db.Get([]byte("1h"), nil)
	utils.HandleErr(err)

	blockChain := BlockChain{lastHash, db}
	return &blockChain
}

func (bc *BlockChain) AddBlock(newBlock *Block) {
	lastHash, err := bc.Database.Get([]byte("1h"), nil)
	utils.HandleErr(err)

	if !bytes.Equal(newBlock.PrevHash, lastHash) {
		runtime.Goexit()
	}

	err = bc.Database.Put(newBlock.Hash, newBlock.Serialize(), nil)
	utils.HandleErr(err)

	err = bc.Database.Put([]byte("1h"), newBlock.Hash, nil)
	utils.HandleErr(err)

	bc.LastHash = newBlock.Hash
}

func (iterator *ChainIterator) Next() *Block {
	item, err := iterator.Database.Get(iterator.CurrentHash, nil)
	utils.HandleErr(err)

	block := DeSerializeBlock(item)
	iterator.CurrentHash = block.PrevHash

	return block
}

func (iterator *ChainIterator) HasNext() bool {
	gbHash, err := iterator.Database.Get([]byte("gbHash"), nil)
	utils.HandleErr(err)

	if bytes.Equal(iterator.CurrentHash, gbHash) {
		return false
	}
	return true
}

func (bc *BlockChain) FindUnspentTransactions(address []byte) []transaction.Transaction {
	var unSpentTxs []transaction.Transaction
	spentTxs := make(map[string][]int)

	iterator := bc.Iterator()
	for iterator.HasNext() {
		block := iterator.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		IterOutputs:
			for outIdx, out := range tx.Outputs {
				if spentTxs[txID] != nil {
					for _, spentOut := range spentTxs[txID] {
						if spentOut == outIdx {
							continue IterOutputs
						}
					}
				}

				if out.ToAddressRight(address) {
					unSpentTxs = append(unSpentTxs, *tx)
				}
			}
			if !tx.IsBase() {
				for _, in := range tx.Inputs {
					if in.FromAddressRight(address) {
						inTxID := hex.EncodeToString(in.TxID)
						spentTxs[inTxID] = append(spentTxs[inTxID], in.OutIdx)
					}
				}
			}
		}
	}

	return unSpentTxs
}

func (bc *BlockChain) FindUTXOs(address []byte) (int, map[string]int) {
	unspentOuts := make(map[string]int)
	unspentTxs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Outputs {
			if out.ToAddressRight(address) {
				accumulated += out.Value
				unspentOuts[txID] = outIdx
				continue Work
			}
		}
	}
	return accumulated, unspentOuts
}

func (bc *BlockChain) FindSpendableOutputs(address []byte, amount int) (int, map[string]int) {
	unspentOuts := make(map[string]int)
	unspentTxs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Outputs {
			if out.ToAddressRight(address) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = outIdx
				if accumulated >= amount {
					break Work
				}
				continue Work
			}
		}
	}
	return accumulated, unspentOuts
}

func (bc *BlockChain) CreateTransaction(fromPubKey, toHashPubKey []byte, amount int, privKey ecdsa.PrivateKey) (*transaction.Transaction, bool) {
	var inputs []transaction.TxInput
	var outputs []transaction.TxOutput

	acc, validOutputs := bc.FindSpendableOutputs(fromPubKey, amount)
	if acc < amount {
		return &transaction.Transaction{}, false
	}
	for txId, outIdx := range validOutputs {
		txID, err := hex.DecodeString(txId)
		utils.HandleErr(err)
		input := transaction.TxInput{TxID: txID, OutIdx: outIdx, PubKey: fromPubKey}
		inputs = append(inputs, input)
	}

	outputs = append(outputs, transaction.TxOutput{Value: amount, HashPubKey: toHashPubKey})
	if acc > amount {
		outputs = append(outputs, transaction.TxOutput{Value: acc - amount, HashPubKey: utils.PublicKeyHash(fromPubKey)})
	}
	tx := transaction.Transaction{Inputs: inputs, Outputs: outputs}
	tx.SetID()

	tx.Sign(privKey)

	return &tx, true
}
