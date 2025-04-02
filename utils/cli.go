package utils

import (
	"BkC/blockchain"
	"encoding/json"
	"flag"
	"fmt"
	"time"
)

// CLICommand représente une commande CLI avec ses sous-commandes et options
type CLICommand struct {
	Name        string
	Description string
	Handler     func(args []string, bc *blockchain.Blockchain) error
	SubCommands map[string]*CLICommand
}

// CLI est la structure principale pour l'interface en ligne de commande
type CLI struct {
	Commands map[string]*CLICommand
	BC       *blockchain.Blockchain
}

// NewCLI crée une nouvelle instance de l'interface en ligne de commande
func NewCLI(bc *blockchain.Blockchain) *CLI {
	cli := &CLI{
		Commands: make(map[string]*CLICommand),
		BC:       bc,
	}

	// Commande blockchain
	blockchainCmd := &CLICommand{
		Name:        "blockchain",
		Description: "Commandes relatives à la blockchain",
		SubCommands: make(map[string]*CLICommand),
	}

	// Sous-commandes pour blockchain
	blockchainCmd.SubCommands["info"] = &CLICommand{
		Name:        "info",
		Description: "Affiche les informations sur la blockchain",
		Handler:     handleBlockchainInfo,
	}
	blockchainCmd.SubCommands["validate"] = &CLICommand{
		Name:        "validate",
		Description: "Valide l'intégrité de la blockchain",
		Handler:     handleBlockchainValidate,
	}

	// Commande block
	blockCmd := &CLICommand{
		Name:        "block",
		Description: "Commandes relatives aux blocs",
		SubCommands: make(map[string]*CLICommand),
	}

	// Sous-commandes pour block
	blockCmd.SubCommands["get"] = &CLICommand{
		Name:        "get",
		Description: "Récupère un bloc par son hash ou son index",
		Handler:     handleBlockGet,
	}
	blockCmd.SubCommands["mine"] = &CLICommand{
		Name:        "mine",
		Description: "Mine un nouveau bloc",
		Handler:     handleBlockMine,
	}

	// Commande transaction
	txCmd := &CLICommand{
		Name:        "tx",
		Description: "Commandes relatives aux transactions",
		SubCommands: make(map[string]*CLICommand),
	}

	// Sous-commandes pour transaction
	txCmd.SubCommands["create"] = &CLICommand{
		Name:        "create",
		Description: "Crée une nouvelle transaction",
		Handler:     handleTxCreate,
	}
	txCmd.SubCommands["get"] = &CLICommand{
		Name:        "get",
		Description: "Récupère une transaction par son ID",
		Handler:     handleTxGet,
	}
	txCmd.SubCommands["pending"] = &CLICommand{
		Name:        "pending",
		Description: "Liste les transactions en attente",
		Handler:     handleTxPending,
	}

	// Commande wallet
	walletCmd := &CLICommand{
		Name:        "wallet",
		Description: "Commandes relatives aux portefeuilles",
		SubCommands: make(map[string]*CLICommand),
	}

	// Sous-commandes pour wallet
	walletCmd.SubCommands["create"] = &CLICommand{
		Name:        "create",
		Description: "Crée un nouveau portefeuille",
		Handler:     handleWalletCreate,
	}
	walletCmd.SubCommands["list"] = &CLICommand{
		Name:        "list",
		Description: "Liste les portefeuilles",
		Handler:     handleWalletList,
	}
	walletCmd.SubCommands["balance"] = &CLICommand{
		Name:        "balance",
		Description: "Affiche le solde d'un portefeuille",
		Handler:     handleWalletBalance,
	}

	// Commande contract
	contractCmd := &CLICommand{
		Name:        "contract",
		Description: "Commandes relatives aux contrats intelligents",
		SubCommands: make(map[string]*CLICommand),
	}

	// Sous-commandes pour contract
	contractCmd.SubCommands["deploy"] = &CLICommand{
		Name:        "deploy",
		Description: "Déploie un nouveau contrat intelligent",
		Handler:     handleContractDeploy,
	}
	contractCmd.SubCommands["call"] = &CLICommand{
		Name:        "call",
		Description: "Appelle une fonction d'un contrat intelligent",
		Handler:     handleContractCall,
	}
	contractCmd.SubCommands["list"] = &CLICommand{
		Name:        "list",
		Description: "Liste les contrats intelligents",
		Handler:     handleContractList,
	}

	// Ajouter les commandes principales
	cli.Commands["blockchain"] = blockchainCmd
	cli.Commands["block"] = blockCmd
	cli.Commands["tx"] = txCmd
	cli.Commands["wallet"] = walletCmd
	cli.Commands["contract"] = contractCmd

	// Commande help
	cli.Commands["help"] = &CLICommand{
		Name:        "help",
		Description: "Affiche l'aide",
		Handler:     handleHelp,
	}

	return cli
}

// Run exécute l'interface en ligne de commande
func (cli *CLI) Run(args []string) error {
	if len(args) < 2 {
		fmt.Println("Usage: bkc-cli <command> [subcommand] [options]")
		fmt.Println("Pour afficher l'aide, tapez: bkc-cli help")
		return nil
	}

	command := args[1]
	if command == "help" || command == "--help" || command == "-h" {
		return cli.printHelp(args[2:])
	}

	cmd, exists := cli.Commands[command]
	if !exists {
		return fmt.Errorf("commande inconnue: %s", command)
	}

	// S'il n'y a pas de sous-commande, exécuter le handler de la commande principale
	if len(args) < 3 {
		if cmd.Handler != nil {
			return cmd.Handler(args[2:], cli.BC)
		}
		// Si la commande n'a pas de handler mais des sous-commandes, afficher l'aide
		return cli.printCommandHelp(command)
	}

	// Sinon, exécuter la sous-commande
	subCommand := args[2]
	subCmd, exists := cmd.SubCommands[subCommand]
	if !exists {
		return fmt.Errorf("sous-commande inconnue: %s", subCommand)
	}

	if subCmd.Handler != nil {
		return subCmd.Handler(args[3:], cli.BC)
	}

	return fmt.Errorf("sous-commande sans handler: %s %s", command, subCommand)
}

// printHelp affiche l'aide générale ou spécifique à une commande
func (cli *CLI) printHelp(args []string) error {
	if len(args) == 0 {
		fmt.Println("BkC - CryptoChain Go CLI")
		fmt.Println("Usage: bkc-cli <command> [subcommand] [options]")
		fmt.Println("\nCommandes disponibles:")
		for name, cmd := range cli.Commands {
			fmt.Printf("  %-12s %s\n", name, cmd.Description)
		}
		fmt.Println("\nPour plus d'informations sur une commande, tapez: bkc-cli help <command>")
		return nil
	}

	return cli.printCommandHelp(args[0])
}

// printCommandHelp affiche l'aide spécifique à une commande
func (cli *CLI) printCommandHelp(commandName string) error {
	cmd, exists := cli.Commands[commandName]
	if !exists {
		return fmt.Errorf("commande inconnue: %s", commandName)
	}

	fmt.Printf("Commande: %s\n", commandName)
	fmt.Printf("Description: %s\n", cmd.Description)
	fmt.Println("\nSous-commandes disponibles:")
	for name, subCmd := range cmd.SubCommands {
		fmt.Printf("  %-12s %s\n", name, subCmd.Description)
	}
	return nil
}

// Handlers pour les différentes commandes

// handleBlockchainInfo affiche les informations sur la blockchain
func handleBlockchainInfo(args []string, bc *blockchain.Blockchain) error {
	fmt.Println("Informations sur la blockchain:")
	fmt.Printf("Nombre de blocs: %d\n", len(bc.Blocks))
	fmt.Printf("Difficulté actuelle: %d\n", bc.MiningDifficulty)
	fmt.Printf("Récompense de minage: %.2f\n", bc.MiningReward)
	fmt.Printf("Transactions en attente: %d\n", len(bc.MemPool.Transactions))
	fmt.Printf("Hash du dernier bloc: %s\n", bc.Blocks[len(bc.Blocks)-1].Hash)
	return nil
}

// handleBlockchainValidate valide l'intégrité de la blockchain
func handleBlockchainValidate(args []string, bc *blockchain.Blockchain) error {
	valid, err := bc.ValidateChain()
	if err != nil {
		return fmt.Errorf("erreur lors de la validation: %v", err)
	}

	if valid {
		fmt.Println("La blockchain est valide.")
	} else {
		fmt.Println("La blockchain est invalide!")
	}
	return nil
}

// handleBlockGet récupère un bloc par son hash ou son index
func handleBlockGet(args []string, bc *blockchain.Blockchain) error {
	flags := flag.NewFlagSet("block get", flag.ExitOnError)
	hashFlag := flags.String("hash", "", "Hash du bloc")
	indexFlag := flags.Int("index", -1, "Index du bloc")
	flags.Parse(args)

	var block *blockchain.Block

	if *hashFlag != "" {
		block = bc.GetBlockByHash(*hashFlag)
	} else if *indexFlag >= 0 {
		block = bc.GetBlockByIndex(*indexFlag)
	} else {
		return fmt.Errorf("vous devez spécifier un hash ou un index")
	}

	if block == nil {
		return fmt.Errorf("bloc introuvable")
	}

	// Formater et afficher le bloc
	blockData, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur lors du formatage du bloc: %v", err)
	}

	fmt.Println(string(blockData))
	return nil
}

// handleBlockMine mine un nouveau bloc
func handleBlockMine(args []string, bc *blockchain.Blockchain) error {
	flags := flag.NewFlagSet("block mine", flag.ExitOnError)
	minerAddress := flags.String("miner", "cli-user", "Adresse du mineur")
	flags.Parse(args)

	fmt.Println("Minage d'un nouveau bloc en cours...")
	startTime := time.Now()

	newBlock := bc.CreateBlock(*minerAddress)

	duration := time.Since(startTime)
	fmt.Printf("Bloc #%d miné en %v\n", newBlock.Index, duration)
	fmt.Printf("Hash: %s\n", newBlock.Hash)
	fmt.Printf("Transactions: %d\n", len(newBlock.Transactions))
	fmt.Printf("Nonce: %d\n", newBlock.Nonce)
	return nil
}

// handleTxCreate crée une nouvelle transaction
func handleTxCreate(args []string, bc *blockchain.Blockchain) error {
	flags := flag.NewFlagSet("tx create", flag.ExitOnError)
	sender := flags.String("from", "", "Adresse de l'expéditeur")
	recipient := flags.String("to", "", "Adresse du destinataire")
	amount := flags.Float64("amount", 0.0, "Montant à envoyer")
	fee := flags.Float64("fee", 0.001, "Frais de transaction")
	flags.Parse(args)

	if *sender == "" || *recipient == "" || *amount <= 0 {
		return fmt.Errorf("les paramètres from, to et amount sont requis")
	}

	// Créer la transaction
	tx := &blockchain.Transaction{
		Sender:    *sender,
		Recipient: *recipient,
		Amount:    *amount,
		Fee:       *fee,
		Timestamp: time.Now(),
	}

	// Ajouter à la blockchain
	if err := bc.AddTransaction(tx); err != nil {
		return fmt.Errorf("erreur lors de l'ajout de la transaction: %v", err)
	}

	fmt.Printf("Transaction créée avec succès: %s\n", tx.ID)
	fmt.Printf("De: %s\n", tx.Sender)
	fmt.Printf("Vers: %s\n", tx.Recipient)
	fmt.Printf("Montant: %.2f\n", tx.Amount)
	fmt.Printf("Frais: %.4f\n", tx.Fee)
	return nil
}

// handleTxGet récupère une transaction par son ID
func handleTxGet(args []string, bc *blockchain.Blockchain) error {
	if len(args) < 1 {
		return fmt.Errorf("vous devez spécifier l'ID de la transaction")
	}

	txID := args[0]
	found := false

	// Chercher dans les transactions en attente
	for _, tx := range bc.MemPool.Transactions {
		if tx.ID == txID {
			txData, _ := json.MarshalIndent(tx, "", "  ")
			fmt.Println("Transaction en attente:")
			fmt.Println(string(txData))
			found = true
			break
		}
	}

	// Si non trouvé, chercher dans les blocs
	if !found {
		for _, block := range bc.Blocks {
			for _, tx := range block.Transactions {
				if tx.ID == txID {
					txData, _ := json.MarshalIndent(tx, "", "  ")
					fmt.Printf("Transaction trouvée dans le bloc #%d:\n", block.Index)
					fmt.Println(string(txData))
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("transaction introuvable: %s", txID)
	}
	return nil
}

// handleTxPending liste les transactions en attente
func handleTxPending(args []string, bc *blockchain.Blockchain) error {
	flags := flag.NewFlagSet("tx pending", flag.ExitOnError)
	limit := flags.Int("limit", 10, "Nombre maximum de transactions à afficher")
	flags.Parse(args)

	pendingTxs := bc.MemPool.GetTransactions(*limit)
	fmt.Printf("Transactions en attente (%d):\n", len(pendingTxs))

	for i, tx := range pendingTxs {
		fmt.Printf("%d. ID: %s, De: %s, Vers: %s, Montant: %.2f, Frais: %.4f\n",
			i+1, tx.ID, tx.Sender, tx.Recipient, tx.Amount, tx.Fee)
	}
	return nil
}

// handleWalletCreate crée un nouveau portefeuille
func handleWalletCreate(args []string, bc *blockchain.Blockchain) error {
	wallet, err := blockchain.NewWallet()
	if err != nil {
		return fmt.Errorf("erreur lors de la création du portefeuille: %v", err)
	}

	fmt.Println("Nouveau portefeuille créé avec succès!")
	fmt.Printf("Adresse: %s\n", wallet.Address)
	fmt.Println("ATTENTION: Conservez votre clé privée en lieu sûr!")
	return nil
}

// handleWalletList liste les portefeuilles
func handleWalletList(args []string, bc *blockchain.Blockchain) error {
	fmt.Println("Cette fonctionnalité nécessite d'accéder au système de fichiers.")
	fmt.Println("Pour l'instant, veuillez utiliser l'interface web pour gérer vos portefeuilles.")
	return nil
}

// handleWalletBalance affiche le solde d'un portefeuille
func handleWalletBalance(args []string, bc *blockchain.Blockchain) error {
	if len(args) < 1 {
		return fmt.Errorf("vous devez spécifier l'adresse du portefeuille")
	}

	address := args[0]
	balance := bc.GetBalance(address)

	fmt.Printf("Solde du portefeuille %s: %.2f BCK\n", address, balance)
	return nil
}

// handleContractDeploy déploie un nouveau contrat intelligent
func handleContractDeploy(args []string, bc *blockchain.Blockchain) error {
	flags := flag.NewFlagSet("contract deploy", flag.ExitOnError)
	contractType := flags.String("type", "TRANSFER", "Type de contrat (TRANSFER, MULTISIG, TIMELOCK, CONDITIONAL, ESCROW)")
	creator := flags.String("creator", "", "Adresse du créateur")
	recipient := flags.String("recipient", "", "Adresse du destinataire")
	amount := flags.Float64("amount", 0.0, "Montant du contrat")
	fee := flags.Float64("fee", 0.001, "Frais de transaction")
	data := flags.String("data", "", "Données supplémentaires")
	expiresIn := flags.Int("expires", 24, "Expiration en heures")
	flags.Parse(args)

	if *creator == "" || *recipient == "" || *amount <= 0 {
		return fmt.Errorf("les paramètres creator, recipient et amount sont requis")
	}

	// Convertir le type de contrat
	var contractTypeEnum blockchain.ContractType
	switch *contractType {
	case "TRANSFER":
		contractTypeEnum = blockchain.ContractTransfer
	case "MULTISIG":
		contractTypeEnum = blockchain.ContractMultiSig
	case "TIMELOCK":
		contractTypeEnum = blockchain.ContractTimeLock
	case "CONDITIONAL":
		contractTypeEnum = blockchain.ContractCondition
	case "ESCROW":
		contractTypeEnum = blockchain.ContractEscrow
	default:
		return fmt.Errorf("type de contrat invalide: %s", *contractType)
	}

	// Créer le contrat
	contract, err := blockchain.NewSmartContract(
		contractTypeEnum,
		*creator,
		[]string{*creator}, // Simplification: créateur comme seul participant
		1,                  // Simplification: 1 approbation requise
		*amount,
		*fee,
		*recipient,
		*data,
		time.Duration(*expiresIn)*time.Hour,
		make(map[string]string), // Pas de conditions pour cet exemple
	)

	if err != nil {
		return fmt.Errorf("erreur lors de la création du contrat: %v", err)
	}

	// Sauvegarder le contrat
	if err := bc.SaveContract(contract); err != nil {
		return fmt.Errorf("erreur lors de la sauvegarde du contrat: %v", err)
	}

	fmt.Printf("Contrat déployé avec succès: %s\n", contract.ID)
	fmt.Printf("Type: %s\n", contractTypeEnum)
	fmt.Printf("Créateur: %s\n", contract.CreatedBy)
	fmt.Printf("Destinataire: %s\n", contract.Recipient)
	fmt.Printf("Montant: %.2f\n", contract.Amount)
	fmt.Printf("Expire le: %s\n", contract.ExpiresAt.Format("02/01/2006 15:04:05"))
	return nil
}

// handleContractCall appelle une fonction d'un contrat intelligent
func handleContractCall(args []string, bc *blockchain.Blockchain) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: bkc-cli contract call <contract_id> <action> [options]")
	}

	contractID := args[0]
	action := args[1]

	contract := bc.GetContractByID(contractID)
	if contract == nil {
		return fmt.Errorf("contrat introuvable: %s", contractID)
	}

	switch action {
	case "approve":
		flags := flag.NewFlagSet("contract approve", flag.ExitOnError)
		participant := flags.String("participant", "", "Adresse du participant qui approuve")
		flags.Parse(args[2:])

		if *participant == "" {
			return fmt.Errorf("l'adresse du participant est requise")
		}

		if err := contract.ApproveContract(*participant); err != nil {
			return fmt.Errorf("erreur lors de l'approbation: %v", err)
		}

		if err := bc.UpdateContract(contract); err != nil {
			return fmt.Errorf("erreur lors de la mise à jour du contrat: %v", err)
		}

		fmt.Println("Contrat approuvé avec succès!")

	case "execute":
		if !contract.CanExecute() {
			return fmt.Errorf("le contrat ne peut pas être exécuté actuellement")
		}

		tx, err := contract.ExecuteContract(bc)
		if err != nil {
			return fmt.Errorf("erreur lors de l'exécution: %v", err)
		}

		txID, err := bc.UpdateContractToBlockchain(contract)
		if err != nil {
			return fmt.Errorf("erreur lors de la mise à jour dans la blockchain: %v", err)
		}

		fmt.Println("Contrat exécuté avec succès!")
		fmt.Printf("Transaction créée: %s\n", txID)
		fmt.Printf("De: %s\n", tx.Sender)
		fmt.Printf("Vers: %s\n", tx.Recipient)
		fmt.Printf("Montant: %.2f\n", tx.Amount)

	case "cancel":
		flags := flag.NewFlagSet("contract cancel", flag.ExitOnError)
		canceller := flags.String("canceller", "", "Adresse de l'annulateur")
		flags.Parse(args[2:])

		if *canceller == "" {
			return fmt.Errorf("l'adresse de l'annulateur est requise")
		}

		if err := contract.CancelContract(*canceller); err != nil {
			return fmt.Errorf("erreur lors de l'annulation: %v", err)
		}

		if err := bc.UpdateContract(contract); err != nil {
			return fmt.Errorf("erreur lors de la mise à jour du contrat: %v", err)
		}

		fmt.Println("Contrat annulé avec succès!")

	default:
		return fmt.Errorf("action inconnue: %s", action)
	}

	return nil
}

// handleContractList liste les contrats intelligents
func handleContractList(args []string, bc *blockchain.Blockchain) error {
	flags := flag.NewFlagSet("contract list", flag.ExitOnError)
	user := flags.String("user", "", "Filtrer par utilisateur (créateur ou participant)")
	flags.Parse(args)

	var contracts []*blockchain.SmartContract
	if *user != "" {
		contracts = bc.GetContractsForUser(*user)
	} else {
		fmt.Println("Pour l'instant, veuillez spécifier un utilisateur avec --user")
		return nil
	}

	fmt.Printf("Contrats (%d):\n", len(contracts))
	for i, contract := range contracts {
		fmt.Printf("%d. ID: %s, Type: %s, Créateur: %s, Destinataire: %s, Montant: %.2f, Statut: %s\n",
			i+1, contract.ID, contract.Type, contract.CreatedBy, contract.Recipient, contract.Amount, contract.Status)
	}
	return nil
}

// handleHelp affiche l'aide
func handleHelp(args []string, bc *blockchain.Blockchain) error {
	fmt.Println("BkC - CryptoChain Go CLI")
	fmt.Println("Usage: bkc-cli <command> [subcommand] [options]")
	fmt.Println("\nCommandes disponibles:")
	fmt.Println("  blockchain   Commandes relatives à la blockchain")
	fmt.Println("  block        Commandes relatives aux blocs")
	fmt.Println("  tx           Commandes relatives aux transactions")
	fmt.Println("  wallet       Commandes relatives aux portefeuilles")
	fmt.Println("  contract     Commandes relatives aux contrats intelligents")
	fmt.Println("  help         Affiche l'aide")
	return nil
}
