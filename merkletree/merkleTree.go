package merkletree

import (
	"bytes"
	"crypto/sha256"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"shineBlockChain/transaction"
	"shineBlockChain/utils"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	LeftNode  *MerkleNode
	RightNode *MerkleNode
	Data      []byte
}

func CreateMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	tempNode := MerkleNode{}

	if left == nil && right == nil {
		tempNode.Data = data
	} else {
		catenateHash := append(left.Data, right.Data...)
		hash := sha256.Sum256(catenateHash)
		tempNode.Data = hash[:]
	}

	tempNode.LeftNode = left
	tempNode.RightNode = right

	return &tempNode
}

func CreateMerkleTree(txs []*transaction.Transaction) *MerkleTree {
	txsLen := len(txs)
	if txsLen%2 != 0 {
		txs = append(txs, txs[txsLen-1])
	}

	var nodePool []*MerkleNode

	for _, tx := range txs {
		nodePool = append(nodePool, CreateMerkleNode(nil, nil, tx.ID))
	}

	for len(nodePool) > 1 {
		var tempNodePool []*MerkleNode
		poolLen := len(nodePool)
		if poolLen%2 != 0 {
			tempNodePool = append(tempNodePool, nodePool[poolLen-1])
		}
		for i := 0; i < poolLen/2; i++ {
			tempNodePool = append(tempNodePool, CreateMerkleNode(nodePool[2*i], nodePool[2*i+1], nil))
		}
		nodePool = tempNodePool
	}

	merkleTree := MerkleTree{nodePool[0]}

	return &merkleTree
}

func (mn *MerkleNode) Find(data []byte, route []int, hashRoute [][]byte) (bool, []int, [][]byte) {
	findFlag := false

	if bytes.Equal(mn.Data, data) {
		findFlag = true
		return findFlag, route, hashRoute
	}
	if mn.LeftNode != nil {
		routeT := append(route, 0)
		hashRouteT := append(hashRoute, mn.RightNode.Data)
		findFlag, routeT, hashRouteT = mn.LeftNode.Find(data, routeT, hashRouteT)
		if findFlag {
			return findFlag, routeT, hashRouteT
		}
		if mn.RightNode != nil {
			routeT = append(route, 1)
			hashRouteT = append(hashRoute, mn.LeftNode.Data)
			findFlag, routeT, hashRouteT = mn.RightNode.Find(data, routeT, hashRouteT)
			if findFlag {
				return findFlag, routeT, hashRouteT
			}
			return findFlag, route, hashRoute
		}
	}
	return findFlag, route, hashRoute
}

func (mt *MerkleTree) BackValidationRoute(txid []byte) ([]int, [][]byte, bool) {
	ok, route, hashRoute := mt.RootNode.Find(txid, []int{}, [][]byte{})
	return route, hashRoute, ok
}

func SimplePaymentValidation(txId, mtRootHash []byte, route []int, hashRoute [][]byte) bool {
	routeLen := len(route)
	var tempHash []byte
	tempHash = txId

	for i := routeLen - 1; i >= 0; i-- {
		if route[i] == 0 {
			catenateHash := append(tempHash, hashRoute[i]...)
			hash := sha256.Sum256(catenateHash)
			tempHash = hash[:]
		} else if route[i] == 1 {
			catenateHash := append(hashRoute[i], tempHash...)
			hash := sha256.Sum256(catenateHash)
			tempHash = hash[:]
		} else {
			utils.HandleErr(errors.New("error in validation route"))
		}
	}
	return bytes.Equal(tempHash, mtRootHash)
}
