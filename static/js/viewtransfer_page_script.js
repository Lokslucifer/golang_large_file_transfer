
const BACKEND_BASE = 'http://localhost:8081/api';
let transfers = [];
let filteredTransfers = [];
let currentEditingId = null;

// Check authentication
function checkAuth() {
    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return false;
    }
    return token;
}

// Logout function
function logout() {
    localStorage.removeItem('auth_token');
    window.location.href = '/';
}

// Format file size
function formatFileSize(bytes) {
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
}

// Format date
function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

// Calculate expiry status
function getTransferStatus(createdAt, expiry) {
    const created = new Date(createdAt);
    const now = new Date();
    const expiryMs = {
        '1h': 60 * 60 * 1000,
        '24h': 24 * 60 * 60 * 1000,
        '7d': 7 * 24 * 60 * 60 * 1000,
        '30d': 30 * 24 * 60 * 60 * 1000
    };
    
    const expiresAt = new Date(created.getTime() + expiryMs[expiry]);
    return now > expiresAt ? 'expired' : 'active';
}

// Load transfers from backend
async function loadTransfers() {
    const token = checkAuth();
    if (!token) return;

    try {
        const response = await fetch(`${BACKEND_BASE}/auth/transfer/all`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        console.log(response,"-response")

        if (response.ok) {
            const data = await response.json();
            transfers = data.data || [];
            console.log("transfers-",transfers)
            
            // Add mock data for demonstration
            if (transfers.length === 0) {
                transfers =[];
            }
            
            filteredTransfers = [...transfers];
            displayTransfers();
        } else {
            throw new Error('Failed to load transfers');
        }
    } catch (error) {
        console.error('Error loading transfers:', error);
        showToast('Error loading transfers', 'error');
  
    }

    document.getElementById('loadingState').style.display = 'none';
}

function FindExpiry(expiryTimeStr) {
    // Ensure expiryTime is in milliseconds
    const expiryTime = new Date(expiryTimeStr).getTime(); 
    if (expiryTime < 1e12) expiryTime *= 1000;
    if (expiryTime < Date.now()) return 'Expired';
    console.log("expiryTime-",expiryTime)

    const now = Date.now();
    const diff = expiryTime - now;
    console.log("diff-",diff)

    if (diff <= 0) return '0';

    const minute = 60 * 1000;
    const hour = 60 * minute;
    const day = 24 * hour;
    const week = 7 * day;
    const month = 30 * day;

    if (diff >= month) return `${Math.floor(diff / month)} Month${Math.floor(diff / month) > 1 ? 's' : ''}`;
    if (diff >= week)  return `${Math.floor(diff / week)} Week${Math.floor(diff / week) > 1 ? 's' : ''}`;
    if (diff >= day)   return `${Math.floor(diff / day)} Day${Math.floor(diff / day) > 1 ? 's' : ''}`;
    if (diff >= hour)  return `${Math.floor(diff / hour)} Hour${Math.floor(diff / hour) > 1 ? 's' : ''}`;
    return `${Math.floor(diff / minute)} Minute${Math.floor(diff / minute) !== 1 ? 's' : ''}`;
}

// Display transfers
function displayTransfers() {
    const grid = document.getElementById('transfersGrid');
    const emptyState = document.getElementById('emptyState');

    if (filteredTransfers.length === 0) {
        grid.style.display = 'none';
        emptyState.style.display = 'block';
        return;
    }

    emptyState.style.display = 'none';
    grid.style.display = 'grid';

    grid.innerHTML = filteredTransfers.map(transfer => {
        const status = getTransferStatus(transfer.created_at, transfer.expiry);
        // const shareUrl = `/share/${transferId}`;
        const shareLink = `${window.location.origin}/share/${transfer.id}`;
        
        return `
            <div class="transfer-card">
                <div class="transfer-header">
                    <div>
                        <div class="transfer-title">Transfer #${transfer.id}</div>
                        <div class="transfer-date">${formatDate(transfer.created_at)}</div>
                    </div>
                    <div class="transfer-status status-${status}">${status}</div>
                </div>

                ${transfer.message ? `
                    <div class="transfer-message">
                        💬 ${transfer.message}
                    </div>
                ` : ''}

                <div class="transfer-info">
                    <div class="info-item">
                        <div class="info-label">Size</div>
                        <div class="info-value">${formatFileSize(transfer.size)}</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">Files</div>
                        <div class="info-value">0 files</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">Downloads</div>
                        <div class="info-value">${transfer.download_count} times</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">Expires</div>
                        <div class="info-value">${FindExpiry(transfer.expiry)}</div>
                    </div>
                </div>

                <div class="transfer-actions">
                    <button class="btn btn-primary btn-sm" onclick="copyShareLink('${shareLink}')">
                        🔗 Copy Link
                    </button>
                    <button class="btn btn-outline btn-sm" onclick="downloadTransfer('${transfer.id}')">
                        ⬇️ Download
                    </button>
                    <button class="btn btn-outline btn-sm" onclick="editTransfer('${transfer.id}')">
                        ✏️ Edit
                    </button>
                    <button class="btn btn-danger btn-sm" onclick="deleteTransfer('${transfer.id}')">
                        🗑️ Delete
                    </button>
                </div>
            </div>
        `;
    }).join('');
}

// Filter transfers
function filterTransfers() {
    const searchTerm = document.getElementById('searchInput').value.toLowerCase();
    const statusFilter = document.getElementById('statusFilter').value;
    const dateFilter = document.getElementById('dateFilter').value;

    filteredTransfers = transfers.filter(transfer => {
        // Search filter
        const matchesSearch = !searchTerm || 
            transfer.message.toLowerCase().includes(searchTerm) ||
            transfer.id.toLowerCase().includes(searchTerm);

        // Status filter
        const transferStatus = getTransferStatus(transfer.created_at, transfer.expiry);
        const matchesStatus = !statusFilter || transferStatus === statusFilter;

        // Date filter
        let matchesDate = true;
        if (dateFilter) {
            const transferDate = new Date(transfer.created_at);
            const now = new Date();
            
            switch (dateFilter) {
                case 'today':
                    matchesDate = transferDate.toDateString() === now.toDateString();
                    break;
                case 'week':
                    const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
                    matchesDate = transferDate >= weekAgo;
                    break;
                case 'month':
                    const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
                    matchesDate = transferDate >= monthAgo;
                    break;
            }
        }

        return matchesSearch && matchesStatus && matchesDate;
    });

    displayTransfers();
}

// Copy share link
function copyShareLink(link) {
    navigator.clipboard.writeText(link).then(() => {
        showToast('Share link copied to clipboard!');
    });
}

// Download transfer
async function downloadTransfer(transferId) {
    const token = checkAuth();
    if (!token) return;

    try {
        // In a real implementation, this would download the actual files
        showToast('Download started for transfer #' + transferId);
        
        // Simulate download
        window.open(`${BACKEND_BASE}/transfer/download/transfer/${transferId}`, '_blank');
    } catch (error) {
        showToast('Download failed: ' + error.message, 'error');
    }
}

// Edit transfer
function editTransfer(transferId) {
    const transfer = transfers.find(t => t.id === transferId);
    if (!transfer) return;

    currentEditingId = transferId;
    document.getElementById('editMessage').value = transfer.message || '';
    document.getElementById('editExpiry').value = transfer.expiry;
    document.getElementById('editModal').classList.add('active');
}

// Close edit modal
function closeEditModal() {
    document.getElementById('editModal').classList.remove('active');
    currentEditingId = null;
}

// Handle edit form submission
document.getElementById('editForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    if (!currentEditingId) return;

    const token = checkAuth();
    if (!token) return;

    const formData = {
        "transfer_id":currentEditingId,
        message: document.getElementById('editMessage').value,
        expiry: document.getElementById('editExpiry').value
    };
    console.log(formData)

    try {
        const response = await fetch(`${BACKEND_BASE}/auth/transfer/update`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(formData)
        });

        if (response.ok) {
            // Update local data
            const transferIndex = transfers.findIndex(t => t.id === currentEditingId);
            if (transferIndex !== -1) {
                transfers[transferIndex] = { ...transfers[transferIndex], ...formData };
                filterTransfers();
            }
            
            showToast('Transfer updated successfully!');
            closeEditModal();
        } else {
            throw new Error('Failed to update transfer');
        }
    } catch (error) {
        showToast('Update failed: ' + error.message, 'error');
    }
});

// Delete transfer
async function deleteTransfer(transferId) {
    if (!confirm('Are you sure you want to delete this transfer? This action cannot be undone.')) {
        return;
    }

    const token = checkAuth();
    if (!token) return;

    try {
        const response = await fetch(`${BACKEND_BASE}/auth/transfer/delete/${transferId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            // Remove from local data
            transfers = transfers.filter(t => t.id !== transferId);
            filterTransfers();
            showToast('Transfer deleted successfully!');
        } else {
            throw new Error('Failed to delete transfer');
        }
    } catch (error) {
        showToast('Delete failed: ' + error.message, 'error');
    }
}

// Show toast notification
function showToast(message, type = 'success') {
    // Simple toast implementation
    const toast = document.createElement('div');
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: ${type === 'error' ? '#ef4444' : '#10b981'};
        color: white;
        padding: 12px 20px;
        border-radius: 8px;
        z-index: 10000;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    `;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(() => {
        toast.remove();
    }, 3000);
}

// Event listeners
document.getElementById('searchInput').addEventListener('input', filterTransfers);
document.getElementById('statusFilter').addEventListener('change', filterTransfers);
document.getElementById('dateFilter').addEventListener('change', filterTransfers);

// Close modal when clicking outside
document.getElementById('editModal').addEventListener('click', (e) => {
    if (e.target === e.currentTarget) {
        closeEditModal();
    }
});

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadTransfers();
});
