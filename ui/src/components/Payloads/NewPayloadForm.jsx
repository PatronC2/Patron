import React, { useState, useEffect, useContext, useMemo } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import './NewPayloadForm.css';

const NewPayloadForm = ({ fetchData, setActiveTab }) => {
    const cfg = window.runtimeConfig;
    const PATRON_C2_IP = `${cfg.REACT_APP_NGINX_IP}`;
    const PATRON_C2_PORT = `${cfg.REACT_APP_C2SERVER_PORT}`;
    const axios = useAxios();
    const { auth } = useContext(AuthContext);
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [selectedListenerIndex, setSelectedListenerIndex] = useState('');
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        type: '',
        serverip: `${PATRON_C2_IP}`,
        serverport: `${PATRON_C2_PORT}`,
        callbackfrequency: '300',
        callbackjitter: '80',
        logging: 'false',
        transportprotocol: 'TCP',
        compression: 'none',
    });
    const [availableTypes, setAvailableTypes] = useState([]);
    const [redirectors, setRedirectors] = useState([]);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        const fetchStuff = async () => {
            try {
                const typesResp = await axios.get('/api/payloadconfs');
                const types = Object.entries(typesResp.data).map(([key, value]) => ({
                    value: key,
                    label: value.type,
                }));
                setAvailableTypes(types);
                setFormData(prevData => ({
                    ...prevData,
                    type: types[0]?.value || '',
                }));

                const redirResp = await axios.get('/api/redirectors');
                const list = Array.isArray(redirResp.data.data)
                    ? redirResp.data.data
                    : [];
                setRedirectors(list);
            } catch (error) {
                console.error('Error fetching payload types / redirectors:', error);
                setNotification('Error fetching payload types or redirectors.');
                setNotificationType('error');
            }
        };

        fetchStuff();
    }, []);

    const onlineListeners = useMemo(
        () =>
            (redirectors || []).filter(
                r => r.status === 'Online' && r.listenport && r.listenport !== ''
            ),
        [redirectors]
    );

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleListenerTargetChange = (e) => {
        const { value } = e.target;
        setSelectedListenerIndex(value);

        if (value === '') return;

        const idx = parseInt(value, 10);
        const target = onlineListeners[idx];
        if (!target) return;

        const protoFromAPI = target.transportprotocol || '';
        const protoUpper = protoFromAPI.toUpperCase();

        setFormData(prev => ({
            ...prev,
            serverip: target.listenip,
            serverport: target.listenport,
            transportprotocol: protoUpper || prev.transportprotocol,
        }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        const url = `/api/payload`;

        setLoading(true);

        try {
            const response = await axios.post(url, formData, {
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `${auth.accessToken}`,
                },
            });

            if (response.status === 200) {
                setNotification('Payload created successfully!');
                setNotificationType('success');
                fetchData();
                setTimeout(() => {
                    setActiveTab('current_payloads');
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
                    console.error(`Failed to compile: ${error.response.data.error}`);
                    setNotification(`Failed to compile: ${error.response.data.error}`);
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
        <div className="payload-form-container">
            {loading && <div className="loading-indicator">Loading...</div>}
            <form onSubmit={handleSubmit}>
                <div>
                    <label htmlFor="name">Payload Name:</label>
                    <input
                        type="text"
                        id="name"
                        name="name"
                        value={formData.name}
                        onChange={handleChange}
                        placeholder="Enter the payload name"
                    />
                </div>
                <div>
                    <label htmlFor="description">Description:</label>
                    <textarea
                        id="description"
                        name="description"
                        value={formData.description}
                        onChange={handleChange}
                        placeholder="Enter a brief description"
                    />
                </div>
                <div>
                    <label htmlFor="type">Type:</label>
                    <select
                        id="type"
                        name="type"
                        value={formData.type}
                        onChange={handleChange}
                    >
                        {availableTypes.map((type) => (
                            <option key={type.value} value={type.value}>
                                {type.label}
                            </option>
                        ))}
                    </select>
                </div>

                <div>
                    <label htmlFor="listenerTarget">
                        Use Online Redirector Listener:
                    </label>
                    <select
                        id="listenerTarget"
                        onChange={handleListenerTargetChange}
                        value=""
                    >
                        <option value="">
                            -- Optional: select an online listener --
                        </option>
                        {onlineListeners.map((r) => (
                            <option
                                key={`${r.id}-${r.listenport}-${r.transportprotocol}`}
                                value={r.idx}
                            >
                                {r.name} — {r.listenip}:{r.listenport}
                                {r.transportprotocol
                                    ? ` (${r.transportprotocol.toUpperCase()})`
                                    : ''}
                            </option>
                        ))}
                    </select>
                </div>

                <div>
                    <label htmlFor="serverip">Listener IP:</label>
                    <input
                        type="text"
                        id="serverip"
                        name="serverip"
                        value={formData.serverip}
                        onChange={handleChange}
                        placeholder="Enter the listener IP"
                    />
                </div>
                <div>
                    <label htmlFor="serverport">Listener Port:</label>
                    <input
                        type="text"
                        id="serverport"
                        name="serverport"
                        value={formData.serverport}
                        onChange={handleChange}
                        placeholder="Enter the listener port"
                    />
                </div>
                <div>
                    <label htmlFor="callbackfrequency">Call Back Frequency:</label>
                    <input
                        type="text"
                        id="callbackfrequency"
                        name="callbackfrequency"
                        value={formData.callbackfrequency}
                        onChange={handleChange}
                        placeholder="Enter callback frequency"
                    />
                </div>
                <div>
                    <label htmlFor="callbackjitter">Call Back Jitter:</label>
                    <input
                        type="text"
                        id="callbackjitter"
                        name="callbackjitter"
                        value={formData.callbackjitter}
                        onChange={handleChange}
                        placeholder="Enter callback jitter"
                    />
                </div>
                <div>
                    <label htmlFor="transportprotocol">Transport Protocol:</label>
                    <select
                        id="transportprotocol"
                        name="transportprotocol"
                        value={formData.transportprotocol}
                        onChange={handleChange}
                    >
                        <option value="TCP">TCP</option>
                        <option value="QUIC">QUIC</option>
                    </select>
                </div>
                <div>
                    <label htmlFor="logging">Enable Logging:</label>
                    <select
                        id="logging"
                        name="logging"
                        value={formData.logging}
                        onChange={handleChange}
                    >
                        <option value="true">True</option>
                        <option value="false">False</option>
                    </select>
                </div>
                <div>
                    <label htmlFor="compression">Compression:</label>
                    <select
                        id="compression"
                        name="compression"
                        value={formData.compression}
                        onChange={handleChange}
                    >
                        <option value="none">None</option>
                        <option value="upx">UPX</option>
                    </select>
                </div>
                <button type="submit" disabled={loading}>
                    {loading ? 'Creating...' : 'Create Payload'}
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

export default NewPayloadForm;
