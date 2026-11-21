import js from '@eslint/js';
import globals from 'globals';
import react from 'eslint-plugin-react';
import hooks from 'eslint-plugin-react-hooks';
import refresh from 'eslint-plugin-react-refresh';

export default [
  { ignores: ['build/**', 'coverage/**', 'node_modules/**'] },
  {
    files: ['**/*.{js,jsx}'],
    ...js.configs.recommended,
    languageOptions: { ecmaVersion: 'latest', sourceType: 'module', parserOptions: { ecmaFeatures: { jsx: true } } },
    plugins: { react, 'react-hooks': hooks, 'react-refresh': refresh },
    rules: {
      ...js.configs.recommended.rules,
      'no-unused-vars': 'warn',
      'react/jsx-uses-react': 'error',
      'react/jsx-uses-vars': 'error',
      'react-hooks/rules-of-hooks': 'warn',
      'react-hooks/exhaustive-deps': 'warn',
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    },
  },
  { files: ['src/**/*.{js,jsx}'], languageOptions: { globals: globals.browser } },
  { files: ['*.config.js', '*.config.test.js'], languageOptions: { globals: globals.node } },
];
