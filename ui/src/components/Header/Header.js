// Header.js
import React from 'react';
import './Header.css';
import logo from '../../assets/images/patron.png'; // Correctly import the image

const Header = () => {
    return (
        <header className="app-header">
            <div className="header-content">
                <div className="center-container">
                    <h1 className="app-name">Patron C2</h1>
                    <img src={logo} alt="App Logo" className="app-logo" />
                </div>
            </div>
        </header>
    );
};

export default Header;
