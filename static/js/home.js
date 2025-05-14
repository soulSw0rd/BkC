document.addEventListener('DOMContentLoaded', () => {
    console.log('🏠 CryptoChain Go - Page d\'accueil chargée');
    
    // Éléments DOM pour les statistiques
    const blockCountElement = document.getElementById('block-count');
    const userCountElement = document.getElementById('user-count');
    const txCountElement = document.getElementById('tx-count');

    // Récupérer les statistiques en temps réel
    function updateStats() {
        fetch('/stats?format=json')
            .then(response => response.json())
            .then(data => {
                blockCountElement.textContent = data.dailyTransactions + 1; // +1 pour inclure le bloc genesis
                userCountElement.textContent = data.activeSessions;
                txCountElement.textContent = data.dailyTransactions;
            })
            .catch(error => console.error('❌ Erreur chargement stats:', error));
    }

    // Initialiser les statistiques et configurer la mise à jour périodique
    updateStats();
    setInterval(updateStats, 5000);

    // Ajouter une section pour miner des blocs
    const heroSection = document.querySelector('.h-screen');
    
    const miningSection = document.createElement('div');
    miningSection.className = 'mt-6 p-5 bg-gray-800 rounded-lg max-w-lg mx-auto';
    miningSection.innerHTML = `
        <h3 class="text-2xl font-bold mb-3">⛏️ Miner un nouveau bloc</h3>
        <div class="flex flex-col space-y-3">
            <input type="text" id="mining-data" placeholder="Vos données à ajouter à la blockchain" 
                class="p-2 rounded bg-gray-700 text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500">
            <button id="mine-button" class="bg-purple-500 hover:bg-purple-600 text-white font-bold py-2 px-4 rounded transition">
                ⛏️ Miner un bloc
            </button>
            <div id="mining-status" class="hidden mt-2 p-2 text-center rounded"></div>
        </div>
    `;
    
    heroSection.appendChild(miningSection);

    // Gérer le clic sur le bouton de minage
    document.getElementById('mine-button').addEventListener('click', async () => {
        const dataInput = document.getElementById('mining-data');
        const miningStatus = document.getElementById('mining-status');
        const mineButton = document.getElementById('mine-button');
        const data = dataInput.value.trim();
        
        if (!data) {
            miningStatus.textContent = '⚠️ Veuillez entrer des données à miner';
            miningStatus.className = 'mt-2 p-2 text-center rounded bg-yellow-500 text-white';
            miningStatus.classList.remove('hidden');
            return;
        }
        
        // Désactiver le bouton pendant le minage
        mineButton.disabled = true;
        mineButton.textContent = '⏳ Minage en cours...';
        
        // Afficher le status
        miningStatus.textContent = '⏳ Le bloc est en cours de minage...';
        miningStatus.className = 'mt-2 p-2 text-center rounded bg-blue-500 text-white';
        miningStatus.classList.remove('hidden');
        
        try {
            const response = await fetch('/mine-block', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ data: data })
            });
            
            const result = await response.json();
            
            if (response.ok) {
                miningStatus.textContent = '✅ Bloc miné avec succès! Il sera ajouté à la blockchain sous peu.';
                miningStatus.className = 'mt-2 p-2 text-center rounded bg-green-500 text-white';
                dataInput.value = '';
                
                // Mettre à jour les stats après un court délai
                setTimeout(updateStats, 1000);
            } else {
                miningStatus.textContent = `❌ Erreur: ${result.message || 'Une erreur s\'est produite'}`;
                miningStatus.className = 'mt-2 p-2 text-center rounded bg-red-500 text-white';
            }
        } catch (error) {
            miningStatus.textContent = '❌ Erreur de connexion au serveur';
            miningStatus.className = 'mt-2 p-2 text-center rounded bg-red-500 text-white';
            console.error('Erreur:', error);
        } finally {
            // Réactiver le bouton
            mineButton.disabled = false;
            mineButton.textContent = '⛏️ Miner un bloc';
        }
    });
    
    // Configurer les WebSockets pour les mises à jour en temps réel
    function connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
        const wsUrl = `${protocol}${window.location.host}/ws`;
        
        const ws = new WebSocket(wsUrl);
        
        ws.onopen = function() {
            console.log('Connexion WebSocket établie');
        };
        
        ws.onmessage = function(event) {
            const message = JSON.parse(event.data);
            
            if (message.type === 'blockchain_update') {
                updateStats();
                
                // Animation subtile pour montrer une mise à jour
                blockCountElement.classList.add('text-purple-400');
                setTimeout(() => {
                    blockCountElement.classList.remove('text-purple-400');
                }, 800);
            }
        };
        
        ws.onclose = function() {
            console.log('Connexion WebSocket fermée');
            // Tentative de reconnexion après un délai
            setTimeout(connectWebSocket, 5000);
        };
        
        ws.onerror = function(error) {
            console.error('Erreur WebSocket:', error);
        };
    }
    
    // Connecter le WebSocket
    connectWebSocket();
}); 