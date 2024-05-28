import axios from 'axios';

const instance = axios.create({
    baseURL: 'http://192.168.50.32:8000'
});

export default instance;
