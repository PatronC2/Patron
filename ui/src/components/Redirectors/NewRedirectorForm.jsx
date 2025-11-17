import React, { useState, useContext, useMemo } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import './NewRedirectorForm.css';

const NewRedirectorForm = ({ fetchData, setActiveTab, redirectors }) => {
    const cfg = window.runtimeConfig;
    const PATRON_C2_IP = `${cfg.REACT_APP_NGINX_IP}`;
    const PATRON_C2_PORT = `${cfg.REACT_APP_C2SERVER_PORT}`;
    const axios = useAxios();
    const { auth } = useContext(AuthContext);
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [selectedForwardTargetId, setSelectedForwardTargetId] = useState('');
    const [formData, setFormData] = useState({
        Name: '',
        Description: '',
        ForwardIP: `${PATRON_C2_IP}`,
        ForwardPort: `${PATRON_C2_PORT}`,
        ListenIP: `x.x.x.x`,
        ListenPort: `${PATRON_C2_PORT}`,
    });
    const [loading, setLoading] = useState(false);

    // Only online redirectors with a valid listener port
    const onlineTargets = useMemo(
        () =>
            (redirectors || []).filter(
                r => r.status === 'Online' && r.transportprotocol === 'tcp' && r.listenport && r.listenport !== ''
            ),
        [redirectors]
    );

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleForwardTargetChange = (e) => {
        const selectedId = e.target.value;
        setSelectedForwardTargetId(selectedId);
        
        if (!selectedId) return;

        const target = onlineTargets.find(r => r.id === selectedId);
        if (!target) return;

        // Use the *listener's* IP/port as ForwardIP/ForwardPort for the new redirector
        setFormData(prev => ({
            ...prev,
            ForwardIP: target.listenip,
            ForwardPort: target.listenport,
        }));
    };

    const handleNotification = (message, type) => {
        setNotification(message);
        setNotificationType(type);
        setTimeout(() => {
            setNotification('');
            setNotificationType('');
        }, 3000);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        const url = `/api/redirector`;
        
        setLoading(true);
        
        try {
            const response = await axios.post(url, formData, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `${auth.accessToken}`,
                },
                responseType: 'blob',
            });
    
            if (response.status === 200) {
                const blob = new Blob([response.data], { type: response.headers['content-type'] });
                const downloadUrl = URL.createObjectURL(blob);
    
                const link = document.createElement('a');
                link.href = downloadUrl;
                link.download = 'redirector_install.sh';
                document.body.appendChild(link);
                link.click();
                document.body.removeChild(link);
    
                URL.revokeObjectURL(downloadUrl);
    
                handleNotification('Redirector created successfully! Install Script downloading.', 'success');
                fetchData();
                setTimeout(() => {
                    setActiveTab('current_redirectors');
                }, 3000);
            } else {
                throw new Error(`Unexpected status code: ${response.status}`);
            }
        } catch (error) {
            if (error.response) {
                if (error.response.status === 401) {
                    handleNotification('Error: Unauthorized.', 'error');
                } else {
                    console.error(`Failed to compile: ${error.response.data}`);
                    handleNotification(`Failed to compile: ${error.response.data}`, 'error');
                }
            } else if (error.request) {
                console.error('Error: No response received from server.');
                handleNotification('Error: No response received from server.', 'error');
            } else {
                console.error(`Error: ${error.message}`);
                handleNotification(`Error: ${error.message}`, 'error');
            }
        } finally {
            setLoading(false);
        }
    };    

    return (
        <div className="redirector-form-container">
            {loading && (
                <div className="loading-indicator">
                    <span>Loading...</span>
                </div>
            )}
            <form onSubmit={handleSubmit}>
                <div>
                    <label htmlFor="Name">Redirector Name:</label>
                    <input
                        type="text"
                        id="Name"
                        name="Name"
                        value={formData.Name}
                        onChange={handleChange}
                        aria-label="Redirector Name"
                        placeholder="Enter the name of the redirector"
                    />
                </div>
                <div>
                    <label htmlFor="Description">Description:</label>
                    <textarea
                        id="Description"
                        name="Description"
                        value={formData.Description}
                        onChange={handleChange}
                        aria-label="Redirector Description"
                        placeholder="Enter a brief description"
                    />
                </div>
                <div>
                    <label htmlFor="ForwardTarget">
                        Forward To (online redirector / listener):
                    </label>
                    <select
                        id="ForwardTarget"
                        onChange={handleForwardTargetChange}
                        value={selectedForwardTargetId}
                    >
                        <option value="">
                            -- Select an online redirector (optional) --
                        </option>
                        {onlineTargets.map(r => (
                            <option key={`${r.id}-${r.listenport}`} value={r.id}>
                                {r.name} — {r.listenip}:{r.listenport}
                                {r.transportprotocol
                                    ? ` (${r.transportprotocol.toUpperCase()})`
                                    : ''}
                            </option>
                        ))}
                    </select>
                </div>

                <div>
                    <label htmlFor="ForwardIP">Forward IP:</label>
                    <input
                        type="text"
                        id="ForwardIP"
                        name="ForwardIP"
                        value={formData.ForwardIP}
                        onChange={handleChange}
                        aria-label="Forward IP Address"
                        placeholder="Enter the Forward IP"
                    />
                </div>
                <div>
                    <label htmlFor="ForwardPort">Forward Port:</label>
                    <input
                        type="text"
                        id="ForwardPort"
                        name="ForwardPort"
                        value={formData.ForwardPort}
                        onChange={handleChange}
                        aria-label="Forward Port"
                        placeholder="Enter the Forward Port"
                    />
                </div>
                <div>
                    <label htmlFor="ListenIP">Listen IP:</label>
                    <input
                        type="text"
                        id="ListenIP"
                        name="ListenIP"
                        value={formData.ListenIP}
                        onChange={handleChange}
                        aria-label="Listen IP"
                        placeholder="Enter the new host's IP"
                    />
                </div>
                <div>
                    <label htmlFor="ListenPort">Listen Port:</label>
                    <input
                        type="text"
                        id="ListenPort"
                        name="ListenPort"
                        value={formData.ListenPort}
                        onChange={handleChange}
                        aria-label="Listen Port"
                        placeholder="Enter the Listen Port"
                    />
                </div>
                <button type="submit" disabled={loading}>
                    {loading ? 'Creating...' : 'Create Redirector'}
                </button>
                {notification && (
                    <div className={`notification ${notificationType}`}>
                        {notification}
                    </div>
                )}
            </form>
        </div>
    );
};

export default NewRedirectorForm;
