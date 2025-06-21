
const BACKEND_BASE = 'http://localhost:8081/api';
const ENDPOINTS = {
    NEW_TRANSFER: `${BACKEND_BASE}/auth/transfer/new`,
    UPLOAD_CHUNK: `${BACKEND_BASE}/auth/transfer/upload`,
    ASSEMBLE: `${BACKEND_BASE}/auth/transfer/assemble`,
    CANCEL: `${BACKEND_BASE}/auth/transfer/cancel`,
    LOGIN: `${BACKEND_BASE}/login`,
    SIGNUP: `${BACKEND_BASE}/signup`,
};

// Global state
let authToken = localStorage.getItem('auth_token');
let isLoginMode = true;
let selectedFiles = [];
let totalSize = 0;
let transferId = null;
let zippedFile = null;
let totalChunks = 0;
let maxChunkSize = 0;
let currentChunk = 0;
let isPaused = false;
let uploadInProgress = false;

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    if (authToken) {
        showUploadCard();
    } else {
        showAboutCard();
    }
});

// Authentication functions
function switchAuthMode(mode) {
    isLoginMode = mode === 'login';
    if(isLoginMode){
        document.getElementById("login-form").classList.remove("hidden");
        document.getElementById("signup-form").classList.add("hidden");
    }else{
        document.getElementById("login-form").classList.add("hidden");
        document.getElementById("signup-form").classList.remove("hidden");
    }

    document.getElementById('login-tab').classList.toggle('active', isLoginMode);
    document.getElementById('signup-tab').classList.toggle('active', !isLoginMode);


}

async function handleLogin(event) {
    event.preventDefault(); // prevent page reload

    const form = document.getElementById('login-form');
    const formData = new FormData(form);

    const email = formData.get('email');
    const password = formData.get('password');
    // Example body object
    const body = {
        email: email,
        password: password
    };
    
    console.log(body,"-",isLoginMode)
    try {
        const response = await fetch(ENDPOINTS.LOGIN, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        });

        const data = await response.json();
        
        if (response.ok) {
            authToken = data.Token || data.ID;
            localStorage.setItem('auth_token', authToken);
            showToast('success', 'Logged in successfully!');
            setTimeout(() => showUploadCard(), 1000);
        } else {
            throw new Error(data.error?.message || 'Authentication failed');
        }
    } catch (error) {
        showToast('error', error.message);
    }
}

async function handleSignup(event) {
    event.preventDefault(); // prevent page reload

    const form = document.getElementById('signup-form');
    const formData = new FormData(form);

    const email = formData.get('email');
    const password = formData.get('password');
    const first_name = formData.get('firstName');
    const last_name = formData.get('lastName');
    
    // Example body object
    const body = {
        email: email,
        password: password,
        first_name: first_name,
        last_name: last_name
    };
    

    try {
        const response = await fetch(ENDPOINTS.SIGNUP, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        });

        const data = await response.json();
        
        if (response.ok) {
            authToken = data.Token || data.ID;
            localStorage.setItem('auth_token', authToken);
            showToast('success', 'Account Created successfully!');
            setTimeout(() => showUploadCard(), 1000);
        } else {
            throw new Error(data.error?.message || 'Authentication failed');
        }
    } catch (error) {
        showToast('error', error.message);
    }
}


function logout() {
    authToken = null;
    localStorage.removeItem('auth_token');
    showAboutCard();
}
function showAboutCard(){
    document.getElementById('about-card').classList.remove('hidden');
    document.getElementById('loginbutton').classList.remove('hidden');

    document.getElementById('upload-card').classList.add('hidden');
    document.getElementById('auth-card').classList.add('hidden');
    document.getElementById('logoutbutton').classList.add('hidden');
    document.getElementById('viewtransferbutton').classList.add('hidden');   


}
function showAuthCard() {
    document.getElementById('auth-card').classList.remove('hidden');

    document.getElementById('upload-card').classList.add('hidden');
    document.getElementById('about-card').classList.add('hidden');
    document.getElementById('viewtransferbutton').classList.add('hidden');
    document.getElementById('loginbutton').classList.add('hidden');
    document.getElementById('logoutbutton').classList.add('hidden');
 
}

function showUploadCard() {
    document.getElementById('upload-card').classList.remove('hidden');
    document.getElementById('logoutbutton').classList.remove('hidden');
    document.getElementById('viewtransferbutton').classList.remove('hidden');

    document.getElementById('about-card').classList.add('hidden');
    document.getElementById('auth-card').classList.add('hidden');
    document.getElementById('loginbutton').classList.add('hidden');



}

// Show toast notification
function showToast(type,message) {
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


// File handling functions
function handleDrop(event) {
    event.preventDefault();
    const files = Array.from(event.dataTransfer.files);
    addFiles(files);
    document.getElementById('drop-zone').classList.remove('drag-over');
}

function handleDragOver(event) {
    event.preventDefault();
    document.getElementById('drop-zone').classList.add('drag-over');
}

function handleDragLeave(event) {
    event.preventDefault();
    document.getElementById('drop-zone').classList.remove('drag-over');
}

function selectFiles() {
    document.getElementById('file-input').click();
}

function selectFolder() {
    document.getElementById('folder-input').click();
}

function handleFileSelect(event) {
    const files = Array.from(event.target.files);
    addFiles(files);
}

async function handleFolderSelect(event) {
    const files = Array.from(event.target.files);
    if (files.length === 0) return;

    showLoader();
    
    try {
        const folderName = files[0].webkitRelativePath.split('/')[0];
        const zip = new JSZip();

        for (const file of files) {
            const relativePath = file.webkitRelativePath.replace(folderName + "/", "");
            zip.file(relativePath, file);
        }

        const zippedBlob = await zip.generateAsync({ type: 'blob' });
        const zippedFile = new File([zippedBlob], `${folderName}.zip`, { type: 'application/zip' });
        addFiles([zippedFile]);
    } catch (error) {
        showToast('error', 'Failed to process folder: ' + error.message);
    }
    
    hideLoader();
}

function addFiles(files) {
    selectedFiles.push(...files);
    totalSize = selectedFiles.reduce((sum, file) => sum + file.size, 0);
    updateFileList();
    updateUploadOptions();
}

function removeFile(index) {
    totalSize -= selectedFiles[index].size;
    selectedFiles.splice(index, 1);
    updateFileList();
    updateUploadOptions();
}

function updateFileList() {
    const fileList = document.getElementById('file-list');
    
    if (selectedFiles.length === 0) {
        fileList.innerHTML = '';
        return;
    }

    fileList.innerHTML = selectedFiles.map((file, index) => `
        <div class="file-item">
            <div class="file-info">
                <div class="file-name">${file.name}</div>
                <div class="file-meta">${formatFileSize(file.size)}</div>
            </div>
            <button class="remove-file" onclick="removeFile(${index})">×</button>
        </div>
    `).join('');
}

function updateUploadOptions() {
    const uploadOptions = document.getElementById('upload-options');
    uploadOptions.classList.toggle('hidden', selectedFiles.length === 0);
}

function formatFileSize(bytes) {
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
}

// Upload functions
async function startTransfer() {
    if (selectedFiles.length === 0) {
        showToast('error', 'Please select files to upload');
        return;
    }

    showUploadProgress();
    setUploadStatus('Initializing transfer...');

    try {
        const expiry = document.getElementById('expiry').value;
        const message = document.getElementById('message').value;

        // Initialize transfer
        const initResponse = await fetch(ENDPOINTS.NEW_TRANSFER, {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({ expiry, message, size: totalSize })
        });

        const initData = await initResponse.json();
        if (!initResponse.ok) throw new Error(initData.error?.message);

        transferId = initData.transfer_id;
        maxChunkSize = initData.max_chunk_size;

        setUploadStatus('Preparing files...');
        showLoader();

        // Create zip file
        const zip = new JSZip();
        for (const file of selectedFiles) {
            zip.file(file.name, file);
        }

        const zippedBlob = await zip.generateAsync({ type: 'blob' });
        zippedFile = new File([zippedBlob], `${transferId}.zip`, { type: 'application/zip' });
        totalChunks = Math.ceil(zippedFile.size / maxChunkSize);
        currentChunk = 0;

        hideLoader();
        setUploadStatus('Uploading files...');
        
        await uploadFiles();

    } catch (error) {
        showToast('error', error.message);
        resetUpload();
    }
}

async function uploadFiles() {
    if (!zippedFile) return;

    uploadInProgress = true;

    for (let i = currentChunk; i < totalChunks; i++) {
        if (isPaused) {
            currentChunk = i;
            uploadInProgress = false;
            return;
        }

        const start = i * maxChunkSize;
        const end = Math.min(start + maxChunkSize, zippedFile.size);
        const chunk = zippedFile.slice(start, end);

        const formData = new FormData();
        formData.append('uploadId', transferId);
        formData.append('index', i.toString());
        formData.append('chunk', chunk);

        try {
            const response = await fetch(ENDPOINTS.UPLOAD_CHUNK, {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${authToken}` },
                body: formData
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error?.message || 'Upload failed');
            }

            const progress = Math.round(((i + 1) / totalChunks) * 100);
            setProgress(progress);

        } catch (error) {
            showToast('error', 'Upload failed: ' + error.message);
            uploadInProgress = false;
            return;
        }
    }

    uploadInProgress = false;
    await finalizeTransfer();
}

async function finalizeTransfer() {
    setUploadStatus('Finalizing transfer...');
    showLoader();

    try {
        const response = await fetch(ENDPOINTS.ASSEMBLE, {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authToken}`
            },
            body: JSON.stringify({ id: transferId })
        });

        const data = await response.json();
        if (!response.ok) throw new Error(data.error?.message);

        hideLoader();
        showUploadSuccess();

    } catch (error) {
        hideLoader();
        showToast('error', 'Failed to finalize transfer: ' + error.message);
    }
}

function togglePause() {
    isPaused = !isPaused;
    const btn = document.getElementById('pause-btn');
    
    if (isPaused) {
        btn.textContent = '▶ Resume';
    } else {
        btn.textContent = '⏸ Pause';
        if (!uploadInProgress) {
            uploadFiles();
        }
    }
}

async function cancelTransfer() {
    isPaused = true;
    uploadInProgress = false;

    if (transferId) {
        try {
            await fetch(ENDPOINTS.CANCEL, {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${authToken}`
                },
                body: JSON.stringify({ transfer_id: transferId })
            });
        } catch (error) {
            console.error('Cancel failed:', error);
        }
    }

    resetUpload();
}

function resetUpload() {
    transferId = null;
    zippedFile = null;
    totalChunks = 0;
    currentChunk = 0;
    isPaused = false;
    uploadInProgress = false;
    setProgress(0);
    showFileSelection();
}

function newTransfer() {
    selectedFiles = [];
    totalSize = 0;
    updateFileList();
    updateUploadOptions();
    resetUpload();
}

// UI helper functions
function showFileSelection() {
    document.getElementById('file-selection').classList.remove('hidden');
    document.getElementById('upload-progress').classList.add('hidden');
    document.getElementById('upload-success').classList.add('hidden');
}

function showUploadProgress() {
    document.getElementById('file-selection').classList.add('hidden');
    document.getElementById('upload-progress').classList.remove('hidden');
    document.getElementById('upload-success').classList.add('hidden');
}

function showUploadSuccess() {
    document.getElementById('file-selection').classList.add('hidden');
    document.getElementById('upload-progress').classList.add('hidden');
    document.getElementById('upload-success').classList.remove('hidden');
    const shareUrl = `${window.location.origin}/share/${transferId}`;
    document.getElementById('share-url').textContent = shareUrl;
}

function setProgress(percent) {
    const circle = document.getElementById('progress-circle');
    const text = document.getElementById('progress-text');
    const circumference = 2 * Math.PI * 65;
    const offset = circumference - (percent / 100) * circumference;
    
    circle.style.strokeDashoffset = offset;
    text.textContent = `${percent}%`;
}

function setUploadStatus(status) {
    document.getElementById('upload-status').textContent = status;
}

function showLoader() {
    document.getElementById('upload-loader').style.display = 'block';
}

function hideLoader() {
    document.getElementById('upload-loader').style.display = 'none';
}

function copyShareLink() {
    const shareUrl = document.getElementById('share-url').textContent;
    navigator.clipboard.writeText(shareUrl).then(() => {
        showToast('success', 'Link copied to clipboard!');
    });
}
