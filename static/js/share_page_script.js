
const BACKEND_BASE = 'http://localhost:8081/api';

const SHARE_BACKEND_URL=BACKEND_BASE+"/transfer/share"
const FILE_DOWNLOAD_BACKEND_URL=BACKEND_BASE+"/transfer/download/file"
const TRANSFER_DOWNLOAD_BACKEND_URL=BACKEND_BASE+"/transfer/download/transfer"


// DOM elements
const loadingState = document.getElementById('loading-state');
const errorState = document.getElementById('error-state');
const contentState = document.getElementById('content-state');
const statusBadge = document.getElementById('status-badge');
const transferIdEl = document.getElementById('transfer-id');
const totalSizeEl = document.getElementById('total-size');
const expiresAtEl = document.getElementById('expires-at');
const messageSection = document.getElementById('message-section');
const messageContent = document.getElementById('message-content');
const filesContainer = document.getElementById('files-container');
const filesCount = document.getElementById('files-count');
const downloadAllBtn = document.getElementById('download-all-btn');

// Utility functions
function formatFileSize(bytes) {
    if (bytes < 1024 * 1024) {
        return (bytes / 1024).toFixed(1) + ' KB';
    } else if (bytes < 1024 * 1024 * 1024) {
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    } else {
        return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
    }
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' at ' + date.toLocaleTimeString();
}

function getFileExtension(filename) {
    return filename.split('.').pop().toLowerCase();
}

function getFileIcon(extension) {
    const iconMap = {
        'pdf': '📄',
        'doc': '📝', 'docx': '📝',
        'xls': '📊', 'xlsx': '📊',
        'ppt': '📈', 'pptx': '📈',
        'jpg': '🖼️', 'jpeg': '🖼️', 'png': '🖼️', 'gif': '🖼️',
        'mp4': '🎥', 'avi': '🎥', 'mov': '🎥',
        'mp3': '🎵', 'wav': '🎵',
        'zip': '🗜️', 'rar': '🗜️',
        'txt': '📃'
    };
    return iconMap[extension] || '📄';
}

// API functions
async function fetchTransferInfo() {
    try {
        console.log("sending request")
        const response = await fetch(`${SHARE_BACKEND_URL}/${transferId}`);
        const data = await response.json();
        console.log(data,"-data received")
        console.log(response.ok)
        
        if (!response.ok) {
            throw new Error(data.error?.message || 'Failed to fetch transfer');
        }
        
        return data.data;
    } catch (error) {
        console.error('Error fetching transfer:', error);
        throw error;
    }
}

function downloadFile(fileId, filename) {
    const downloadUrl = `${FILE_DOWNLOAD_BACKEND_URL}/${fileId}`;
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = filename;
    link.target = '_blank';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

function downloadAllFiles() {
    const downloadUrl = `${TRANSFER_DOWNLOAD_BACKEND_URL}/${transferId}`;
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.target = '_blank';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

// UI functions
function showError(message = 'Transfer not found or expired') {
    loadingState.style.display = 'none';
    contentState.style.display = 'none';
    errorState.style.display = 'block';
    errorState.querySelector('.error-message').textContent = message;
}

function showContent(transferData) {
    loadingState.style.display = 'none';
    errorState.style.display = 'none';
    contentState.style.display = 'block';

    // Update transfer info
    transferIdEl.textContent = transferData.id;
    totalSizeEl.textContent = formatFileSize(transferData.size);
    expiresAtEl.textContent = formatDate(transferData.expiry);

    // Check if expired
    const now = new Date();
    const expiry = new Date(transferData.expiry);
    if (expiry < now) {
        statusBadge.textContent = 'Expired';
        statusBadge.className = 'expired-badge';
        downloadAllBtn.disabled = true;
        downloadAllBtn.textContent = 'Transfer Expired';
        downloadAllBtn.style.opacity = '0.5';
        downloadAllBtn.style.cursor = 'not-allowed';
    }

    // Show message if exists
    if (transferData.message && transferData.message.trim()) {
        messageSection.style.display = 'block';
        messageContent.textContent = transferData.message;
    }

    // Update files count
    const fileCount = transferData.file_info_list?.length || 0;
    filesCount.textContent = `${fileCount} file${fileCount !== 1 ? 's' : ''}`;

    // Render files
    renderFiles(transferData.file_info_list || []);
}

function renderFiles(files) {
    filesContainer.innerHTML = '';

    if (files.length === 0) {
        filesContainer.innerHTML = '<p style="text-align: center; color: #6b7280; padding: 40px;">No files available</p>';
        return;
    }

    files.forEach(file => {
        const fileItem = document.createElement('div');
        fileItem.className = 'file-item';
        
        const extension = getFileExtension(file.file_name);
        const icon = getFileIcon(extension);
        
        fileItem.innerHTML = `
            <div class="file-info">
                <div class="file-name">${icon} ${file.file_name}</div>
                <div class="file-meta">${formatFileSize(file.file_size)} • ${extension.toUpperCase()}</div>
            </div>
            <button class="download-btn" onclick="downloadFile('${file.id}', '${file.file_name}')">
                <svg class="icon" viewBox="0 0 20 20">
                    <path d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z"/>
                </svg>
                Download
            </button>
        `;
        
        filesContainer.appendChild(fileItem);
    });
}

// Event listeners
downloadAllBtn.addEventListener('click', downloadAllFiles);

// Make functions globally available
window.downloadFile = downloadFile;

// Initialize
async function init() {
    if (!transferId) {
        showError('No transfer ID provided');
        return;
    }

    try {
        console.log("fetching data-",transferId)
        const transferData = await fetchTransferInfo();
        showContent(transferData);
    } catch (error) {
        showError(error.message);
    }
}

// Start the app
init();
