import React, { createContext, useContext, useMemo } from 'react';
import { createAxios } from '../api/axios';

const AxiosContext = createContext(null);

export const AxiosProvider = ({ children }) => {
    const axiosInstance = useMemo(() => {
        return createAxios();
    }, []);

    return (
        <AxiosContext.Provider value={axiosInstance}>
            {children}
        </AxiosContext.Provider>
    );
};

export const useAxios = () => {
    const context = useContext(AxiosContext);
    if (!context) {
        throw new Error('useAxios must be used within an AxiosProvider');
    }
    return context;
};
