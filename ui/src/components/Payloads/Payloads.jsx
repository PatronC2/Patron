import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import NewPayloadForm from './NewPayloadForm';
import './Payloads.css';

const FILE_SERVER = `https://${process.env.REACT_APP_NGINX_IP}:${process.env.REACT_APP_NGINX_PORT}/fileserver/`;

const Payloads = () => {
    const { auth } = useContext(AuthContext);
    const [data, setData] = useState([]);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('current_payloads');
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');

    useEffect(() => {
        document.body.classList.add('payloads-page');
        fetchData();
        const interval = setInterval(() => {
            fetchData();
        }, 10000);

        return () => {
            document.body.classList.remove('payloads-page');
            clearInterval(interval);
        };
    }, [auth.accessToken]);

    const fetchData = async () => {
        try {
            const response = await axios.get('/api/payloads', {
                headers: {
                    'Authorization': `${auth.accessToken}`,
                },
            });

            const responseData = response.data.data;
            if (Array.isArray(responseData)) {
                setData(responseData);
            } else {
                setData('');
            }
        } catch (err) {
            setError(err.message);
        }
    };

    const handleDelete = async (payloadid) => {
        try {
            await axios.delete(`/api/payloads/${payloadid}`, {
                headers: {
                    'Authorization': `${auth.accessToken}`,
                },
            });

            setNotification('Payload deleted successfully!');
            setNotificationType('success');
            fetchData();
        } catch (err) {
            setNotification('Failed to delete payload.');
            setNotificationType('error');
        } finally {
            setTimeout(() => setNotification(''), 3000);
        }
    };

    const handleTabChange = (tab) => {
        setActiveTab(tab);
    };

    if (error) {
        return <div>Error: {error}</div>;
    }

    return (
        <div className="payloads-content">
            <div className="header">
                <h1>Payloads</h1>
                <div className="header-buttons">
                    <button
                        className={activeTab === 'current_payloads' ? 'active' : ''}
                        onClick={() => handleTabChange('current_payloads')}
                    >
                        Existing Payloads
                    </button>
                    <button
                        className={activeTab === 'new' ? 'active' : ''}
                        onClick={() => handleTabChange('new')}
                    >
                        Create New Payload
                    </button>
                </div>
            </div>
            {activeTab === 'current_payloads' ? (
                <div className="payloads-container">
                    {data.length > 0 ? (
                        <table>
                            <thead>
                                <tr>
                                    <th>UUID</th>
                                    <th>Name</th>
                                    <th>Description</th>
                                    <th>Listener IP</th>
                                    <th>Listener Port</th>
                                    <th>Callback Frequency</th>
                                    <th>Callback Jitter</th>
                                    <th>Action</th>
                                </tr>
                            </thead>
                            <tbody>
                                {data.map((item) => (
                                    <tr key={item.uuid}>
                                        <td>{item.uuid.substring(0, 6)}</td>
                                        <td>
                                            <a
                                                href={`${FILE_SERVER}${item.concat}`}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                            >
                                                {item.concat}
                                            </a>
                                        </td>
                                        <td>{item.description}</td>
                                        <td>{item.serverip}</td>
                                        <td>{item.serverport}</td>
                                        <td>{item.callbackfrequency}</td>
                                        <td>{item.callbackjitter}</td>
                                        <td>
                                            <button
                                                onClick={() => handleDelete(item.payloadid)}
                                                className="delete-button"
                                            >
                                                Delete
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    ) : (
                        <p>No Payloads</p>
                    )}
                </div>
            ) : (
                <div>
                    <NewPayloadForm fetchData={fetchData} setActiveTab={setActiveTab} />
                </div>
            )}
            {notification && (
                <div className={`notification ${notificationType}`}>
                    {notification}
                </div>
            )}
        </div>
    );
};

export default Payloads;
