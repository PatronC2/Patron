import React, { useState, useContext } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import './ApiKey.css';

const ApiKeyForm = ({ username }) => {
    const axios = useAxios();
    const { auth } = useContext(AuthContext);
    const [formData, setFormData] = useState({
        password: '',
        duration: '',
    });
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [apiKey, setApiKey] = useState('');

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData({ ...formData, [name]: value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        const { password, duration } = formData;

        try {
            const response = await axios.post('/api/login', {
                username,
                password,
                duration: parseInt(duration, 10),
            });

            setNotification('API key generated successfully!');
            setNotificationType('success');
            setApiKey(response.data.token);

            setTimeout(() => setNotification(''), 3000);
        } catch (error) {
            const errorMessage = error.response?.data?.error || 'Failed to generate API key';
            setNotification(errorMessage);
            setNotificationType('error');
            setTimeout(() => setNotification(''), 3000);
        }
    };

    const handleCopy = () => {
        navigator.clipboard.writeText(apiKey).then(() => {
            setNotification('API key copied to clipboard!');
            setNotificationType('success');
            setTimeout(() => setNotification(''), 3000);
        }).catch(() => {
            setNotification('Failed to copy API key.');
            setNotificationType('error');
            setTimeout(() => setNotification(''), 3000);
        });
    };

    return (
        <div className="api-key-form-container">
            <form onSubmit={handleSubmit}>
                <div>
                    <label htmlFor="password">Password:</label>
                    <input
                        type="password"
                        id="password"
                        name="password"
                        value={formData.password}
                        onChange={handleChange}
                        required
                    />
                </div>
                <div>
                    <label htmlFor="duration">Duration (hours):</label>
                    <input
                        type="number"
                        id="duration"
                        name="duration"
                        value={formData.duration}
                        onChange={handleChange}
                        required
                    />
                </div>
                <button type="submit">Generate API Key</button>
                {notification && (
                    <div className={`notification ${notificationType}`}>
                        {notification}
                    </div>
                )}
            </form>
            {apiKey && (
                <div className="api-key-display">
                    <p>Your API Key (copy it now, as it won’t be shown again):</p>
                    <pre>{apiKey}</pre>
                    <button onClick={handleCopy}>Copy to Clipboard</button>
                </div>
            )}
        </div>
    );
};

export default ApiKeyForm;
