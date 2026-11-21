// @vitest-environment node
import { mkdtemp, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterEach, expect, it, vi } from 'vitest';
import { runtimeConfigPlugin } from './vite.config';

let directory;
afterEach(async () => {
  if (directory) await rm(directory, { recursive: true, force: true });
});

it('serves current runtime JSON without caching, and reports missing or invalid files', async () => {
  directory = await mkdtemp(join(tmpdir(), 'patron-config-'));
  const filename = join(directory, 'config.json');
  let middleware;
  runtimeConfigPlugin(filename).configureServer({ middlewares: { use: handler => { middleware = handler; } } });
  const request = async url => {
    const res = { setHeader: vi.fn(), end: vi.fn() };
    const next = vi.fn();
    await middleware({ url }, res, next);
    return { res, next };
  };
  const missing = await request('/config.json');
  expect(missing.res.statusCode).toBe(500);
  expect(JSON.parse(missing.res.end.mock.calls[0][0]).error).toMatch(/Missing/);
  for (const host of ['first.test', 'second.test']) {
    const content = JSON.stringify({ REACT_APP_NGINX_IP: host });
    await writeFile(filename, content);
    const { res, next } = await request('/config.json?reload=1');
    expect(res.end).toHaveBeenCalledWith(content);
    expect(res.setHeader).toHaveBeenCalledWith('Content-Type', 'application/json');
    expect(res.setHeader).toHaveBeenCalledWith('Cache-Control', 'no-store');
    expect(next).not.toHaveBeenCalled();
  }
  await writeFile(filename, 'invalid');
  const invalid = await request('/config.json');
  expect(invalid.res.statusCode).toBe(500);
  expect(JSON.parse(invalid.res.end.mock.calls[0][0]).error).toMatch(/valid JSON/);
  expect((await request('/home')).next).toHaveBeenCalledOnce();
});
