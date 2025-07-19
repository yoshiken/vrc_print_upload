import './style.css';

// Import Wails runtime and Go functions
import {
    IsAuthenticated,
    Login,
    VerifyTwoFactor,
    Logout,
    GetCurrentUser,
    UploadImage,
    ValidateImageFile,
    OpenFileDialog
} from '../wailsjs/go/main/App';

// Application state
let currentUser = null;
let selectedFile = null;
let selectedFilePath = null;

// Initialize the application
document.addEventListener('DOMContentLoaded', async () => {
    await initializeApp();
});

async function initializeApp() {
    try {
        // Check if user is already authenticated
        const isAuth = await IsAuthenticated();
        
        if (isAuth) {
            // Get current user info and show main screen
            await loadUserInfo();
            showMainScreen();
        } else {
            // Show login screen
            showLoginScreen();
        }
    } catch (error) {
        console.error('Failed to initialize app:', error);
        showLoginScreen();
    }
    
    // Setup event listeners
    setupEventListeners();
}

function setupEventListeners() {
    // Login form
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }
    
    // 2FA verification
    const verify2FABtn = document.getElementById('verify-2fa-btn');
    if (verify2FABtn) {
        verify2FABtn.addEventListener('click', handleTwoFactorVerification);
    }
    
    // Logout button
    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', handleLogout);
    }
    
    // File selection
    const fileSelectBtn = document.getElementById('file-select-btn');
    const fileInput = document.getElementById('file-input');
    const dropZone = document.getElementById('drop-zone');
    const clearFileBtn = document.getElementById('clear-file-btn');
    
    if (fileSelectBtn) {
        fileSelectBtn.addEventListener('click', handleFileDialogOpen);
    }
    
    if (fileInput) {
        fileInput.addEventListener('change', handleFileSelect);
    }
    
    if (dropZone) {
        setupDragAndDrop(dropZone);
    }
    
    if (clearFileBtn) {
        clearFileBtn.addEventListener('click', clearSelectedFile);
    }
    
    // Upload button
    const uploadBtn = document.getElementById('upload-btn');
    if (uploadBtn) {
        uploadBtn.addEventListener('click', handleUpload);
    }
    
    // Recovery code checkbox
    const recoveryCheckbox = document.getElementById('recovery-code');
    const twoFactorInput = document.getElementById('two-factor-code');
    if (recoveryCheckbox && twoFactorInput) {
        recoveryCheckbox.addEventListener('change', (e) => {
            if (e.target.checked) {
                twoFactorInput.placeholder = 'リカバリーコード (XXXX-XXXX-XXXX)';
                twoFactorInput.maxLength = 20;
            } else {
                twoFactorInput.placeholder = '123456';
                twoFactorInput.maxLength = 6;
            }
        });
    }
}

function setupDragAndDrop(dropZone) {
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, preventDefaults, false);
    });
    
    ['dragenter', 'dragover'].forEach(eventName => {
        dropZone.addEventListener(eventName, () => dropZone.classList.add('drag-over'), false);
    });
    
    ['dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, () => dropZone.classList.remove('drag-over'), false);
    });
    
    dropZone.addEventListener('drop', handleDrop, false);
    dropZone.addEventListener('click', handleFileDialogOpen);
}

function preventDefaults(e) {
    e.preventDefault();
    e.stopPropagation();
}

async function handleDrop(e) {
    const dt = e.dataTransfer;
    const files = dt.files;
    
    if (files.length > 0) {
        const file = files[0];
        await processSelectedFile(file);
    }
}

async function handleFileDialogOpen() {
    try {
        const filePath = await OpenFileDialog();
        if (filePath) {
            await processSelectedFilePath(filePath);
        }
    } catch (error) {
        console.error('Error opening file dialog:', error);
        showStatusMessage('error', 'ファイルダイアログを開けませんでした');
    }
}

async function handleFileSelect(e) {
    const files = e.target.files;
    if (files.length > 0) {
        const file = files[0];
        await processSelectedFile(file);
    }
}

async function processSelectedFilePath(filePath) {
    try {
        // Validate file
        const validation = await ValidateImageFile(filePath);
        
        if (!validation.valid) {
            showStatusMessage('error', validation.error || 'Invalid file selected');
            return;
        }
        
        selectedFile = null; // No File object when using dialog
        selectedFilePath = filePath;
        
        // Show file info
        displaySelectedFilePath(filePath);
        
        // Enable upload button
        const uploadBtn = document.getElementById('upload-btn');
        if (uploadBtn) {
            uploadBtn.disabled = false;
        }
        
    } catch (error) {
        console.error('Error processing file:', error);
        showStatusMessage('error', 'Failed to process selected file');
    }
}

async function processSelectedFile(file) {
    try {
        // Validate file
        const validation = await ValidateImageFile(file.name);
        
        if (!validation.valid) {
            showStatusMessage('error', validation.error || 'Invalid file selected');
            return;
        }
        
        selectedFile = file;
        selectedFilePath = file.name; // For display only - won't work for upload
        
        // Show file info
        displaySelectedFile(file);
        
        // Show warning that drag & drop won't work for upload
        showStatusMessage('warning', 'ドラッグ&ドロップが検出されました。実際のアップロードには「画像ファイルを選択」ボタンをご利用ください。');
        
        // Enable upload button but it will show error
        const uploadBtn = document.getElementById('upload-btn');
        if (uploadBtn) {
            uploadBtn.disabled = false;
        }
        
    } catch (error) {
        console.error('Error processing file:', error);
        showStatusMessage('error', 'Failed to process selected file');
    }
}

function displaySelectedFilePath(filePath) {
    const fileInfo = document.getElementById('file-info');
    const fileName = document.getElementById('file-name');
    const imagePreview = document.getElementById('image-preview');
    const dropZone = document.getElementById('drop-zone');
    
    if (fileInfo && fileName) {
        fileName.textContent = filePath.split(/[\\/]/).pop(); // Get filename from path
        
        // Hide preview for now since we can't access file contents directly
        if (imagePreview) {
            imagePreview.style.display = 'none';
        }
        
        fileInfo.classList.remove('hidden');
        dropZone.style.display = 'none';
    }
}

function displaySelectedFile(file) {
    const fileInfo = document.getElementById('file-info');
    const fileName = document.getElementById('file-name');
    const imagePreview = document.getElementById('image-preview');
    const dropZone = document.getElementById('drop-zone');
    
    if (fileInfo && fileName && imagePreview) {
        fileName.textContent = file.name;
        
        // Create preview URL
        if (file.type && file.type.startsWith('image/')) {
            const reader = new FileReader();
            reader.onload = (e) => {
                imagePreview.src = e.target.result;
                imagePreview.style.display = 'block';
            };
            reader.readAsDataURL(file);
        }
        
        fileInfo.classList.remove('hidden');
        dropZone.style.display = 'none';
    }
}

function clearSelectedFile() {
    selectedFile = null;
    selectedFilePath = null;
    
    const fileInfo = document.getElementById('file-info');
    const dropZone = document.getElementById('drop-zone');
    const uploadBtn = document.getElementById('upload-btn');
    const fileInput = document.getElementById('file-input');
    
    if (fileInfo) fileInfo.classList.add('hidden');
    if (dropZone) dropZone.style.display = 'block';
    if (uploadBtn) uploadBtn.disabled = true;
    if (fileInput) fileInput.value = '';
    
    clearStatusMessage();
}

async function handleLogin(e) {
    e.preventDefault();
    
    const username = document.getElementById('username').value.trim();
    const password = document.getElementById('password').value;
    const loginBtn = document.getElementById('login-btn');
    
    if (!username || !password) {
        showStatusMessage('error', 'ユーザー名とパスワードを入力してください', 'login-status');
        return;
    }
    
    // Show loading state
    setButtonLoading(loginBtn, true);
    clearStatusMessage('login-status');
    
    try {
        const response = await Login({ username, password });
        
        if (response.success) {
            currentUser = { displayName: response.userDisplayName };
            showMainScreen();
            showStatusMessage('success', 'ログインに成功しました！');
        } else if (response.requiresTwoFactor) {
            show2FASection();
            showStatusMessage('info', response.message, 'login-status');
            // Focus on 2FA input field
            setTimeout(() => {
                const twoFactorInput = document.getElementById('two-factor-code');
                if (twoFactorInput) {
                    twoFactorInput.focus();
                }
            }, 100);
        } else {
            showStatusMessage('error', response.message, 'login-status');
        }
    } catch (error) {
        console.error('Login error:', error);
        showStatusMessage('error', 'ログインに失敗しました。再度お試しください。', 'login-status');
    } finally {
        setButtonLoading(loginBtn, false);
    }
}

async function handleTwoFactorVerification() {
    const code = document.getElementById('two-factor-code').value.trim();
    const isRecoveryCode = document.getElementById('recovery-code').checked;
    const verifyBtn = document.getElementById('verify-2fa-btn');
    
    if (!code) {
        showStatusMessage('error', '2FAコードを入力してください', 'login-status');
        return;
    }
    
    setButtonLoading(verifyBtn, true);
    
    try {
        const response = await VerifyTwoFactor({ code, isRecoveryCode });
        
        if (response.success) {
            currentUser = { displayName: response.userDisplayName };
            showMainScreen();
            showStatusMessage('success', 'ログインに成功しました！');
        } else {
            showStatusMessage('error', response.message, 'login-status');
        }
    } catch (error) {
        console.error('2FA verification error:', error);
        showStatusMessage('error', '2FA認証に失敗しました。再度お試しください。', 'login-status');
    } finally {
        setButtonLoading(verifyBtn, false);
    }
}

async function handleLogout() {
    try {
        const response = await Logout();
        
        if (response.success) {
            currentUser = null;
            selectedFile = null;
            selectedFilePath = null;
            showLoginScreen();
            showStatusMessage('success', 'ログアウトしました', 'login-status');
        } else {
            showStatusMessage('error', response.message);
        }
    } catch (error) {
        console.error('Logout error:', error);
        showStatusMessage('error', 'ログアウトに失敗しました');
    }
}

async function handleUpload() {
    if (!selectedFilePath) {
        showStatusMessage('error', '画像ファイルを選択してください');
        return;
    }
    
    const uploadBtn = document.getElementById('upload-btn');
    const progressContainer = document.getElementById('upload-progress');
    const progressFill = document.getElementById('progress-fill');
    const progressText = document.getElementById('progress-text');
    
    // Get form data
    const resizeOption = document.querySelector('input[name="resize"]:checked').value;
    const note = document.getElementById('note').value.trim();
    const worldId = document.getElementById('world-id').value.trim();
    const worldName = document.getElementById('world-name').value.trim();
    
    const uploadRequest = {
        imagePath: selectedFilePath,
        note: note,
        worldId: worldId,
        worldName: worldName,
        noResize: resizeOption === 'keep'
    };
    
    // Show progress
    setButtonLoading(uploadBtn, true);
    if (progressContainer) {
        progressContainer.classList.remove('hidden');
        progressFill.style.width = '0%';
        progressText.textContent = 'アップロード準備中...';
    }
    
    try {
        // Simulate progress (since we don't have real-time progress from Go)
        const progressInterval = setInterval(() => {
            const currentWidth = parseFloat(progressFill.style.width) || 0;
            if (currentWidth < 90) {
                progressFill.style.width = (currentWidth + 10) + '%';
            }
        }, 200);
        
        const response = await UploadImage(uploadRequest);
        
        clearInterval(progressInterval);
        
        if (response.success) {
            progressFill.style.width = '100%';
            progressText.textContent = 'アップロード完了！';
            
            showStatusMessage('success', `アップロードに成功しました！ファイルID: ${response.fileId}`);
            
            // Clear form after successful upload
            setTimeout(() => {
                clearSelectedFile();
                clearForm();
                if (progressContainer) progressContainer.classList.add('hidden');
            }, 2000);
        } else {
            showStatusMessage('error', response.error || 'アップロードに失敗しました');
            if (progressContainer) progressContainer.classList.add('hidden');
        }
    } catch (error) {
        console.error('Upload error:', error);
        showStatusMessage('error', 'アップロードに失敗しました。再度お試しください。');
        if (progressContainer) progressContainer.classList.add('hidden');
    } finally {
        setButtonLoading(uploadBtn, false);
    }
}

function clearForm() {
    document.getElementById('note').value = '';
    document.getElementById('world-id').value = '';
    document.getElementById('world-name').value = '';
    document.querySelector('input[name="resize"][value="resize"]').checked = true;
}

async function loadUserInfo() {
    try {
        const response = await GetCurrentUser();
        if (response.success) {
            currentUser = { displayName: response.userDisplayName };
            updateUserDisplay();
        }
    } catch (error) {
        console.error('Failed to load user info:', error);
    }
}

function updateUserDisplay() {
    const userInfo = document.getElementById('user-info');
    if (userInfo && currentUser) {
        userInfo.textContent = `ログイン中: ${currentUser.displayName}`;
    }
}

function showLoginScreen() {
    const loginScreen = document.getElementById('login-screen');
    const mainScreen = document.getElementById('main-screen');
    
    if (loginScreen) loginScreen.classList.remove('hidden');
    if (mainScreen) mainScreen.classList.add('hidden');
    
    // Reset login form
    const loginForm = document.getElementById('login-form');
    if (loginForm) loginForm.reset();
    
    hide2FASection();
    clearStatusMessage('login-status');
}

function showMainScreen() {
    const loginScreen = document.getElementById('login-screen');
    const mainScreen = document.getElementById('main-screen');
    
    if (loginScreen) loginScreen.classList.add('hidden');
    if (mainScreen) mainScreen.classList.remove('hidden');
    
    // Reset scroll position to top
    window.scrollTo(0, 0);
    document.body.scrollTop = 0;
    document.documentElement.scrollTop = 0;
    
    updateUserDisplay();
    clearStatusMessage();
}

function show2FASection() {
    const twoFactorSection = document.getElementById('two-factor-section');
    if (twoFactorSection) {
        twoFactorSection.classList.remove('hidden');
    }
}

function hide2FASection() {
    const twoFactorSection = document.getElementById('two-factor-section');
    if (twoFactorSection) {
        twoFactorSection.classList.add('hidden');
    }
}

function setButtonLoading(button, loading) {
    if (!button) return;
    
    const btnText = button.querySelector('.btn-text');
    const btnLoading = button.querySelector('.btn-loading');
    
    if (loading) {
        button.disabled = true;
        if (btnText) btnText.classList.add('hidden');
        if (btnLoading) btnLoading.classList.remove('hidden');
    } else {
        button.disabled = false;
        if (btnText) btnText.classList.remove('hidden');
        if (btnLoading) btnLoading.classList.add('hidden');
    }
}

function showStatusMessage(type, message, elementId = 'main-status') {
    let statusElement = document.getElementById(elementId);
    
    if (!statusElement) {
        // Try to find the status element in the current screen
        statusElement = document.querySelector('.screen:not(.hidden) .status-message');
    }
    
    if (statusElement) {
        statusElement.textContent = message;
        statusElement.className = `status-message ${type}`;
        statusElement.style.display = 'block';
        
        // Auto-hide success messages after 5 seconds
        if (type === 'success') {
            setTimeout(() => {
                clearStatusMessage(elementId);
            }, 5000);
        }
    }
}

function clearStatusMessage(elementId = 'main-status') {
    const statusElement = document.getElementById(elementId);
    if (statusElement) {
        statusElement.textContent = '';
        statusElement.className = 'status-message';
        statusElement.style.display = 'none';
    }
}
