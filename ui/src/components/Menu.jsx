import React, { useState, useContext } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import AuthContext from '../context/AuthProvider';

const SideMenu = ({ setIsLoggedIn }) => {
    const [isOpen, setIsOpen] = useState(true);
    const { logout } = useContext(AuthContext);
    const navigate = useNavigate();

    const toggleMenu = () => {
        setIsOpen(!isOpen);
    };

    const handleLogout = () => {
        logout();
        setIsLoggedIn(false); // Update the isLoggedIn state
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
                        <button className="menu-button" onClick={handleLogout}>Logout</button>
                    </li>
                </ul>
            </nav>
        </div>
    );
};

export default SideMenu;
