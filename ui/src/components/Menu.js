import React, { useState } from 'react';
import { Link } from 'react-router-dom';

const SideMenu = () => {
    const [isOpen, setIsOpen] = useState(true);

    const toggleMenu = () => {
        setIsOpen(!isOpen);
    };

    return (
        <div className={`side-menu ${isOpen ? 'open' : ''}`}>
            <button className="toggle-button" onClick={toggleMenu}>
                {isOpen ? '<' : '>'}
            </button>
            <nav className="menu-nav">
                <ul>
                    <li>
                        <Link to="/">Home</Link>
                    </li>
                    <li>
                        <Link to="/payloads">Payloads</Link>
                    </li>
                </ul>
            </nav>
        </div>
    );
};

export default SideMenu;
