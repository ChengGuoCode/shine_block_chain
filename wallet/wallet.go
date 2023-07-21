package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/gob"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"os"
	"shineBlockChain/constcoe"
	"shineBlockChain/utils"
)

type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
}

func NewKeyPair() (*ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	utils.HandleErr(err)
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return privateKey, publicKey
}

func NewWallet() *Wallet {
	privateKey, publicKey := NewKeyPair()
	key, err := x509.MarshalECPrivateKey(privateKey)
	utils.HandleErr(err)
	wallet := Wallet{key, publicKey}
	return &wallet
}

func (w *Wallet) Address() []byte {
	pubHash := utils.PublicKeyHash(w.PublicKey)
	return utils.PubHash2Address(pubHash)
}

func (w *Wallet) Save() {
	filename := constcoe.Wallets + string(w.Address()) + ".wlt"
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(w)
	utils.HandleErr(err)
	err = os.WriteFile(filename, content.Bytes(), 0644)
	utils.HandleErr(err)
}

func LoadWallet(address string) *Wallet {
	filename := constcoe.Wallets + address + ".wlt"
	if !utils.FileExists(filename) {
		utils.HandleErr(errors.New("no wallet with such address"))
	}
	var w Wallet
	gob.Register(elliptic.P256())
	fileContent, err := os.ReadFile(filename)
	utils.HandleErr(err)
	decoder := gob.NewDecoder(bytes.NewBuffer(fileContent))
	err = decoder.Decode(&w)
	utils.HandleErr(err)
	return &w
}
