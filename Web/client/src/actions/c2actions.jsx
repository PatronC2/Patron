import axios from 'axios';

const C2_ENDPOINT = 'http://localhost:3001/api';

export var getCallbacks = async () => {

    const request = await axios.get(`${C2_ENDPOINT}/agents`, { withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var getAgent = async (id) => {

    const request = await axios.get(`${C2_ENDPOINT}/agent/${id}`, { withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}

export var sendCommand = async (id,command) => {

    const request = await axios.post(`${C2_ENDPOINT}/agent/${id}`,command, {headers: {
        'Content-Type': 'application/json'
      }, withCredentials: true })
        .then(response => response.data);
    console.log(request)
    return {
        payload: request
    }

}