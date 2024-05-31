import React, { useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import './NewPayloadForm.css'

const NewPayloadForm = ({ fetchData, setActiveTab }) => {
    const { auth } = useContext(AuthContext);
    const [notification, setNotification] = useState('');
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        type: 'original',
        serverip: '',
        serverport: '',
        callbackfrequency: '',
        callbackjitter: '',
    });

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData({ ...formData, [name]: value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        const url = `http://${process.env.REACT_APP_API_HOST}:${process.env.REACT_APP_API_PORT}/api/payload`;
        try {
            const response = await axios.post(url, formData, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `${auth.accessToken}`,
                },
            });

            if (response.status !== 200) {
                throw new Error(`Failed to compile: ${response.data}`);
            }

            setNotification('Payload created successfully!');
            fetchData();
            setTimeout(() => {
                setActiveTab('current_payloads');
            }, 3000);
        } catch (error) {
            console.error(`Failed to make compile request: ${error.message}`);
            setNotification(`Failed to compile payload ${error.message}`);
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <div>
                <label htmlFor="name">Payload Name:</label>
                <input type="text" id="name" name="name" value={formData.name} onChange={handleChange} />
            </div>
            <div>
                <label htmlFor="description">Description:</label>
                <textarea id="description" name="description" value={formData.description} onChange={handleChange} />
            </div>
            <div>
                <label htmlFor="type">Type:</label>
                <select id="type" name="type" value={formData.type} onChange={handleChange}>
                    <option value="wkeys">Keylogger (requires root)</option>
                    <option value="original">No Keylogger</option>
                </select>
            </div>
            <div>
                <label htmlFor="serverip">Listener IP:</label>
                <input type="text" id="serverip" name="serverip" value={formData.serverip} onChange={handleChange} />
            </div>
            <div>
                <label htmlFor="serverport">Listener Port:</label>
                <input type="text" id="serverport" name="serverport" value={formData.serverport} onChange={handleChange} />
            </div>
            <div>
                <label htmlFor="callbackfrequency">Call Back Frequency:</label>
                <input type="text" id="callbackfrequency" name="callbackfrequency" value={formData.callbackfrequency} onChange={handleChange} />
            </div>
            <div>
                <label htmlFor="callbackjitter">Call Back Jitter:</label>
                <input type="text" id="callbackjitter" name="callbackjitter" value={formData.callbackjitter} onChange={handleChange} />
            </div>
            <button type="submit">Create Payload</button>
            {notification && <div className="notification">{notification}</div>}
        </form>
    );
};

export default NewPayloadForm;
