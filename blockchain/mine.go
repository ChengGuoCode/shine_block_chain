package blockchain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"shineBlockChain/transaction"
	"shineBlockChain/utils"
)

func (bc *BlockChain) RunMine() {
	transactionPool := CreateTransactionPool()
	if !bc.VerifyTransaction(transactionPool.PubTx) {
		err := RemoveTransactionPoolFile()
		utils.HandleErr(err)
		return
	}
	candidateBlock := CreateBlock(bc.LastHash, transactionPool.PubTx)
	if candidateBlock.ValidatePoW() {
		bc.AddBlock(candidateBlock)
		err := RemoveTransactionPoolFile()
		utils.HandleErr(err)
	} else {
		fmt.Println("Block has invalid nonce.")
	}
}

func (bc *BlockChain) VerifyTransaction(txs []*transaction.Transaction) bool {
	if len(txs) == 0 {
		return true
	}
	spentOutputs := make(map[string]int)
	for _, tx := range txs {
		pubKey := tx.Inputs[0].PubKey
		unspentOutputs := bc.FindUnspentTransactions(pubKey)
		inputAmount := 0
		outputAmount := 0

		for _, input := range tx.Inputs {
			if outIdx, ok := spentOutputs[hex.EncodeToString(input.TxID)]; ok && outIdx == input.OutIdx {
				return false
			}
			ok, amount := isInputRight(unspentOutputs, input)
			if !ok {
				return false
			}
			inputAmount += amount
			spentOutputs[hex.EncodeToString(input.TxID)] = input.OutIdx
		}

		for _, output := range tx.Outputs {
			outputAmount += output.Value
		}
		if inputAmount != outputAmount {
			return false
		}
		if !tx.Verify() {
			return false
		}
	}
	return true
}

func isInputRight(txs []transaction.Transaction, in transaction.TxInput) (bool, int) {
	for _, tx := range txs {
		if bytes.Equal(tx.ID, in.TxID) {
			return true, tx.Outputs[in.OutIdx].Value
		}
	}
	return false, 0
}
