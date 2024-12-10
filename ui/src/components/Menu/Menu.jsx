import React, { useContext, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { FaBars, FaChevronLeft } from 'react-icons/fa';
import AuthContext from '../../context/AuthProvider';
import './Menu.css';

const SideMenu = ({ setIsLoggedIn, isOpen, setIsOpen }) => {
  const { logout } = useContext(AuthContext);
  const navigate = useNavigate();

  const toggleMenu = () => {
    const newState = !isOpen;
    setIsOpen(newState);
    localStorage.setItem('isMenuOpen', newState);
  };

  const handleLogout = () => {
    logout();
    setIsLoggedIn(false);
    navigate('/login');
  };

  useEffect(() => {
    return () => document.body.classList.remove('menu-open');
  }, []);

  return (
        <div className="side-menu-container">
            <button
                className={`toggle-button ${isOpen ? 'open' : ''}`}
                aria-label="Toggle Menu"
                onClick={toggleMenu}
            >
                {isOpen ? <FaChevronLeft size={20} /> : <FaBars size={20} />}
            </button>
            <div className={`side-menu ${isOpen ? 'open' : ''}`}>
                <nav className="menu-nav" role="navigation" aria-label="Main Navigation">
                    <ul>
                        <li>
                            <Link to="/home">Home</Link>
                        </li>
                        <li>
                            <Link to="/payloads">Payloads</Link>
                        </li>
                        <li>
                            <Link to="/redirectors">Redirectors</Link>
                        </li>
                        <li>
                            <Link to="/events">Events</Link>
                        </li>
                        <li>
                            <Link to="/triggers">Triggers</Link>
                        </li>
                        <li>
                            <Link to="/actions">Actions</Link>
                        </li>
                        <li>
                            <Link to="/profile">Profile</Link>
                        </li>
                        <li>
                            <Link to="/users">Admin</Link>
                        </li>
                        <li>
                            <button className="menu-button" onClick={handleLogout}>
                                Logout
                            </button>
                        </li>
                    </ul>
                </nav>
            </div>
        </div>
    );
};

export default SideMenu;
