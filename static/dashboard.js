const REFRESH_INTERVAL_MS = 5000;
let refreshTimer = null;


function formatUptime(ms) {
  const totalSeconds = Math.floor(ms / 1000);

  const hours = String(Math.floor(totalSeconds / 3600)).padStart(2, "0");
  const minutes = String(Math.floor((totalSeconds % 3600) / 60)).padStart(2, "0");
  const seconds = String(totalSeconds % 60).padStart(2, "0");

  return `${hours}:${minutes}:${seconds}`;
}

function RenderData(data) {
    const uptimeContainer = document.getElementById('uptime-container');
    const container = document.getElementById('routes-container');
    const activeCount = document.getElementById('active-routes');
    container.innerHTML = '';

    let total_routes = data?.total_routes ?? 0
    let routes = data?.routes ?? []
    let uptime = data?.uptime ?? ""
    
    activeCount.textContent = total_routes
    formattedTime = formatUptime(uptime)
    uptimeContainer.innerHTML = `
         <div class="plate-label">Uptime</div>
          <div id="uptime" class="plate-value">${escapeHtml(formattedTime)}</div>
    `
    if (!routes || total_routes === 0) {
        container.innerHTML = `
            <div class="empty">
                <p>NO ACTIVE ROUTES</p>
                <p style="font-size: 0.875rem; margin-top: 0.5rem;">Start a container with proxy labels to see it here</p>
            </div>
        `;
        return;
    }

    routes.forEach(route => {
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

async function GetDashboardData({ showLoadingState = false } = {}) {
    const url = "http://localhost:8900/dashboard/data";

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
        console.log(result)
        RenderData(result);
    } catch (error) {
        console.error(error.message);
        showError(error.message);
    } finally {
        setRefreshing(false);
    }
}

function startAutoRefresh() {
    if (refreshTimer) clearInterval(refreshTimer);
    refreshTimer = setInterval(() => GetDashboardData(), REFRESH_INTERVAL_MS);
}
document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('refresh-btn').addEventListener('click', () => {
        GetDashboardData({ showLoadingState: true });
        startAutoRefresh();
    });

    GetDashboardData({ showLoadingState: true });
    startAutoRefresh();
});