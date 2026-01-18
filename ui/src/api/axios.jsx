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

    instance.interceptors.response.use(
        (response) => response,
        (error) => {
            if (error.response?.status === 401) {
                // Clear stale auth
                localStorage.removeItem('auth');

                // Avoid infinite loops if we're already on /login
                if (window.location.pathname !== '/login') {
                    window.location.href = '/login';
                }
            }

            return Promise.reject(error);
        }
    );

    return instance;
};
