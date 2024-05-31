import React, { useState, useContext, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import SideMenu from './Menu/Menu';
import Header from './Header/Header';
import Login from './Login/Login';
import Home from './Home/Home';
import Payloads from './Payloads/Payloads';
import Users from './Users/Users';
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
          <MainContent isLoggedIn={isLoggedIn} onSuccessfulLogin={handleSuccessfulLogin} />
        </div>
      </Router>
    </AuthProvider>
  );
}

const MainContent = ({ isLoggedIn, onSuccessfulLogin }) => {
  const location = useLocation();
  const isLoginPage = location.pathname === '/login';

  return (
    <div className="main-content">
      {!isLoginPage && <Header />}
      <Routes>
        <Route path="/" element={isLoggedIn ? <Navigate to="/home" /> : <Navigate to="/login" />} />
        <Route path="/login" element={<Login onSuccessfulLogin={onSuccessfulLogin} />} />
        <Route path="/home" element={isLoggedIn ? <Home /> : <Navigate to="/login" />} />
        <Route path="/payloads" element={isLoggedIn ? <Payloads /> : <Navigate to="/login" />} />
        <Route path="/users" element={isLoggedIn ? <Users /> : <Navigate to="/login" />} />
      </Routes>
    </div>
  );
};

export default App;
