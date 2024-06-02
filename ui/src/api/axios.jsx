import axios from 'axios';

const API_ENDPOINT = `https://${process.env.REACT_APP_NGINX_IP}:${process.env.REACT_APP_NGINX_PORT}`;

const instance = axios.create({
    baseURL: API_ENDPOINT,
});

instance.interceptors.request.use(
    (config) => {
        const auth = JSON.parse(localStorage.getItem('auth'));
        if (auth?.token) {
            config.headers['Authorization'] = `${auth.token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

export default instance;
