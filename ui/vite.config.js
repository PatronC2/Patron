import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export function runtimeConfigPlugin(configFile = fileURLToPath(new URL('../frontend-config.json', import.meta.url))) {
  return {
    name: 'patron-runtime-config',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use(async (req, res, next) => {
        if (req.url?.split('?')[0] !== '/config.json') return next();
        res.setHeader('Content-Type', 'application/json');
        res.setHeader('Cache-Control', 'no-store');
        try {
          const config = await readFile(configFile, 'utf8');
          JSON.parse(config);
          res.end(config);
        } catch (error) {
          res.statusCode = 500;
          res.end(JSON.stringify({ error: error.code === 'ENOENT'
            ? 'Missing frontend-config.json in the repository root'
            : 'Unable to read valid JSON from frontend-config.json' }));
        }
      });
    },
  };
}

export default defineConfig({
  plugins: [react(), runtimeConfigPlugin()],
  server: { port: 3000, strictPort: true },
  build: { outDir: 'build', assetsDir: 'static' },
  test: { environment: 'jsdom', setupFiles: ['./src/test/setup.js'], clearMocks: true },
});
