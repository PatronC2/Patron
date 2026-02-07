import React, { useState } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import './ProfileWizard.css';

const ApiKeyWizard = ({ username, onClose }) => {
    const axios = useAxios();
    const [loading, setLoading] = useState(false);
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [apiKey, setApiKey] = useState('');
    const [formData, setFormData] = useState({
        password: '',
        duration: '',
    });

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const setNotice = (msg, type) => {
        setNotification(msg);
        setNotificationType(type);
        setTimeout(() => setNotification(''), 3000);
    };

    const handleSubmit = async () => {
        setLoading(true);
        try {
            const response = await axios.post('/api/login', {
                username,
                password: formData.password,
                duration: parseInt(formData.duration, 10),
            });
            setApiKey(response.data.token);
            setNotice('API key generated successfully!', 'success');
        } catch (error) {
            const msg = error.response?.data?.error || 'Failed to generate API key';
            setNotice(msg, 'error');
        } finally {
            setLoading(false);
        }
    };

    const handleCopy = () => {
        navigator.clipboard.writeText(apiKey).then(() => {
            setNotice('API key copied to clipboard!', 'success');
        }).catch(() => {
            setNotice('Failed to copy API key.', 'error');
        });
    };

    return (
        <div className="wizard-overlay" onClick={onClose}>
            <div className="wizard-card" onClick={(e) => e.stopPropagation()}>
                <div className="wizard-header">
                    <div>
                        <h2>Generate API Key</h2>
                        <div className="wizard-steps">
                            <span className="active">API Key</span>
                        </div>
                    </div>
                    <button type="button" className="wizard-close" onClick={onClose}>Close</button>
                </div>
                <div className="wizard-body">
                    <div className="wizard-form">
                        <label>Password</label>
                        <input
                            type="password"
                            name="password"
                            value={formData.password}
                            onChange={handleChange}
                            required
                        />
                        <label>Duration (hours)</label>
                        <input
                            type="number"
                            name="duration"
                            value={formData.duration}
                            onChange={handleChange}
                            required
                        />
                        <button type="button" onClick={handleSubmit} disabled={loading}>
                            {loading ? 'Generating...' : 'Generate API Key'}
                        </button>
                    </div>
                    {apiKey && (
                        <div className="wizard-api-key">
                            <p>Your API Key (copy it now, as it won’t be shown again):</p>
                            <pre>{apiKey}</pre>
                            <button type="button" onClick={handleCopy}>Copy to Clipboard</button>
                        </div>
                    )}
                </div>
                {notification && (
                    <div className={`notification ${notificationType}`}>
                        {notification}
                    </div>
                )}
            </div>
        </div>
    );
};

export default ApiKeyWizard;
