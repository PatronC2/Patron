import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import SideMenu from './Menu';
import Login from './Login';
import Home from './Home';
import Payloads from './Payloads';
import { AuthProvider } from '../context/AuthProvider';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  const handleSuccessfulLogin = () => {
    setIsLoggedIn(true);
  };

  return (
    <AuthProvider>
      <Router>
        <div className="App">
          {isLoggedIn && <SideMenu />}
          <main className="content">
            <Routes>
              <Route path="/" element={isLoggedIn ? <Navigate to="/home" /> : <Login onSuccessfulLogin={handleSuccessfulLogin} />} />
              <Route path="/home" element={isLoggedIn ? <Home /> : <Navigate to="/" />} />
              <Route path="/payloads" element={<Payloads />} />
            </Routes>
          </main>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
