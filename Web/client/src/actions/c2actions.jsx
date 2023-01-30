import axios from 'axios';
const C2_ENDPOINT = process.env.REACT_APP_NGINX_IP ? `http://${process.env.REACT_APP_NGINX_IP}:${process.env.REACT_APP_NGINX_PORT}` : `http://${process.env.REACT_APP_WEBSERVER_IP}:${process.env.REACT_APP_WEBSERVER_PORT}`;
export { C2_ENDPOINT };

export var getCallbacks = async () => {

    const request = await axios.get(`${C2_ENDPOINT}/api/agents`, { withCredentials: true })
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