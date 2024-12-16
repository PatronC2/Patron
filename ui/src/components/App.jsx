import React, { useState, useEffect, useContext } from 'react';
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
  useLocation,
} from 'react-router-dom';
import SideMenu from './Menu/Menu';
import Header from './Header/Header';
import Login from './Login/Login';
import Home from './Home/Home';
import Payloads from './Payloads/Payloads';
import Redirectors from './Redirectors/Redirectors';
import Profile from './Profile/Profile';
import Users from './Users/Users';
import Agent from './Agent/Agent';
import Actions from './Actions/Actions';
import EditAction from './Actions/EditAction';
import Events from './Events/Events';
import EditEvent from './Events/EditEvent';
import Triggers from './Triggers/Triggers';
import { AuthProvider } from '../context/AuthProvider';
import AuthContext from '../context/AuthProvider';

function App() {
  const { auth } = useContext(AuthContext);
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  const [isMenuOpen, setIsMenuOpen] = useState(() => {
    const savedState = localStorage.getItem('isMenuOpen');
    return savedState === 'true';
  });

  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setIsLoggedIn(!!auth?.token);
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
          {isLoggedIn && (
            <>
              <SideMenu
                setIsLoggedIn={setIsLoggedIn}
                isOpen={isMenuOpen}
                setIsOpen={setIsMenuOpen}
              />
              <Header />
            </>
          )}
          <div className={isLoggedIn ? `main-content ${isMenuOpen ? 'menu-open' : ''}` : ''}>
            <MainContent
              isLoggedIn={isLoggedIn}
              onSuccessfulLogin={handleSuccessfulLogin}
            />
          </div>
        </div>
      </Router>
    </AuthProvider>
  );
}

const MainContent = ({ isLoggedIn, onSuccessfulLogin, isMenuOpen }) => {
  const location = useLocation();
  const isLoginPage = location.pathname === '/login';

  return (
    <>
      {!isLoginPage && !isLoggedIn && <Header />}
      <Routes>
        <Route
          path="/"
          element={
            isLoggedIn ? <Navigate to="/home" /> : <Navigate to="/login" />
          }
        />
        <Route
          path="/login"
          element={<Login onSuccessfulLogin={onSuccessfulLogin} />}
        />
        <Route
          path="/home"
          element={
            isLoggedIn ? (
              <Home isMenuOpen={isMenuOpen} />
            ) : (
              <Navigate to="/login" />
            )
          }
        />
        <Route
          path="/payloads"
          element={
            isLoggedIn ? (
              <Payloads isMenuOpen={isMenuOpen} />
            ) : (
              <Navigate to="/login" />
            )
          }
        />
        <Route
          path="/redirectors"
          element={
            isLoggedIn ? (
              <Redirectors isMenuOpen={isMenuOpen} />
            ) : (
              <Navigate to="/login" />
            )
          }
        />
        <Route
          path="/profile"
          element={
            isLoggedIn ? (
              <Profile isMenuOpen={isMenuOpen} />
            ) : (
              <Navigate to="/login" />
            )
          }
        />
        <Route
          path="/users"
          element={
            isLoggedIn ? (
              <Users isMenuOpen={isMenuOpen} />
            ) : (
              <Navigate to="/login" />
            )
          }
        />
        <Route
          path="/agent"
          element={
            isLoggedIn ? (
              <Agent isMenuOpen={isMenuOpen} />
            ) : (
              <Navigate to="/login" />
            )
          }
        />
        <Route
          path="/actions"
          element={
            isLoggedIn ? (
              <Actions isMenuOpen={isMenuOpen} />
            ) : (
              <Navigate to="/login" />
            )
          }
        />
        <Route
            path="/actions/edit"
            element={
                isLoggedIn ? (
                    <EditAction isMenuOpen={isMenuOpen} />
                ) : (
                    <Navigate to="/login" />
                )
            }
        />
        <Route
          path="/events"
          element={
            isLoggedIn ? (
              <Events isMenuOpen={isMenuOpen} />
            ) : (
              <Navigate to="/login" />
            )
          }
        />
        <Route
            path="/events/edit"
            element={
                isLoggedIn ? (
                    <EditEvent isMenuOpen={isMenuOpen} />
                ) : (
                    <Navigate to="/login" />
                )
            }
        />
        <Route
          path="/triggers"
          element={
            isLoggedIn ? (
              <Triggers isMenuOpen={isMenuOpen} />
            ) : (
              <Navigate to="/login" />
            )
          }
        />
      </Routes>
    </>
  );
};

export default App;
