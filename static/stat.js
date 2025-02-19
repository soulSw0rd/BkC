
document.addEventListener('DOMContentLoaded', () => {
    console.log('📊 CryptoChain Go - Statistiques Loaded');

    // Mise à jour des statistiques toutes les 5 secondes
    function updateStats() {
        fetch('/stats')
            .then(response => response.json())
            .then(data => {
                document.getElementById('visitor-count').innerText = data.visitorCount;
                document.getElementById('active-sessions').innerText = data.activeSessionCount;
                document.getElementById('daily-transactions').innerText = data.dailyTransactions;
                document.getElementById('last-block').innerText = JSON.stringify(data.lastBlock, null, 2);
                
                updateChart(data.transactionsByDay);
            })
            .catch(error => console.error('❌ Erreur chargement stats:', error));
    }

    // Initialisation des données
    updateStats();
    setInterval(updateStats, 5000);

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
        transactionsChart.data.labels = Object.keys(data);
        transactionsChart.data.datasets[0].data = Object.values(data);
        transactionsChart.update();
    }
});
