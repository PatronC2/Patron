import { afterEach, expect, it, vi } from 'vitest';

const { render, createRoot } = vi.hoisted(() => {
  const render = vi.fn();
  return { render, createRoot: vi.fn(() => ({ render })) };
});
vi.mock('react-dom/client', () => ({ default: { createRoot } }));
vi.mock('./components/App', () => ({ default: () => null }));
vi.mock('./context/AuthProvider', () => ({ AuthProvider: ({ children }) => children }));
vi.mock('./context/Themes', () => ({ ThemeProvider: ({ children }) => children }));
vi.mock('./context/AxiosProvider', () => ({ AxiosProvider: ({ children }) => children }));

afterEach(() => {
  vi.unstubAllGlobals();
  delete window.runtimeConfig;
});

it('waits for runtime configuration before creating and rendering the React root', async () => {
  let resolveResponse;
  let resolveConfig;
  const response = new Promise(resolve => { resolveResponse = resolve; });
  const config = new Promise(resolve => { resolveConfig = resolve; });
  vi.stubGlobal('fetch', vi.fn(() => response));
  document.body.innerHTML = '<div id="root"></div>';
  await import('./index.jsx');
  expect(fetch).toHaveBeenCalledWith('/config.json');
  expect(createRoot).not.toHaveBeenCalled();
  resolveResponse({ json: () => config });
  await Promise.resolve();
  expect(render).not.toHaveBeenCalled();
  const runtime = { REACT_APP_NGINX_IP: 'example.test', REACT_APP_NGINX_PORT: '8443' };
  resolveConfig(runtime);
  await vi.waitFor(() => expect(render).toHaveBeenCalledOnce());
  expect(window.runtimeConfig).toEqual(runtime);
  expect(createRoot).toHaveBeenCalledWith(document.getElementById('root'));
});
