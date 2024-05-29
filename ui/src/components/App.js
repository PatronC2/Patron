import React, { useState, useContext, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import SideMenu from './Menu';
import Login from './Login';
import Home from './Home';
import Payloads from './Payloads';
import { AuthProvider } from '../context/AuthProvider';
import AuthContext from '../context/AuthProvider';

function App() {
  const { auth } = useContext(AuthContext);
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    setIsLoggedIn(!!auth.token);
  }, [auth]);

  const handleSuccessfulLogin = () => {
    setIsLoggedIn(true);
  };

  return (
    <AuthProvider>
      <Router>
        <div className="App">
          {isLoggedIn && <SideMenu setIsLoggedIn={setIsLoggedIn} />}
          <main className="content">
            <Routes>
              <Route path="/" element={isLoggedIn ? <Navigate to="/home" /> : <Navigate to="/login" />} />
              <Route path="/login" element={<Login onSuccessfulLogin={handleSuccessfulLogin} />} />
              <Route path="/home" element={isLoggedIn ? <Home /> : <Navigate to="/login" />} />
              <Route path="/payloads" element={isLoggedIn ? <Payloads /> : <Navigate to="/login" />} />
            </Routes>
          </main>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
