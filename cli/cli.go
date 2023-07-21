package cli

import (
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"runtime"
	"shineBlockChain/blockchain"
	"shineBlockChain/utils"
	"shineBlockChain/wallet"
	"strconv"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Welcome to Leo tiny blockchain system, usage is as follows:")
	fmt.Println("--------------------------------------------------------------------------------------------------------------")
	fmt.Println("All you need is to first create a blockchain and declare the owner.")
	fmt.Println("And then you can make transactions.")
	fmt.Println("--------------------------------------------------------------------------------------------------------------")
	fmt.Println("createblockchain -address ADDRESS                   ----> Creates a blockchain with the owner you input")
	fmt.Println("balance -address ADDRESS                            ----> Back the balance of the address you input")
	fmt.Println("blockchaininfo                                      ----> Prints the blocks in the chain")
	fmt.Println("send -from FROADDRESS -to TOADDRESS -amount AMOUNT  ----> Make a transaction and put it into candidate block")
	fmt.Println("mine                                                ----> Mine and add a block to the chain")
	fmt.Println("--------------------------------------------------------------------------------------------------------------")
}

func (cli *CommandLine) createBlockChain(address string) {
	newChain := blockchain.InitBlockChain(utils.Address2PubHash([]byte(address)))
	err := newChain.Database.Close()
	utils.HandleErr(err)
}

func (cli *CommandLine) balance(address string) {
	chain := blockchain.ContinueBlockChain()
	defer func(Database *leveldb.DB) {
		err := Database.Close()
		utils.HandleErr(err)
	}(chain.Database)

	loadWallet := wallet.LoadWallet(address)
	balance, _ := chain.FindUTXOs(loadWallet.PublicKey)
	fmt.Printf("Address: %s, Balance: %d \n", address, balance)
}

func (cli *CommandLine) getBlockChainInfo() {
	chain := blockchain.ContinueBlockChain()
	defer func(Database *leveldb.DB) {
		err := Database.Close()
		utils.HandleErr(err)
	}(chain.Database)
	iterator := chain.Iterator()
	for iterator.HasNext() {
		block := iterator.Next()
		fmt.Println("--------------------------------------------------------------------------------------------------------------")
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		fmt.Printf("Transactions: %v\n", block.Transactions)
		fmt.Printf("hash: %x\n", block.Hash)
		fmt.Printf("Pow: %s\n", strconv.FormatBool(block.ValidatePoW()))
		fmt.Println("--------------------------------------------------------------------------------------------------------------")
		fmt.Println()
	}
}

func (cli *CommandLine) send(from, to string, amount int) {
	chain := blockchain.ContinueBlockChain()
	defer func(Database *leveldb.DB) {
		err := Database.Close()
		utils.HandleErr(err)
	}(chain.Database)
	fromWallet := wallet.LoadWallet(from)
	key, err := x509.ParseECPrivateKey(fromWallet.PrivateKey)
	utils.HandleErr(err)
	tx, ok := chain.CreateTransaction(fromWallet.PublicKey, utils.Address2PubHash([]byte(to)), amount, *key)
	if !ok {
		fmt.Println("Failed to create transaction")
		return
	}
	tp := blockchain.CreateTransactionPool()
	tp.AddTransaction(tx)
	tp.SaveFile()
	fmt.Println("Success!")
}

func (cli *CommandLine) mine() {
	chain := blockchain.ContinueBlockChain()
	defer func(Database *leveldb.DB) {
		err := Database.Close()
		utils.HandleErr(err)
	}(chain.Database)
	chain.RunMine()
	fmt.Println("Finish Mining")
}

func (cli *CommandLine) createWallet(refName string) {
	newWallet := wallet.NewWallet()
	newWallet.Save()
	refList := wallet.LoadRefList()
	refList.BindRef(string(newWallet.Address()), refName)
	refList.Save()
	fmt.Println("Succeed in creating wallet.")
}

func (cli *CommandLine) walletInfoRefName(refName string) {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refName)
	utils.HandleErr(err)
	cli.walletInfo(address)
}

func (cli *CommandLine) walletInfo(address string) {
	wlt := wallet.LoadWallet(address)
	refList := wallet.LoadRefList()
	fmt.Printf("Wallet address: %x\n", wlt.Address())
	fmt.Printf("Public Key: %x\n", wlt.PublicKey)
	fmt.Printf("Reference Name: %s\n", (*refList)[address])
}

func (cli *CommandLine) walletsUpdate() {
	refList := wallet.LoadRefList()
	refList.Update()
	refList.Save()
	fmt.Println("Succeed in updating wallets.")
}

func (cli *CommandLine) walletsList() {
	refList := wallet.LoadRefList()
	for address, _ := range *refList {
		wlt := wallet.LoadWallet(address)
		fmt.Println("--------------------------------------------------------------------------------------------------------------")
		fmt.Printf("Wallet address:%s\n", address)
		fmt.Printf("Public Key:%x\n", wlt.PublicKey)
		fmt.Printf("Reference Name:%s\n", (*refList)[address])
		fmt.Println("--------------------------------------------------------------------------------------------------------------")
		fmt.Println()
	}
}

func (cli *CommandLine) sendRefName(fromRefName, toRefName string, amount int) {
	refList := wallet.LoadRefList()
	fromAddress, err := refList.FindRef(fromRefName)
	utils.HandleErr(err)
	toAddress, err := refList.FindRef(toRefName)
	utils.HandleErr(err)
	cli.send(fromAddress, toAddress, amount)
}

func (cli *CommandLine) createBlockChainRefName(refName string) {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refName)
	utils.HandleErr(err)
	cli.createBlockChain(address)
}

func (cli *CommandLine) balanceRefName(refName string) {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refName)
	utils.HandleErr(err)
	cli.balance(address)
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)    // new command
	walletInfoCmd := flag.NewFlagSet("walletinfo", flag.ExitOnError)        // new command
	walletsUpdateCmd := flag.NewFlagSet("walletsupdate ", flag.ExitOnError) // new command
	walletsListCmd := flag.NewFlagSet("walletslist", flag.ExitOnError)      // new command
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	getBlockChainInfoCmd := flag.NewFlagSet("blockchaininfo", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendByRefNameCmd := flag.NewFlagSet("sendbyrefname", flag.ExitOnError)
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)

	createWalletRefName := createWalletCmd.String("refname", "", "The refname of the wallet, and this is optimal") // this line is new
	walletInfoRefName := walletInfoCmd.String("refname", "", "The refname of the wallet")                          // this line is new
	walletInfoAddress := walletInfoCmd.String("address", "", "The address of the wallet")                          // this line is new
	createBlockChainOwner := createBlockChainCmd.String("address", "", "The address refer to the owner of blockchain")
	createBlockChainByRefNameOwner := createBlockChainCmd.String("refname", "", "The name refer to the owner of blockchain") // this line is new
	balanceAddress := balanceCmd.String("address", "", "Who need to get balance amount")
	balanceRefName := balanceCmd.String("refname", "", "Who needs to get balance amount") // this line is new
	sendByRefNameFrom := sendByRefNameCmd.String("from", "", "Source refname")            // this line is new
	sendByRefNameTo := sendByRefNameCmd.String("to", "", "Destination refname")           // this line is new
	sendByRefNameAmount := sendByRefNameCmd.Int("amount", 0, "Amount to send")            // this line is new
	sendFromAddress := sendCmd.String("from", "", "Source address")
	sendToAddress := sendCmd.String("to", "", "Destination address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "createwallet": // this case is new
		err := createWalletCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "walletinfo": // this case is new
		err := walletInfoCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "walletsupdate": // this case is new
		err := walletsUpdateCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "walletslist": // this case is new
		err := walletsListCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "createblockchain":
		err := createBlockChainCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "balance":
		err := balanceCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "blockchaininfo":
		err := getBlockChainInfoCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "sendbyrefname": // this case is new
		err := sendByRefNameCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "mine":
		err := mineCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(*createWalletRefName)
	}

	if walletInfoCmd.Parsed() {
		if *walletInfoAddress == "" {
			if *walletInfoRefName == "" {
				walletInfoCmd.Usage()
				runtime.Goexit()
			} else {
				cli.walletInfoRefName(*walletInfoRefName)
			}
		} else {
			cli.walletInfo(*walletInfoAddress)
		}
	}

	if walletsUpdateCmd.Parsed() {
		cli.walletsUpdate()
	}

	if walletsListCmd.Parsed() {
		cli.walletsList()
	}

	if createBlockChainCmd.Parsed() {
		if *createBlockChainOwner == "" {
			if *createBlockChainByRefNameOwner == "" {
				createBlockChainCmd.Usage()
				runtime.Goexit()
			} else {
				cli.createBlockChainRefName(*createBlockChainByRefNameOwner)
			}
		} else {
			cli.createBlockChain(*createBlockChainOwner)
		}
	}

	if balanceCmd.Parsed() {
		if *balanceAddress == "" {
			if *balanceRefName == "" {
				balanceCmd.Usage()
				runtime.Goexit()
			} else {
				cli.balanceRefName(*balanceRefName)
			}
		} else {
			cli.balance(*balanceAddress)
		}
	}

	if sendCmd.Parsed() {
		if *sendFromAddress == "" || *sendToAddress == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFromAddress, *sendToAddress, *sendAmount)
	}

	if sendByRefNameCmd.Parsed() {
		if *sendByRefNameFrom == "" || *sendByRefNameTo == "" || *sendByRefNameAmount <= 0 {
			sendByRefNameCmd.Usage()
			runtime.Goexit()
		}
		cli.sendRefName(*sendByRefNameFrom, *sendByRefNameTo, *sendByRefNameAmount)
	}

	if getBlockChainInfoCmd.Parsed() {
		cli.getBlockChainInfo()
	}

	if mineCmd.Parsed() {
		cli.mine()
	}
}
