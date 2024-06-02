import { createContext, useState, useEffect } from 'react';

const AuthContext = createContext({});

export const AuthProvider = ({ children }) => {
    const [auth, setAuth] = useState(() => {
        const storedAuth = localStorage.getItem('auth');
        return storedAuth ? JSON.parse(storedAuth) : {};
    });

    useEffect(() => {
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
        setAuthWithLocalStorage({});
    };

    return (
        <AuthContext.Provider value={{ auth, setAuth: setAuthWithLocalStorage, logout }}>
            {children}
        </AuthContext.Provider>
    );
};

export default AuthContext;
