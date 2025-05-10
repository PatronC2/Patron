import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './components/App';
import { AuthProvider } from './context/AuthProvider';
import { ThemeProvider } from './context/Themes';
import { AxiosProvider } from './context/AxiosProvider';

const loadConfig = async () => {
  const res = await fetch('/config.json');
  window.runtimeConfig = await res.json();
};

loadConfig().then(() => {
  ReactDOM.createRoot(document.getElementById('root')).render(
    <React.StrictMode>
      <AxiosProvider>
        <ThemeProvider>
          <AuthProvider>
            <App />
          </AuthProvider>
        </ThemeProvider>
      </AxiosProvider>
    </React.StrictMode>
  );
});