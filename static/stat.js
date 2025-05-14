document.addEventListener('DOMContentLoaded', () => {
    console.log('📊 CryptoChain Go - Statistiques Loaded');

    // Initialisation des données
    updateStats();
    
    // Mise à jour régulière en cas d'échec WebSocket
    const statsInterval = setInterval(updateStats, 5000);

    // Établir une connexion WebSocket pour les mises à jour en temps réel
    let ws;
    function connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
        const wsUrl = `${protocol}${window.location.host}/ws`;
        
        ws = new WebSocket(wsUrl);
        
        ws.onopen = function() {
            console.log('Connexion WebSocket établie pour les statistiques');
        };
        
        ws.onmessage = function(event) {
            const message = JSON.parse(event.data);
            
            if (message.type === 'blockchain_update') {
                // Une mise à jour de la blockchain a eu lieu, rafraîchir les statistiques
                updateStats();
                
                // Animation subtile pour indiquer une mise à jour
                document.getElementById('daily-transactions').classList.add('text-purple-400');
                setTimeout(() => {
                    document.getElementById('daily-transactions').classList.remove('text-purple-400');
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
    
    // Connexion au WebSocket
    connectWebSocket();

    // Mise à jour des statistiques via API
    function updateStats() {
        fetch('/stats?format=json')
            .then(response => response.json())
            .then(data => {
                document.getElementById('visitor-count').innerText = data.visitorCount;
                document.getElementById('active-sessions').innerText = data.activeSessions;
                document.getElementById('daily-transactions').innerText = data.dailyTransactions;
                document.getElementById('last-block').innerText = JSON.stringify(data.lastBlock, null, 2);
                
                // Mise à jour du tableau des connexions
                updateConnectionsTable(data.recentConnections);
                
                // Mise à jour du graphique
                updateChart(data.transactionHistory);
            })
            .catch(error => console.error('❌ Erreur chargement stats:', error));
    }

    // Mise à jour du tableau des connexions IP
    function updateConnectionsTable(connections) {
        const connectionsTable = document.getElementById('connections-table');
        if (!connectionsTable || !connections || connections.length === 0) {
            return;
        }
        
        // Vider le tableau existant
        connectionsTable.innerHTML = '';
        
        // Ajouter les nouvelles lignes
        connections.forEach(conn => {
            const row = document.createElement('tr');
            
            // Nom d'utilisateur
            const usernameCell = document.createElement('td');
            usernameCell.className = 'px-6 py-4 whitespace-nowrap font-medium';
            usernameCell.textContent = conn.username;
            
            // Heure de dernière activité
            const timeCell = document.createElement('td');
            timeCell.className = 'px-6 py-4 whitespace-nowrap';
            
            // Calculer le temps écoulé depuis la dernière activité
            const lastActivity = new Date(conn.timestamp);
            const timeElapsed = getTimeElapsed(lastActivity);
            timeCell.textContent = timeElapsed;
            
            // Pays
            const countryCell = document.createElement('td');
            countryCell.className = 'px-6 py-4 whitespace-nowrap';
            
            // Afficher un drapeau en fonction du code pays
            const flag = getFlagEmoji(conn.country);
            countryCell.textContent = `${flag} ${conn.country}`;
            
            // Navigateur
            const browserCell = document.createElement('td');
            browserCell.className = 'px-6 py-4 whitespace-nowrap';
            browserCell.textContent = conn.userAgent;
            
            // Statut
            const statusCell = document.createElement('td');
            statusCell.className = 'px-6 py-4 whitespace-nowrap';
            
            // Définir la couleur en fonction du statut
            let statusColor = 'text-gray-400';
            if (conn.lastAction === 'Actif') {
                statusColor = 'text-green-400';
            } else if (conn.lastAction === 'Inactif') {
                statusColor = 'text-yellow-400';
            } else if (conn.lastAction === 'Déconnecté') {
                statusColor = 'text-red-400';
            }
            
            statusCell.className = `px-6 py-4 whitespace-nowrap ${statusColor}`;
            statusCell.textContent = conn.lastAction;
            
            // Ajouter les cellules à la ligne
            row.appendChild(usernameCell);
            row.appendChild(timeCell);
            row.appendChild(countryCell);
            row.appendChild(browserCell);
            row.appendChild(statusCell);
            
            // Ajouter la ligne au tableau
            connectionsTable.appendChild(row);
        });
    }
    
    // Récupérer le temps écoulé depuis une date au format humain
    function getTimeElapsed(date) {
        const now = new Date();
        const elapsed = now - date;
        
        const seconds = Math.floor(elapsed / 1000);
        if (seconds < 60) {
            return 'Il y a quelques secondes';
        }
        
        const minutes = Math.floor(seconds / 60);
        if (minutes < 60) {
            return `Il y a ${minutes} minute${minutes > 1 ? 's' : ''}`;
        }
        
        const hours = Math.floor(minutes / 60);
        if (hours < 24) {
            return `Il y a ${hours} heure${hours > 1 ? 's' : ''}`;
        }
        
        // Pour les dates plus anciennes, afficher la date complète
        return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`;
    }
    
    // Récupérer l'emoji du drapeau à partir du code pays
    function getFlagEmoji(countryCode) {
        if (!countryCode || countryCode === '??' || countryCode.length !== 2) {
            return '🌐'; // Globe pour les pays inconnus
        }
        
        // Convertir le code pays en caractères régionaux Unicode
        // Les codes régionaux sont à 127397 caractères de leurs équivalents ASCII
        const codePoints = countryCode
            .toUpperCase()
            .split('')
            .map(char => 127397 + char.charCodeAt(0));
            
        return String.fromCodePoint(...codePoints);
    }

    // Configuration du graphique Chart.js
    const ctx = document.getElementById('transactionsChart').getContext('2d');
    let transactionsChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: [], // Jours
            datasets: [{
                label: 'Transactions',
                data: [], // Nombre de transactions
                borderColor: 'rgba(93, 31, 142, 1)',
                backgroundColor: 'rgba(93, 31, 142, 0.2)',
                borderWidth: 2,
                tension: 0.3
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false
        }
    });

    // Mise à jour du graphique
    function updateChart(data) {
        if (data && data.dates && data.counts) {
            transactionsChart.data.labels = data.dates;
            transactionsChart.data.datasets[0].data = data.counts;
            transactionsChart.update();
        }
    }
});
