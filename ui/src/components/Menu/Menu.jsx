import React, { useState, useContext } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import AuthContext from '../../context/AuthProvider';
import './Menu.css';

const SideMenu = ({ setIsLoggedIn }) => {
    const [isOpen, setIsOpen] = useState(true);
    const { logout } = useContext(AuthContext);
    const navigate = useNavigate();

    const toggleMenu = () => {
        setIsOpen(!isOpen);
        document.body.classList.toggle('menu-open', isOpen);
    };

    const handleLogout = () => {
        logout();
        setIsLoggedIn(false);
        navigate('/login');
    };

    return (
        <div className={`side-menu ${isOpen ? 'open' : ''}`}>
            <button className="toggle-button" onClick={toggleMenu}>
                {isOpen ? '<' : '>'}
            </button>
            <nav className="menu-nav">
                <ul>
                    <li>
                        <Link to="/home">Home</Link>
                    </li>
                    <li>
                        <Link to="/payloads">Payloads</Link>
                    </li>
                    <li>
                        <Link to="/profile">Profile</Link>
                    </li>
                    <li>
                        <Link to="/users">Admin</Link>
                    </li>
                    <li>
                        <button className="menu-button" onClick={handleLogout}>Logout</button>
                    </li>
                </ul>
            </nav>
        </div>
    );
};

export default SideMenu;
