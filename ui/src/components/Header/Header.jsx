// Header.js
import React, { useContext } from 'react';
import './Header.css';
import ThemeContext from '../../context/Themes';
import logo from '../../assets/images/patron.png';

const Header = () => {
    const { theme, toggleTheme } = useContext(ThemeContext);
    return (
        <header className="app-header">
            <div className="header-content">
                <div className="center-container">
                    <h1 className="app-name">Patron C2</h1>
                    <img src={logo} alt="App Logo" className="app-logo" />
                </div>
                <button className="theme-toggle" onClick={toggleTheme}>
                    {theme === 'light' ? 'Dark Mode' : 'Light Mode'}
                </button>
            </div>
        </header>
    );
};

export default Header;
