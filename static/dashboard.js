const REFRESH_INTERVAL_MS = 5000;
let refreshTimer = null;
let hostname = () => {
    return location.protocol + '//' + location.host
}
const BASE_URL = hostname() + "/dashboard";

// ---------- Formatting helpers ----------

function formatUptime(ms) {
    const totalSeconds = Math.floor(ms / 1000);

    const hours = String(Math.floor(totalSeconds / 3600)).padStart(2, "0");
    const minutes = String(Math.floor((totalSeconds % 3600) / 60)).padStart(2, "0");
    const seconds = String(totalSeconds % 60).padStart(2, "0");

    return `${hours}:${minutes}:${seconds}`;
}

function escapeHtml(str) {
    if (str === undefined || str === null) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

function getStatusClass(code) {
    if (code >= 500) return 'server-error';
    if (code >= 400) return 'client-error';
    if (code >= 300) return 'redirect';
    return 'success';
}

function formatDuration(ns) {
    const ms = ns / 1e6;
    if (ms < 1) return `${(ns / 1e3).toFixed(0)}µs`;
    if (ms < 1000) return `${ms.toFixed(1)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
}

function formatBytes(bytes) {
    if (bytes < 1024) return `${bytes}B`;
    if (bytes < 1024 ** 2) return `${(bytes / 1024).toFixed(1)}KB`;
    return `${(bytes / 1024 ** 2).toFixed(1)}MB`;
}

function formatTime(iso) {
    return new Date(iso).toLocaleTimeString([], { hour12: false });
}

// ---------- Rendering ----------

function renderUptime(uptime) {
    const uptimeContainer = document.getElementById('uptime-container');
    const formattedTime = formatUptime(uptime);
    uptimeContainer.innerHTML = `
        <div class="plate-label">Uptime</div>
        <div id="uptime" class="plate-value">${escapeHtml(formattedTime)}</div>
    `;
}

function renderActiveCount(newCount) {
    const activeCount = document.getElementById('active-routes');
    activeCount.textContent = newCount;
}

function renderRoutesList(container, routes) {
    routes.forEach(route => {
        const card = document.createElement('div');
        card.className = 'route-card';

        card.innerHTML = `
                <div class="route-status" aria-hidden="true"></div>
                <div class="route-main">
                    <span class="route-host">${escapeHtml(route.hostname)}</span>
                    ${route.url_path ? `<span class="route-path">${escapeHtml(route.url_path)}</span>` : ''}
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

function renderRoutes(data) {
    const container = document.getElementById('routes-container');
    container.innerHTML = '';

    const totalRoutes = data?.total_routes ?? 0;
    const routes = data?.routes ?? [];
    const uptime = data?.uptime ?? "";

    renderUptime(uptime);
    renderActiveCount(totalRoutes);

    if (!routes.length || totalRoutes === 0) {
        container.innerHTML = `
            <div class="empty">
                <p>NO ACTIVE ROUTES</p>
                <p style="font-size: 0.875rem; margin-top: 0.5rem;">Start a container with proxy labels to see it here</p>
            </div>
        `;
        return;
    }

    renderRoutesList(container, routes);
}

function renderTripsList(trips, container) {
    trips.forEach(trip => {
        const card = document.createElement('div');
        card.className = 'trip-card';
        card.innerHTML = `
        <div class="trip-primary">
            <span class="trip-method" data-method="${trip.method}">${escapeHtml(trip.method)}</span>
            <span class="trip-path">
            ${escapeHtml(trip.path)}${trip.query ? `<span class="trip-query">?${escapeHtml(trip.query)}</span>` : ''}
            </span>
            <span class="trip-status" data-status-class="${getStatusClass(trip.status_code)}">${trip.status_code}</span>
        </div>

        <div class="trip-secondary">
            <span class="trip-host">${escapeHtml(trip.host)}</span>
            <span class="trip-duration">${formatDuration(trip.duration)}</span>
            <span class="trip-size">${formatBytes(trip.response_size_bytes)}</span>
            <span class="trip-time">${formatTime(trip.timestamp)}</span>
        </div>

        ${(trip.remote_addr || trip.user_agent) ? `
        <div class="trip-tertiary">
            ${trip.remote_addr ? `<span class="trip-remote">${escapeHtml(trip.remote_addr)}</span>` : ''}
            ${trip.user_agent ? `<span class="trip-ua">${escapeHtml(trip.user_agent)}</span>` : ''}
        </div>` : ''}

        ${trip.error ? `<div class="trip-error">${escapeHtml(trip.error)}</div>` : ''}
                `;
        container.appendChild(card)
    })
}


function renderTrips(trips) {
    const container = document.getElementById('trips-container')
    container.innerHTML = ''
    if (!trips || trips?.length == 0) {
        container.innerHTML = `
         <div class="empty">
            <p>NO TRIPS FOUND</p>
            <p style="font-size: 0.875rem; margin-top: 0.5rem;">Make a request to any of the routes to see it here.</p>
        </div>
        `;
        return;
    }
    renderTripsList(trips, container)
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

function setRefreshing(isRefreshing) {
    const btn = document.getElementById('refresh-btn');
    btn.disabled = isRefreshing;
    btn.classList.toggle('refreshing', isRefreshing);
}

// ---------- Data fetching ----------

async function fetchJson(path) {
    const response = await fetch(`${BASE_URL}/${path}`);
    if (!response.ok) {
        throw new Error(`Response status: ${response.status}`);
    }
    return response.json();
}

async function getDashboardData({ showLoadingState = false } = {}) {
    if (showLoadingState) showLoading();
    setRefreshing(true);

    try {
        const result = await fetchJson('data');
        renderRoutes(result);
    } catch (error) {
        console.error(error.message);
        showError(error.message);
    } finally {
        setRefreshing(false);
    }
}

async function getDashboardTrips() {
    try {
        const result = await fetchJson('trips');
        renderTrips(result)
    } catch (error) {
        console.error(error.message);
    }
}

// ---------- Auto-refresh ----------

function startAutoRefresh() {
    if (refreshTimer) clearInterval(refreshTimer);
    refreshTimer = setInterval(() => {
        getDashboardData();
        getDashboardTrips();
    }, REFRESH_INTERVAL_MS);
}

document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('refresh-btn').addEventListener('click', () => {
        getDashboardData({ showLoadingState: true });
        startAutoRefresh();
    });

    getDashboardData({ showLoadingState: true });
    startAutoRefresh();
});