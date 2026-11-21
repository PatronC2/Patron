import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createAxios } from './axios';

const browserWindow = window;
let location;
beforeEach(() => {
  localStorage.clear();
  let href = '/home';
  location = { pathname: '/home', get href() { return href; }, set href(value) { href = value; } };
  vi.stubGlobal('window', {
    runtimeConfig: { REACT_APP_NGINX_IP: 'api.example.test', REACT_APP_NGINX_PORT: '8443' },
    location,
  });
});
afterEach(() => {
  vi.unstubAllGlobals();
  delete browserWindow.runtimeConfig;
});

describe('runtime API configuration', () => {
  it('requires configuration before creating the client', () => {
    delete window.runtimeConfig;
    expect(createAxios).toThrow('Runtime config not loaded');
  });

  it('uses the configured endpoint and current token for requests', async () => {
    const client = createAxios();
    const adapter = vi.fn(async config => ({ data: {}, status: 200, statusText: 'OK', headers: {}, config }));
    localStorage.setItem('auth', JSON.stringify({ token: 'first-token' }));
    await client.get('/example', { adapter });
    expect(adapter.mock.calls[0][0].baseURL).toBe('https://api.example.test:8443');
    expect(adapter.mock.calls[0][0].headers.Authorization).toBe('first-token');
    localStorage.setItem('auth', JSON.stringify({ token: 'new-token' }));
    await client.get('/example', { adapter });
    expect(adapter.mock.calls[1][0].headers.Authorization).toBe('new-token');
  });

  it.each(['/home', '/login'])('clears expired auth on %s without a login redirect loop', async pathname => {
    location.pathname = pathname;
    location.href = pathname;
    const redirect = vi.spyOn(location, 'href', 'set');
    localStorage.setItem('auth', JSON.stringify({ token: 'expired' }));
    const error = { response: { status: 401 } };
    await expect(createAxios().get('/example', { adapter: () => Promise.reject(error) })).rejects.toBe(error);
    expect(localStorage.getItem('auth')).toBeNull();
    expect(location.href).toBe('/login');
    expect(redirect).toHaveBeenCalledTimes(pathname === '/login' ? 0 : 1);
  });
});
