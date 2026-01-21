function toggleIcon(el) {
    const folderIcon = el.parentElement.querySelector('.folder-icon');

    if (el.classList.contains('pi-chevron-right')) {
        el.classList.remove('pi-chevron-right');
        el.classList.add('pi-chevron-down');

        if (folderIcon) {
            folderIcon.classList.remove('pi-folder');
            folderIcon.classList.add('pi-folder-open');
        }
    } else {
        el.classList.remove('pi-chevron-down');
        el.classList.add('pi-chevron-right');

        if (folderIcon) {
            folderIcon.classList.remove('pi-folder-open');
            folderIcon.classList.add('pi-folder');
        }
        
        let currentRow = el.closest('tr');
        let nextRow = currentRow.nextElementSibling;
        
        const currentPadding = el.closest('td').style.paddingLeft;
        const currentLevelValue = parseFloat(currentPadding) || 0;
        
        while (nextRow) {
            let nextTd = nextRow.querySelector('td');
            if (!nextTd) break;
            
            let nextPadding = nextTd.style.paddingLeft;
            let nextLevelValue = parseFloat(nextPadding) || 0;
            
            if (nextLevelValue > currentLevelValue) {
                let toRemove = nextRow;
                nextRow = nextRow.nextElementSibling;
                toRemove.remove();
            } else {
                break;
            }
        }
    }
}

document.addEventListener('htmx:responseError', function(evt) {
    showToast(evt.detail.xhr.responseText || "An unexpected error has occurred", "error");
});

function showToast(message, type = "error") {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `toast toast-${type} shadow-8`;
    
    const icon = type === "error" ? "pi-exclamation-circle" : "pi-check-circle";
    
    toast.innerHTML = `
        <i class="pi ${icon} mr-2"></i>
        <span>${message}</span>
    `;
    
    container.appendChild(toast);

    setTimeout(() => {
        toast.style.opacity = '0';
        setTimeout(() => toast.remove(), 500);
    }, 3000);
}
