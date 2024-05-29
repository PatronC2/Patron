import axios from 'axios';

const API_ENDPOINT = `http://${process.env.APP_API_HOST}:${process.env.APP_API_PORT}`

const instance = axios.create({
    // I have no idea why using the .env does not work
    baseURL: `http://192.168.50.32:8000`
});

export default instance;
