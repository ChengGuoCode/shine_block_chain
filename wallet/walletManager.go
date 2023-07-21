package wallet

import (
	"bytes"
	"encoding/gob"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"os"
	"path/filepath"
	"shineBlockChain/constcoe"
	"shineBlockChain/utils"
	"strings"
)

type RefList map[string]string

func (r *RefList) Save() {
	filename := constcoe.WalletsRefList + "ref_list.data"
	var content bytes.Buffer
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(r)
	utils.HandleErr(err)
	err = os.WriteFile(filename, content.Bytes(), 0644)
	utils.HandleErr(err)
}

func (r *RefList) Update() {
	err := filepath.Walk(constcoe.Wallets, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		filename := f.Name()
		if strings.Compare(filename[len(filename)-4:], ".wlt") == 0 {
			_, ok := (*r)[filename[:len(filename)-4]]
			if !ok {
				(*r)[filename[:len(filename)-4]] = ""
			}
		}
		return nil
	})
	utils.HandleErr(err)
}

func LoadRefList() *RefList {
	filename := constcoe.WalletsRefList + "ref_list.data"
	var refList RefList
	if utils.FileExists(filename) {
		fileContent, err := os.ReadFile(filename)
		utils.HandleErr(err)
		decoder := gob.NewDecoder(bytes.NewBuffer(fileContent))
		err = decoder.Decode(&refList)
		utils.HandleErr(err)
	} else {
		refList = make(RefList)
		refList.Update()
	}
	return &refList
}

func (r *RefList) BindRef(address, refName string) {
	(*r)[address] = refName
}

func (r *RefList) FindRef(refName string) (string, error) {
	temp := ""
	for key, val := range *r {
		if val == refName {
			temp = key
			break
		}
	}
	if temp == "" {
		err := errors.New("the refName is not found")
		return temp, err
	}
	return temp, nil
}
