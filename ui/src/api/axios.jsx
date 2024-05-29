import axios from 'axios';

const API_ENDPOINT = `http://${process.env.REACT_APP_API_HOST}:${process.env.REACT_APP_API_PORT}`

const instance = axios.create({
    baseURL: `${API_ENDPOINT}`
});

export default instance;
