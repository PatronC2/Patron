import React, { useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import './NewRedirectorForm.css';

const PATRON_C2_IP = `${process.env.REACT_APP_NGINX_IP}`;
const PATRON_C2_PORT = `${process.env.REACT_APP_C2SERVER_PORT}`;

const NewRedirectorForm = ({ fetchData, setActiveTab }) => {
    const { auth } = useContext(AuthContext);
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [formData, setFormData] = useState({
        Name: '',
        Description: '',
        ForwardIP: `${PATRON_C2_IP}`,
        ForwardPort: `${PATRON_C2_PORT}`,
        ListenIP: `0.0.0.0`,
        ListenPort: `${PATRON_C2_PORT}`,
    });
    const [loading, setLoading] = useState(false);

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData({ ...formData, [name]: value });
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
    
                setNotification('Redirector created successfully! Install Script downloading.');
                setNotificationType('success');
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
                    setNotification('Error: Unauthorized.');
                    setNotificationType('error');
                } else {
                    console.error(`Failed to compile: ${error.response.data}`);
                    setNotification(`Failed to compile: ${error.response.data}`);
                    setNotificationType('error');
                }
            } else if (error.request) {
                console.error('Error: No response received from server.');
                setNotification('Error: No response received from server.');
                setNotificationType('error');
            } else {
                console.error(`Error: ${error.message}`);
                setNotification(`Error: ${error.message}`);
                setNotificationType('error');
            }
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            {loading && <div className="loading-indicator">Loading...</div>} {/* Loading indicator */}
            <form onSubmit={handleSubmit}>
                <div>
                    <label htmlFor="Name">Redirector Name:</label>
                    <input type="text" id="Name" name="Name" value={formData.Name} onChange={handleChange} />
                </div>
                <div>
                    <label htmlFor="Description">Description:</label>
                    <textarea id="Description" name="Description" value={formData.Description} onChange={handleChange} />
                </div>
                <div>
                    <label htmlFor="ForwardIP">Forward IP:</label>
                    <input type="text" id="ForwardIP" name="ForwardIP" value={formData.ForwardIP} onChange={handleChange} />
                </div>
                <div>
                    <label htmlFor="ForwardPort">Forward Port:</label>
                    <input type="text" id="ForwardPort" name="ForwardPort" value={formData.ForwardPort} onChange={handleChange} />
                </div>
                <div>
                    <label htmlFor="ListenIP">Listen IP:</label>
                    <input type="text" id="ListenIP" name="ListenIP" value={formData.ListenIP} onChange={handleChange} />
                </div>
                <div>
                    <label htmlFor="ListenPort">Listen Port:</label>
                    <input type="text" id="ListenPort" name="ListenPort" value={formData.ListenPort} onChange={handleChange} />
                </div>
                <button type="submit">Create Redirector</button>
                {notification && (
                    <div className={`notification ${notificationType}`}>{notification}</div>
                )}
            </form>
        </div>
    );
};

export default NewRedirectorForm;
