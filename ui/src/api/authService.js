import axios from 'axios';
import { C2_ENDPOINT } from './config'; // Import the API endpoint

const loginUser = async (username, password) => {
    try {
        const response = await axios.post(`${C2_ENDPOINT}/login`, { username, password });
        const { token } = response.data;
        // Store token in local storage or session storage
        localStorage.setItem('authToken', token);
        return token;
    } catch (error) {
        throw new Error('Login failed. Please check your credentials.');
    }
};

export { loginUser };
