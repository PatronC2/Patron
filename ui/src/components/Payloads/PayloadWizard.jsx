import React, { useEffect, useMemo, useState } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import './PayloadWizard.css';

const steps = [
    'OS',
    'Type',
    'Network',
    'Options',
    'Review'
];

const PayloadWizard = ({ fetchData, onClose }) => {
    const cfg = window.runtimeConfig;
    const PATRON_C2_IP = `${cfg.REACT_APP_NGINX_IP}`;
    const PATRON_C2_PORT = `${cfg.REACT_APP_C2SERVER_PORT}`;
    const axios = useAxios();

    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [loading, setLoading] = useState(false);

    const [payloadConfs, setPayloadConfs] = useState({});
    const [selectedOS, setSelectedOS] = useState('');
    const [selectedTypeKey, setSelectedTypeKey] = useState('');

    const [selectedListenerIndex, setSelectedListenerIndex] = useState('');
    const [redirectors, setRedirectors] = useState([]);
    const [stepIndex, setStepIndex] = useState(0);

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

    useEffect(() => {
        const fetchStuff = async () => {
            try {
                const typesResp = await axios.get('/api/payloadconfs');
                const confs = typesResp.data || {};
                setPayloadConfs(confs);
                const osList = Array.from(
                    new Set(Object.values(confs).map(c => c.os).filter(Boolean))
                );
                if (osList.length > 0) {
                    setSelectedOS(osList[0]);
                }

                const redirResp = await axios.get('/api/redirectors');
                const list = Array.isArray(redirResp.data.data) ? redirResp.data.data : [];
                setRedirectors(list);
            } catch (error) {
                setNotification('Error fetching payload types or redirectors.');
                setNotificationType('error');
            }
        };
        fetchStuff();
    }, [axios]);

    const availableOS = useMemo(
        () =>
            Array.from(new Set(Object.values(payloadConfs).map(c => c.os).filter(Boolean))),
        [payloadConfs]
    );

    const availableTypes = useMemo(() => {
        const entries = Object.entries(payloadConfs)
            .filter(([, conf]) => conf.os === selectedOS)
            .map(([key, conf]) => ({
                key,
                label: conf.type,
                description: conf.description || ''
            }));
        return entries;
    }, [payloadConfs, selectedOS]);

    useEffect(() => {
        if (availableTypes.length > 0) {
            setSelectedTypeKey(availableTypes[0].key);
            setFormData(prev => ({ ...prev, type: availableTypes[0].key }));
        } else {
            setSelectedTypeKey('');
            setFormData(prev => ({ ...prev, type: '' }));
        }
    }, [availableTypes]);

    const onlineListeners = useMemo(
        () =>
            (redirectors || []).filter(
                r => r.status === 'Online' && r.listenport && r.listenport !== ''
            ),
        [redirectors]
    );

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

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const canContinue = () => {
        if (stepIndex === 0) return !!selectedOS;
        if (stepIndex === 1) return !!selectedTypeKey;
        return true;
    };

    const nextStep = () => {
        if (!canContinue()) return;
        setStepIndex(prev => Math.min(prev + 1, steps.length - 1));
    };

    const prevStep = () => {
        setStepIndex(prev => Math.max(prev - 1, 0));
    };

    const handleSubmit = async () => {
        setLoading(true);
        try {
            const payload = {
                ...formData,
                type: selectedTypeKey,
                serverport: Number(formData.serverport),
                callbackfrequency: Number(formData.callbackfrequency),
                callbackjitter: Number(formData.callbackjitter),
                logging: formData.logging === 'true',
                compression: formData.compression === 'none' ? '' : formData.compression,
            };

            const response = await axios.post('/api/payload', payload, {
                headers: { 'Content-Type': 'application/json' },
            });

            if (response.status === 200) {
                setNotification('Payload created successfully!');
                setNotificationType('success');
                fetchData();
                setTimeout(() => onClose(), 1500);
            } else {
                throw new Error(`Unexpected status code: ${response.status}`);
            }
        } catch (error) {
            const msg =
                error?.response?.data?.error ||
                error?.message ||
                'Failed to create payload';
            setNotification(`Failed to create payload: ${msg}`);
            setNotificationType('error');
        } finally {
            setLoading(false);
        }
    };

    const renderStep = () => {
        switch (stepIndex) {
            case 0:
                return (
                    <div className="wizard-step">
                        <h3>Select Operating System</h3>
                        <div className="wizard-options">
                            {availableOS.map(os => (
                                <button
                                    key={os}
                                    type="button"
                                    className={`wizard-option ${selectedOS === os ? 'active' : ''}`}
                                    onClick={() => setSelectedOS(os)}
                                >
                                    {os ? os.charAt(0).toUpperCase() + os.slice(1) : os}
                                </button>
                            ))}
                        </div>
                    </div>
                );
            case 1:
                return (
                    <div className="wizard-step">
                        <h3>Select Payload Type</h3>
                        <div className="wizard-grid">
                            {availableTypes.map(t => (
                                <button
                                    key={t.key}
                                    type="button"
                                    className={`wizard-type-card ${selectedTypeKey === t.key ? 'active' : ''}`}
                                    onClick={() => {
                                        setSelectedTypeKey(t.key);
                                        setFormData(prev => ({ ...prev, type: t.key }));
                                    }}
                                >
                                    <div className="wizard-card-title">{t.label}</div>
                                    <div className="wizard-card-desc">{t.description}</div>
                                </button>
                            ))}
                        </div>
                    </div>
                );
            case 2:
                return (
                    <div className="wizard-step">
                        <h3>Network Settings</h3>
                        <div className="wizard-form">
                            <label>Use Online Redirector Listener</label>
                            <select
                                id="listenerTarget"
                                onChange={handleListenerTargetChange}
                                value={selectedListenerIndex}
                            >
                                <option value="">-- Optional: select an online listener --</option>
                                {onlineListeners.map((r, idx) => (
                                    <option
                                        key={`${r.id}-${r.listenport}-${r.transportprotocol}`}
                                        value={idx}
                                    >
                                        {r.name} — {r.listenip}:{r.listenport}
                                        {r.transportprotocol
                                            ? ` (${r.transportprotocol.toUpperCase()})`
                                            : ''}
                                    </option>
                                ))}
                            </select>
                            <label>Listener IP</label>
                            <input
                                type="text"
                                name="serverip"
                                value={formData.serverip}
                                onChange={handleChange}
                            />
                            <label>Listener Port</label>
                            <input
                                type="text"
                                name="serverport"
                                value={formData.serverport}
                                onChange={handleChange}
                            />
                            <label>Callback Frequency</label>
                            <input
                                type="text"
                                name="callbackfrequency"
                                value={formData.callbackfrequency}
                                onChange={handleChange}
                            />
                            <label>Callback Jitter</label>
                            <input
                                type="text"
                                name="callbackjitter"
                                value={formData.callbackjitter}
                                onChange={handleChange}
                            />
                            <label>Transport Protocol</label>
                            <select
                                name="transportprotocol"
                                value={formData.transportprotocol}
                                onChange={handleChange}
                            >
                                <option value="TCP">TCP</option>
                                <option value="QUIC">QUIC</option>
                            </select>
                        </div>
                    </div>
                );
            case 3:
                return (
                    <div className="wizard-step">
                        <h3>Additional Options</h3>
                        <div className="wizard-form">
                            <label>Payload Name</label>
                            <input
                                type="text"
                                name="name"
                                value={formData.name}
                                onChange={handleChange}
                                placeholder="Enter the payload name"
                            />
                            <label>Description</label>
                            <textarea
                                name="description"
                                value={formData.description}
                                onChange={handleChange}
                                placeholder="Enter a brief description"
                            />
                            <label>Enable Logging</label>
                            <select name="logging" value={formData.logging} onChange={handleChange}>
                                <option value="true">True</option>
                                <option value="false">False</option>
                            </select>
                            <label>Compression</label>
                            <select name="compression" value={formData.compression} onChange={handleChange}>
                                <option value="none">None</option>
                                <option value="upx">UPX</option>
                            </select>
                        </div>
                    </div>
                );
            case 4:
                return (
                    <div className="wizard-step">
                        <h3>Review & Create</h3>
                        <div className="wizard-review">
                            <div><strong>OS:</strong> {selectedOS || '—'}</div>
                            <div><strong>Type:</strong> {payloadConfs[selectedTypeKey]?.type || '—'}</div>
                            <div><strong>Description:</strong> {formData.description || '—'}</div>
                            <div><strong>Listener:</strong> {formData.serverip}:{formData.serverport}</div>
                            <div><strong>Callback:</strong> {formData.callbackfrequency}s / jitter {formData.callbackjitter}%</div>
                            <div><strong>Protocol:</strong> {formData.transportprotocol}</div>
                            <div><strong>Logging:</strong> {formData.logging}</div>
                            <div><strong>Compression:</strong> {formData.compression}</div>
                        </div>
                    </div>
                );
            default:
                return null;
        }
    };

    return (
        <div className="wizard-overlay" onClick={onClose}>
            <div className="wizard-card" onClick={(e) => e.stopPropagation()}>
                <div className="wizard-header">
                    <div>
                        <h2>Create Payload</h2>
                        <div className="wizard-steps">
                            {steps.map((s, idx) => (
                                <span key={s} className={idx === stepIndex ? 'active' : ''}>{s}</span>
                            ))}
                        </div>
                    </div>
                    <button type="button" className="wizard-close" onClick={onClose}>Close</button>
                </div>
                <div className="wizard-body">
                    {renderStep()}
                </div>
                <div className="wizard-footer">
                    <button type="button" onClick={prevStep} disabled={stepIndex === 0}>Back</button>
                    {stepIndex < steps.length - 1 ? (
                        <button type="button" onClick={nextStep} disabled={!canContinue()}>Next</button>
                    ) : (
                        <button type="button" onClick={handleSubmit} disabled={loading}>
                            {loading ? 'Creating...' : 'Create Payload'}
                        </button>
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

export default PayloadWizard;
