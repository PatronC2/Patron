import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './components/App'; // Import the actual App component
import { AuthProvider } from './context/AuthProvider';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <AuthProvider>
      <App />
    </AuthProvider>
  </React.StrictMode>
);
