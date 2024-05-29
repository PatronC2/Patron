import { createContext, useState, useEffect } from 'react';

const AuthContext = createContext({});

export const AuthProvider = ({ children }) => {
    const [auth, setAuth] = useState({});

    useEffect(() => {
        // Load token from localStorage when the component mounts
        const storedAuth = localStorage.getItem('auth');
        if (storedAuth) {
            setAuth(JSON.parse(storedAuth));
        }
    }, []);

    const setAuthWithLocalStorage = (authData) => {
        setAuth(authData);
        if (authData.token) {
            localStorage.setItem('auth', JSON.stringify(authData));
        } else {
            localStorage.removeItem('auth');
        }
    };

    const logout = () => {
        setAuthWithLocalStorage({}); // Clear auth state and remove from localStorage
    };

    return (
        <AuthContext.Provider value={{ auth, setAuth: setAuthWithLocalStorage, logout }}>
            {children}
        </AuthContext.Provider>
    )
}

export default AuthContext;
