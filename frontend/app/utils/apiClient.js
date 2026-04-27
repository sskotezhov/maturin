const API_BASE_URL = 'https://матурин15.рф/api/v1';

let isRefreshing = false;
let refreshQueue = [];

function drainQueue(error, token) {
  refreshQueue.forEach(({ resolve, reject }) =>
    error ? reject(error) : resolve(token)
  );
  refreshQueue = [];
}

async function refreshAccessToken() {
  const refreshToken = localStorage.getItem('refresh_token');
  if (!refreshToken) throw new Error('No refresh token');

  const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: {
      accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!response.ok) throw new Error('Refresh failed');

  const data = await response.json();
  localStorage.setItem('access_token', data.access_token);
  if (data.refresh_token) {
    localStorage.setItem('refresh_token', data.refresh_token);
  }
  return data.access_token;
}

function clearAuth() {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('user');
  window.dispatchEvent(new Event('auth:logout'));
}

export async function apiFetch(path, options = {}) {
  const accessToken = localStorage.getItem('access_token');

  const makeRequest = (token) =>
    fetch(`${API_BASE_URL}${path}`, {
      ...options,
      headers: {
        accept: 'application/json',
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    });

  let response = await makeRequest(accessToken);

  if (response.status !== 401) return response;

  if (isRefreshing) {
    const newToken = await new Promise((resolve, reject) =>
      refreshQueue.push({ resolve, reject })
    );
    return makeRequest(newToken);
  }

  isRefreshing = true;
  try {
    const newToken = await refreshAccessToken();
    drainQueue(null, newToken);
    return makeRequest(newToken);
  } catch (err) {
    drainQueue(err, null);
    clearAuth();
    throw err;
  } finally {
    isRefreshing = false;
  }
}
