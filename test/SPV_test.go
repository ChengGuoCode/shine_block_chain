package test

import (
	"crypto/sha256"
	"shineBlockChain/blockchain"
	"shineBlockChain/transaction"
)

func GenerateTransaction(outCash int, inAccount string, toAccount string, prevTxID string, outIdx int) *transaction.Transaction {
	prevTxIDHash := sha256.Sum256([]byte(prevTxID))
	inAccountHash := sha256.Sum256([]byte(inAccount))
	toAccountHash := sha256.Sum256([]byte(toAccount))
	txIn := transaction.TxInput{TxID: prevTxIDHash[:], OutIdx: outIdx, PubKey: inAccountHash[:]}
	txOut := transaction.TxOutput{Value: outCash, HashPubKey: toAccountHash[:]}
	tx := transaction.Transaction{ID: []byte("This is the Base Transaction!"), Inputs: []transaction.TxInput{txIn}, Outputs: []transaction.TxOutput{txOut}}
	tx.SetID()
	return &tx
}

var transactionTests = []struct {
	outCash   int
	inAccount string
	toAccount string
	prevTxID  string
	outIdx    int
}{
	{
		outCash:   10,
		inAccount: "LLL",
		toAccount: "CCC",
		prevTxID:  "prev1",
		outIdx:    1,
	},
	{
		outCash:   20,
		inAccount: "EEE",
		toAccount: "OOO",
		prevTxID:  "prev2",
		outIdx:    1,
	},
	{
		outCash:   30,
		inAccount: "OOO",
		toAccount: "EEE",
		prevTxID:  "prev3",
		outIdx:    0,
	},
	{
		outCash:   100,
		inAccount: "CCC",
		toAccount: "LLL",
		prevTxID:  "prev4",
		outIdx:    1,
	},
	{
		outCash:   50,
		inAccount: "AAA",
		toAccount: "OOO",
		prevTxID:  "prev5",
		outIdx:    1,
	},
	{
		outCash:   110,
		inAccount: "OOO",
		toAccount: "AAA",
		prevTxID:  "prev6",
		outIdx:    0,
	},
	{
		outCash:   200,
		inAccount: "LLL",
		toAccount: "CCC",
		prevTxID:  "prev7",
		outIdx:    1,
	},
	{
		outCash:   500,
		inAccount: "EEE",
		toAccount: "OOO",
		prevTxID:  "prev8",
		outIdx:    1,
	},
}

func GenerateBlock(txs []*transaction.Transaction, prevBlock string) *blockchain.Block {
	prevBlockHash := sha256.Sum256([]byte(prevBlock))
	testBlock := blockchain.CreateBlock(prevBlockHash[:], txs)
	return testBlock
}

var spvTests = []struct {
	txContained []int
	prevBlock   string
	findTX      []int
	wants       []bool
}{
	{
		txContained: []int{0},
		prevBlock:   "prev1",
		findTX:      []int{0, 1},
		wants:       []bool{true, false},
	},
	{
		txContained: []int{0, 1, 2, 3, 4, 5, 6, 7},
		prevBlock:   "prev2",
		findTX:      []int{3, 7, 5},
		wants:       []bool{true, true, true},
	},
	{
		txContained: []int{0, 1, 2, 3},
		prevBlock:   "prev3",
		findTX:      []int{0, 1, 5},
		wants:       []bool{true, true, false},
	},
	{
		txContained: []int{0, 3, 5, 6, 7},
		prevBlock:   "prev4",
		findTX:      []int{0, 1, 6, 7},
		wants:       []bool{true, false, true, true},
	},
	{
		txContained: []int{0, 1, 2, 4, 5, 6, 7},
		prevBlock:   "prev5",
		findTX:      []int{0, 1, 3},
		wants:       []bool{true, true, false},
	},
}
