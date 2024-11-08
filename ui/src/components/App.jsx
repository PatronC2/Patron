import React, { useState, useEffect, useContext } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import SideMenu from './Menu/Menu';
import Header from './Header/Header';
import Login from './Login/Login';
import Home from './Home/Home';
import Payloads from './Payloads/Payloads';
import Redirectors from './Redirectors/Redirectors';
import Profile from './Profile/Profile';
import Users from './Users/Users';
import Agent from './Agent/Agent';
import { AuthProvider } from '../context/AuthProvider';
import AuthContext from '../context/AuthProvider';

function App() {
  const { auth } = useContext(AuthContext);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setIsLoggedIn(!!auth.token);
    setLoading(false);
  }, [auth]);

  const handleSuccessfulLogin = () => {
    setIsLoggedIn(true);
  };

  if (loading) {
    return <div>Loading...</div>;
  }

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
        <Route path="/redirectors" element={isLoggedIn ? <Redirectors /> : <Navigate to="/redirectors" />} />
        <Route path="/profile" element={isLoggedIn ? <Profile /> : <Navigate to="/profile" />} />
        <Route path="/users" element={isLoggedIn ? <Users /> : <Navigate to="/login" />} />
        <Route path="/agent" element={isLoggedIn ? <Agent /> : <Navigate to="/agent" />} />
      </Routes>
    </div>
  );
};

export default App;
