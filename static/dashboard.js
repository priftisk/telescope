const REFRESH_INTERVAL_MS = 5000;
let refreshTimer = null;

function CreateRoutesList(routesData) {
    const container = document.getElementById('routes-container');
    const activeCount = document.getElementById('active-routes');
    container.innerHTML = '';

    activeCount.textContent = routesData ? routesData.length : 0;

    if (!routesData || routesData.length === 0) {
        container.innerHTML = `
            <div class="empty">
                <p>NO ACTIVE ROUTES</p>
                <p style="font-size: 0.875rem; margin-top: 0.5rem;">Start a container with proxy labels to see it here</p>
            </div>
        `;
        return;
    }

    routesData.forEach(route => {
        const card = document.createElement('div');
        card.className = 'route-card';

        card.innerHTML = `
            <div class="route-main">
                <span class="route-host">${escapeHtml(route.hostname)}</span>
                ${route.URLPath ? `<span class="route-path">${escapeHtml(route.url_path)}</span>` : ''}
                <span class="route-arrow">→</span>
                <span class="route-target">${escapeHtml(route.address)}</span>
            </div>
            <div class="route-meta">
                <span class="route-container-name">${escapeHtml(route.container_name)}</span>
                <span class="route-container-id">${escapeHtml(route.container_id)}</span>
            </div>
        `;

        container.appendChild(card);
    });
}

function showLoading() {
    const container = document.getElementById('routes-container');
    container.innerHTML = `
        <div class="empty">
            <p>LOADING ROUTES...</p>
        </div>
    `;
}

function showError(message) {
    const container = document.getElementById('routes-container');
    container.innerHTML = `
        <div class="empty error">
            <p>CONNECTION LOST</p>
            <p style="font-size: 0.875rem; margin-top: 0.5rem;">${escapeHtml(message)}</p>
        </div>
    `;
}

function escapeHtml(str) {
    if (str === undefined || str === null) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

function setRefreshing(isRefreshing) {
    const btn = document.getElementById('refresh-btn');
    btn.disabled = isRefreshing;
    btn.classList.toggle('refreshing', isRefreshing);
}

async function GetRoutes({ showLoadingState = false } = {}) {
    const url = "http://localhost:8900/routes";

    if (showLoadingState) {
        showLoading();
    }
    setRefreshing(true);

    try {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`Response status: ${response.status}`);
        }
        const result = await response.json();
        CreateRoutesList(result);
    } catch (error) {
        console.error(error.message);
        showError(error.message);
    } finally {
        setRefreshing(false);
    }
}

function startAutoRefresh() {
    if (refreshTimer) clearInterval(refreshTimer);
    refreshTimer = setInterval(() => GetRoutes(), REFRESH_INTERVAL_MS);
}
document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('refresh-btn').addEventListener('click', () => {
        GetRoutes({ showLoadingState: true });
        startAutoRefresh();
    });

    GetRoutes({ showLoadingState: true });
    startAutoRefresh();
});