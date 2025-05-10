import axios from 'axios';

export const createAxios = () => {
    const cfg = window.runtimeConfig;
    if (!cfg) {
        throw new Error('Runtime config not loaded');
    }

    const API_ENDPOINT = `https://${cfg.REACT_APP_NGINX_IP}:${cfg.REACT_APP_NGINX_PORT}`;

    const instance = axios.create({ baseURL: API_ENDPOINT });

    instance.interceptors.request.use(
        (config) => {
            const auth = JSON.parse(localStorage.getItem('auth'));
            if (auth?.token) {
                config.headers['Authorization'] = `${auth.token}`;
            }
            return config;
        },
        (error) => Promise.reject(error)
    );

    return instance;
};
