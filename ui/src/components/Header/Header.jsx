import React, { useContext } from 'react';
import './Header.css';
import ThemeContext from '../../context/Themes';
import logo from '../../assets/images/patron.png';

const Header = () => {
    const { theme, toggleTheme } = useContext(ThemeContext);

    return (
        <header className="app-header">
            <div className="shared-container header-content">
                <div className="center-container">
                    <h1 className="app-name">Patron C2</h1>
                    <img src={logo} alt="App Logo" className="app-logo" />
                </div>
                <div className="theme-slider-container">
                    <label className="theme-slider">
                        <input
                            type="checkbox"
                            onChange={toggleTheme}
                            checked={theme === 'dark'}
                        />
                        <span className="slider"></span>
                    </label>
                </div>
            </div>
        </header>
    );
};

export default Header;
