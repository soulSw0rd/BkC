package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Mutex global pour la protection des accès aux sessions
var mu sync.Mutex

// ContractRequest représente une demande de création de contrat
type ContractRequest struct {
	Type              string            `json:"type"`
	Amount            float64           `json:"amount"`
	Fee               float64           `json:"fee"`
	Recipient         string            `json:"recipient"`
	Data              string            `json:"data"`
	ExpiresAt         string            `json:"expiresAt,omitempty"`
	Participants      []string          `json:"participants,omitempty"`
	RequiredApprovals int               `json:"requiredApprovals,omitempty"`
	Conditions        map[string]string `json:"conditions,omitempty"`
}

// ContractResponse représente la réponse après création d'un contrat
type ContractResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ContractID string `json:"contractId,omitempty"`
}

// ContractsHandler gère les requêtes relatives aux contrats intelligents
func ContractsHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)
		w.Header().Set("Content-Type", "application/json")

		// Extraction de l'identifiant du contrat et de l'action depuis l'URL, si présent
		// Format : /api/contracts/{contract_id}/{action}
		parts := strings.Split(r.URL.Path, "/")
		var contractID, action string

		// Extraction de l'ID du contrat s'il est présent dans l'URL
		if len(parts) >= 4 && parts[3] != "" {
			contractID = parts[3]
		}

		// Extraction de l'action s'il est présent dans l'URL
		if len(parts) >= 5 {
			action = parts[4]
		}

		// Vérification de l'authentification via cookie
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Authentification requise", http.StatusUnauthorized)
			return
		}

		mu.Lock()
		session, exists := sessions[cookie.Value]
		mu.Unlock()
		if !exists || session == nil {
			http.Error(w, "Session invalide", http.StatusUnauthorized)
			return
		}

		// Traçage de l'événement d'audit
		utils.LogAuditEvent(
			utils.EventTypeContractAction,
			session.Username,
			r.RemoteAddr,
			fmt.Sprintf("Accès à l'API des contrats: %s %s", r.Method, r.URL.Path),
			utils.RiskLow,
			nil,
		)

		// Gérer les différentes méthodes HTTP
		switch r.Method {
		case "GET":
			// GET /api/contracts - liste tous les contrats
			// GET /api/contracts/{contract_id} - détails d'un contrat spécifique
			if contractID == "" {
				listContracts(w, r, bc, session.Username)
			} else {
				getContractDetails(w, r, bc, contractID, session.Username)
			}

		case "POST":
			// POST /api/contracts - crée un nouveau contrat
			// POST /api/contracts/{contract_id}/approve - approuve un contrat
			// POST /api/contracts/{contract_id}/cancel - annule un contrat
			// POST /api/contracts/{contract_id}/execute - exécute un contrat
			if contractID == "" {
				createContract(w, r, bc, session.Username)
			} else {
				switch action {
				case "approve":
					approveContract(w, r, bc, contractID, session.Username)
				case "cancel":
					cancelContract(w, r, bc, contractID, session.Username)
				case "execute":
					executeContract(w, r, bc, contractID, session.Username)
				default:
					http.Error(w, "Action non reconnue", http.StatusBadRequest)
				}
			}

		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}

// listContracts liste tous les contrats disponibles pour un utilisateur
func listContracts(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, username string) {
	// Récupérer les contrats depuis la blockchain
	contracts := getContractsForUser(bc, username)

	// Répondre avec la liste des contrats
	response := map[string]interface{}{
		"success":   true,
		"contracts": contracts,
		"count":     len(contracts),
	}

	// Journaliser l'action
	utils.LogAuditEvent(
		utils.EventTypeContractListed,
		username,
		r.RemoteAddr,
		fmt.Sprintf("Liste des contrats récupérée (%d contrats)", len(contracts)),
		utils.RiskLow,
		nil,
	)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erreur lors de l'encodage de la réponse: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// getContractDetails récupère les détails d'un contrat spécifique
func getContractDetails(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, contractID, username string) {
	// Récupérer le contrat depuis la blockchain
	contract := getContractByID(bc, contractID)
	if contract == nil {
		http.Error(w, "Contrat non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier si l'utilisateur a accès à ce contrat
	if !hasAccessToContract(contract, username) {
		utils.LogAuditEvent(
			utils.EventTypeSecurityAlert,
			username,
			r.RemoteAddr,
			fmt.Sprintf("Tentative d'accès non autorisé au contrat %s", contractID),
			utils.RiskMedium,
			map[string]interface{}{"contract_id": contractID},
		)
		http.Error(w, "Non autorisé à accéder à ce contrat", http.StatusForbidden)
		return
	}

	// Répondre avec les détails du contrat
	response := map[string]interface{}{
		"success":  true,
		"contract": contract,
	}

	// Journaliser l'action
	utils.LogAuditEvent(
		utils.EventTypeContractViewed,
		username,
		r.RemoteAddr,
		fmt.Sprintf("Détails du contrat %s consultés", contractID),
		utils.RiskLow,
		map[string]interface{}{"contract_id": contractID},
	)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erreur lors de l'encodage de la réponse: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// createContract crée un nouveau contrat intelligent
func createContract(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, username string) {
	// Décoder la requête
	var req ContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Format de requête invalide: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Valider les champs obligatoires
	if req.Amount <= 0 {
		http.Error(w, "Le montant doit être positif", http.StatusBadRequest)
		return
	}

	if req.Fee < 0 {
		http.Error(w, "Les frais ne peuvent pas être négatifs", http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, "Le destinataire est requis", http.StatusBadRequest)
		return
	}

	// Vérifier si l'utilisateur a assez de fonds
	balance := bc.GetBalance(username)
	if balance < req.Amount+req.Fee {
		http.Error(w, "Solde insuffisant pour créer ce contrat", http.StatusBadRequest)
		return
	}

	// Vérifier le type de contrat
	contractType, err := getContractType(req.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Traitement spécifique selon le type de contrat
	var expiresIn time.Duration
	var participants []string
	var requiredApprovals int
	var conditions map[string]string

	// Date d'expiration par défaut (+24h)
	expiresIn = 24 * time.Hour

	// Définir la date d'expiration personnalisée pour les contrats qui le supportent
	if req.ExpiresAt != "" {
		expTime, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			http.Error(w, "Format de date d'expiration invalide", http.StatusBadRequest)
			return
		}

		// Calculer la durée à partir de maintenant
		expiresIn = expTime.Sub(time.Now())
		if expiresIn <= 0 {
			http.Error(w, "La date d'expiration doit être dans le futur", http.StatusBadRequest)
			return
		}
	}

	// Traitement des participants pour les contrats multi-signatures
	if contractType == blockchain.ContractMultiSig || contractType == blockchain.ContractEscrow {
		if len(req.Participants) == 0 {
			http.Error(w, "Liste de participants requise", http.StatusBadRequest)
			return
		}

		participants = req.Participants
		requiredApprovals = req.RequiredApprovals

		if requiredApprovals <= 0 || requiredApprovals > len(participants) {
			http.Error(w, "Nombre d'approbations requises invalide", http.StatusBadRequest)
			return
		}
	}

	// Traitement des conditions pour les contrats conditionnels
	if contractType == blockchain.ContractCondition || contractType == blockchain.ContractEscrow {
		if len(req.Conditions) == 0 {
			http.Error(w, "Conditions requises", http.StatusBadRequest)
			return
		}

		conditions = req.Conditions
	}

	// Ajouter l'utilisateur actuel comme participant s'il n'est pas déjà inclus
	if len(participants) > 0 {
		isAlreadyParticipant := false
		for _, p := range participants {
			if p == username {
				isAlreadyParticipant = true
				break
			}
		}

		if !isAlreadyParticipant {
			participants = append(participants, username)
		}
	}

	// Création du contrat intelligent
	contract, err := blockchain.NewSmartContract(
		contractType,
		username,
		participants,
		requiredApprovals,
		req.Amount,
		req.Fee,
		req.Recipient,
		req.Data,
		expiresIn,
		conditions,
	)

	if err != nil {
		http.Error(w, "Erreur lors de la création du contrat: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Sauvegarder le contrat dans la blockchain
	if err := saveContract(bc, contract); err != nil {
		http.Error(w, "Erreur lors de la sauvegarde du contrat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Journaliser l'événement
	utils.LogAuditEvent(
		utils.EventTypeContractCreated,
		username,
		r.RemoteAddr,
		fmt.Sprintf("Création d'un contrat intelligent de type %s", contractType),
		utils.RiskLow,
		map[string]interface{}{
			"contract_id": contract.ID,
			"type":        contractType,
			"amount":      req.Amount,
		},
	)

	// Répondre avec succès
	response := ContractResponse{
		Success:    true,
		Message:    "Contrat créé avec succès",
		ContractID: contract.ID,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erreur lors de l'encodage de la réponse: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// approveContract approuve un contrat existant
func approveContract(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, contractID, username string) {
	// Récupérer le contrat
	contract := getContractByID(bc, contractID)
	if contract == nil {
		http.Error(w, "Contrat non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier si l'utilisateur est un participant autorisé
	isParticipant := false
	for _, p := range contract.Participants {
		if p == username {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		utils.LogAuditEvent(
			utils.EventTypeSecurityAlert,
			username,
			r.RemoteAddr,
			fmt.Sprintf("Tentative d'approbation non autorisée du contrat %s", contractID),
			utils.RiskMedium,
			map[string]interface{}{"contract_id": contractID},
		)
		http.Error(w, "Non autorisé à approuver ce contrat", http.StatusForbidden)
		return
	}

	// Approuver le contrat
	err := contract.ApproveContract(username)
	if err != nil {
		http.Error(w, "Erreur lors de l'approbation: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Sauvegarder le contrat mis à jour
	if err := updateContract(bc, contract); err != nil {
		http.Error(w, "Erreur lors de la mise à jour du contrat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Journaliser l'action
	utils.LogAuditEvent(
		utils.EventTypeContractApproved,
		username,
		r.RemoteAddr,
		fmt.Sprintf("Approbation du contrat %s", contractID),
		utils.RiskLow,
		map[string]interface{}{"contract_id": contractID},
	)

	// Vérifier si le contrat peut maintenant être exécuté
	if contract.CanExecute() {
		// Exécuter automatiquement le contrat
		tx, err := contract.ExecuteContract(bc)
		if err != nil {
			http.Error(w, "Erreur lors de l'exécution automatique: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Mettre à jour le contrat dans la blockchain
		txID, err := bc.UpdateContractToBlockchain(contract)
		if err != nil {
			http.Error(w, "Erreur lors de la mise à jour dans la blockchain: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"success":     true,
			"message":     "Contrat approuvé et exécuté avec succès",
			"contractId":  contract.ID,
			"status":      contract.Status,
			"transaction": tx,
			"txId":        txID,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Erreur lors de l'encodage de la réponse: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Répondre avec succès (contrat approuvé mais pas encore exécutable)
	response := map[string]interface{}{
		"success":    true,
		"message":    "Contrat approuvé avec succès",
		"contractId": contract.ID,
		"status":     contract.Status,
		"approvals":  getApprovalCount(contract),
		"required":   contract.RequiredApprovals,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erreur lors de l'encodage de la réponse: "+err.Error(), http.StatusInternalServerError)
	}
}

// cancelContract annule un contrat existant
func cancelContract(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, contractID, username string) {
	// Récupérer le contrat
	contract := getContractByID(bc, contractID)
	if contract == nil {
		http.Error(w, "Contrat non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier si l'utilisateur est autorisé à annuler (créateur ou participant)
	isAuthorized := contract.CreatedBy == username
	if !isAuthorized {
		for _, p := range contract.Participants {
			if p == username {
				isAuthorized = true
				break
			}
		}
	}

	if !isAuthorized {
		utils.LogAuditEvent(
			utils.EventTypeSecurityAlert,
			username,
			r.RemoteAddr,
			fmt.Sprintf("Tentative d'annulation non autorisée du contrat %s", contractID),
			utils.RiskMedium,
			map[string]interface{}{"contract_id": contractID},
		)
		http.Error(w, "Non autorisé à annuler ce contrat", http.StatusForbidden)
		return
	}

	// Annuler le contrat
	err := contract.CancelContract(username)
	if err != nil {
		http.Error(w, "Erreur lors de l'annulation: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Sauvegarder le contrat mis à jour
	if err := updateContract(bc, contract); err != nil {
		http.Error(w, "Erreur lors de la mise à jour du contrat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Journaliser l'action
	utils.LogAuditEvent(
		utils.EventTypeContractCancelled,
		username,
		r.RemoteAddr,
		fmt.Sprintf("Annulation du contrat %s", contractID),
		utils.RiskLow,
		map[string]interface{}{"contract_id": contractID},
	)

	// Répondre avec succès
	response := map[string]interface{}{
		"success":    true,
		"message":    "Contrat annulé avec succès",
		"contractId": contract.ID,
		"status":     contract.Status,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erreur lors de l'encodage de la réponse: "+err.Error(), http.StatusInternalServerError)
	}
}

// executeContract exécute manuellement un contrat
func executeContract(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, contractID, username string) {
	// Récupérer le contrat
	contract := getContractByID(bc, contractID)
	if contract == nil {
		http.Error(w, "Contrat non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier si l'utilisateur est le créateur du contrat
	if contract.CreatedBy != username {
		utils.LogAuditEvent(
			utils.EventTypeSecurityAlert,
			username,
			r.RemoteAddr,
			fmt.Sprintf("Tentative d'exécution non autorisée du contrat %s", contractID),
			utils.RiskMedium,
			map[string]interface{}{"contract_id": contractID},
		)
		http.Error(w, "Seul le créateur peut exécuter ce contrat manuellement", http.StatusForbidden)
		return
	}

	// Vérifier si le contrat peut être exécuté
	if !contract.CanExecute() {
		http.Error(w, "Le contrat ne peut pas être exécuté actuellement", http.StatusBadRequest)
		return
	}

	// Exécuter le contrat
	tx, err := contract.ExecuteContract(bc)
	if err != nil {
		http.Error(w, "Erreur lors de l'exécution: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Mettre à jour le contrat dans la blockchain
	txID, err := bc.UpdateContractToBlockchain(contract)
	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour dans la blockchain: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Journaliser l'action
	utils.LogAuditEvent(
		utils.EventTypeContractExecuted,
		username,
		r.RemoteAddr,
		fmt.Sprintf("Exécution du contrat %s", contractID),
		utils.RiskLow,
		map[string]interface{}{
			"contract_id": contractID,
			"tx_id":       txID,
		},
	)

	// Répondre avec succès
	response := map[string]interface{}{
		"success":       true,
		"message":       "Contrat exécuté avec succès",
		"contractId":    contract.ID,
		"status":        contract.Status,
		"transactionId": txID,
		"transaction":   tx,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erreur lors de l'encodage de la réponse: "+err.Error(), http.StatusInternalServerError)
	}
}

// ------ Fonctions auxiliaires ------

// getContractType convertit une chaîne en type de contrat
func getContractType(typeStr string) (blockchain.ContractType, error) {
	switch typeStr {
	case "TRANSFER":
		return blockchain.ContractTransfer, nil
	case "MULTISIG":
		return blockchain.ContractMultiSig, nil
	case "TIMELOCK":
		return blockchain.ContractTimeLock, nil
	case "CONDITIONAL":
		return blockchain.ContractCondition, nil
	case "ESCROW":
		return blockchain.ContractEscrow, nil
	default:
		return "", fmt.Errorf("type de contrat non reconnu: %s", typeStr)
	}
}

// getContractsForUser récupère les contrats associés à un utilisateur
func getContractsForUser(bc *blockchain.Blockchain, username string) []*blockchain.SmartContract {
	contracts := bc.GetContractsForUser(username)
	if contracts == nil {
		// Si pas encore implémenté dans la blockchain, retourner un tableau vide
		return []*blockchain.SmartContract{}
	}
	return contracts
}

// getContractByID récupère un contrat par son ID
func getContractByID(bc *blockchain.Blockchain, contractID string) *blockchain.SmartContract {
	contract := bc.GetContractByID(contractID)
	return contract
}

// hasAccessToContract vérifie si un utilisateur a accès à un contrat
func hasAccessToContract(contract *blockchain.SmartContract, username string) bool {
	// Le créateur a toujours accès
	if contract.CreatedBy == username {
		return true
	}

	// Vérifier si l'utilisateur est un participant
	for _, participant := range contract.Participants {
		if participant == username {
			return true
		}
	}

	// Vérifier si l'utilisateur est le destinataire
	if contract.Recipient == username {
		return true
	}

	return false
}

// saveContract sauvegarde un nouveau contrat
func saveContract(bc *blockchain.Blockchain, contract *blockchain.SmartContract) error {
	// Dans une implémentation réelle, vous sauvegarderiez le contrat dans la blockchain
	// Pour l'instant, nous simulons cette opération
	return bc.SaveContract(contract)
}

// updateContract met à jour un contrat existant
func updateContract(bc *blockchain.Blockchain, contract *blockchain.SmartContract) error {
	// Dans une implémentation réelle, vous mettriez à jour le contrat dans la blockchain
	// Pour l'instant, nous simulons cette opération
	return bc.UpdateContract(contract)
}

// getApprovalCount compte le nombre d'approbations reçues
func getApprovalCount(contract *blockchain.SmartContract) int {
	count := 0
	for _, approved := range contract.Approvals {
		if approved {
			count++
		}
	}
	return count
}

// ContractUIHandler gère l'affichage de l'interface utilisateur des contrats
func ContractUIHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)

		// Vérification de l'authentification
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		mu.Lock()
		session, exists := sessions[cookie.Value]
		mu.Unlock()
		if !exists || session == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Journaliser l'accès
		utils.LogAuditEvent(
			utils.EventTypeUIAccess,
			session.Username,
			r.RemoteAddr,
			"Accès à l'interface des contrats",
			utils.RiskLow,
			nil,
		)

		// Traitement des demandes spécifiques à un contrat individuel
		// Format URL: /contract/{contract_id}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 2 && parts[1] == "contract" && parts[2] != "" {
			contractID := parts[2]
			renderContractDetail(w, r, bc, contractID, session.Username)
			return
		}

		// Récupérer les contrats de l'utilisateur
		contracts := getContractsForUser(bc, session.Username)

		// Récupérer les templates
		tmpl, err := template.New("contracts.html").Funcs(template.FuncMap{
			"isParticipant": func(participants []string, username string) bool {
				for _, p := range participants {
					if p == username {
						return true
					}
				}
				return false
			},
			"hasApproved": func(approvals map[string]bool, username string) bool {
				approved, exists := approvals[username]
				return exists && approved
			},
			"slice": func(s string, i, j int) string {
				if i < 0 || i >= len(s) || j > len(s) {
					return s
				}
				return s[i:j]
			},
		}).ParseFiles("templates/contracts.html")

		if err != nil {
			http.Error(w, "Erreur lors du chargement du template: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Contracts": contracts,
			"Username":  session.Username,
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Erreur lors du rendu du template: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// renderContractDetail affiche les détails d'un contrat spécifique
func renderContractDetail(w http.ResponseWriter, r *http.Request, bc *blockchain.Blockchain, contractID, username string) {
	// Récupérer le contrat
	contract := getContractByID(bc, contractID)
	if contract == nil {
		http.Error(w, "Contrat non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier si l'utilisateur a accès à ce contrat
	if !hasAccessToContract(contract, username) {
		// Journaliser la tentative d'accès non autorisée
		utils.LogAuditEvent(
			utils.EventTypeSecurityAlert,
			username,
			r.RemoteAddr,
			fmt.Sprintf("Tentative d'accès non autorisé à l'interface du contrat %s", contractID),
			utils.RiskMedium,
			map[string]interface{}{"contract_id": contractID},
		)
		http.Error(w, "Non autorisé à accéder à ce contrat", http.StatusForbidden)
		return
	}

	// Charger les templates avec leurs fonctions
	tmpl, err := template.New("contract_detail.html").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("02/01/2006 15:04:05")
		},
		"getApprovalCount": getApprovalCount,
		"isApproved": func(approvals map[string]bool, user string) bool {
			approved, exists := approvals[user]
			return exists && approved
		},
	}).ParseFiles("templates/contract_detail.html")

	if err != nil {
		http.Error(w, "Erreur lors du chargement du template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Exécuter le template
	data := map[string]interface{}{
		"Contract": contract,
		"Username": username,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Erreur lors du rendu du template: "+err.Error(), http.StatusInternalServerError)
	}

	// Journaliser la visualisation du contrat
	utils.LogAuditEvent(
		utils.EventTypeContractViewed,
		username,
		r.RemoteAddr,
		fmt.Sprintf("Visualisation du contrat %s via l'interface", contractID),
		utils.RiskLow,
		map[string]interface{}{"contract_id": contractID},
	)
}
