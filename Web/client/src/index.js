import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import './index.css'
import './fonts/Lucida_Console_Regular.ttf'; 
import App from './App';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  // <React.StrictMode>
  <BrowserRouter>
  <App/>
  </BrowserRouter>
  // </React.StrictMode>
);
