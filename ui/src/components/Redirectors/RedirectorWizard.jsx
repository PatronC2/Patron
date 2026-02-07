import React, { useContext, useMemo, useState } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import './RedirectorWizard.css';

const steps = [
    'Basics',
    'Listen',
    'Forward',
    'Review'
];

const RedirectorWizard = ({ redirectors, fetchData, onClose }) => {
    const cfg = window.runtimeConfig;
    const PATRON_C2_IP = `${cfg.REACT_APP_NGINX_IP}`;
    const PATRON_C2_PORT = `${cfg.REACT_APP_C2SERVER_PORT}`;
    const axios = useAxios();
    const { auth } = useContext(AuthContext);

    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [loading, setLoading] = useState(false);
    const [stepIndex, setStepIndex] = useState(0);
    const [selectedForwardTargetId, setSelectedForwardTargetId] = useState('');

    const [formData, setFormData] = useState({
        Name: '',
        Description: '',
        ForwardIP: `${PATRON_C2_IP}`,
        ForwardPort: `${PATRON_C2_PORT}`,
        ListenIPv4: 'x.x.x.x',
        ListenIPv6: '',
        ListenPort: `${PATRON_C2_PORT}`,
    });

    const onlineTargets = useMemo(
        () =>
            (redirectors || []).filter(
                r =>
                    r.status === 'Online' &&
                    r.transportprotocol === 'tcp' &&
                    r.listenport &&
                    r.listenport !== ''
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
        setFormData(prev => ({
            ...prev,
            ForwardIP: target.listenip,
            ForwardPort: target.listenport,
        }));
    };

    const canContinue = () => {
        if (stepIndex === 0) return formData.Name.trim() !== '';
        if (stepIndex === 1) return formData.ListenPort.trim() !== '';
        if (stepIndex === 2) return formData.ForwardPort.trim() !== '';
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

                setNotification('Redirector created successfully! Install script downloading.');
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
                'Failed to create redirector';
            setNotification(`Failed to create redirector: ${msg}`);
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
                        <h3>Redirector Basics</h3>
                        <div className="wizard-form">
                            <label>Redirector Name</label>
                            <input
                                type="text"
                                name="Name"
                                value={formData.Name}
                                onChange={handleChange}
                                placeholder="Enter the name of the redirector"
                            />
                            <label>Description</label>
                            <textarea
                                name="Description"
                                value={formData.Description}
                                onChange={handleChange}
                                placeholder="Enter a brief description"
                            />
                        </div>
                    </div>
                );
            case 1:
                return (
                    <div className="wizard-step">
                        <h3>Listen Settings</h3>
                        <div className="wizard-form">
                            <label>Listen IPv4</label>
                            <input
                                type="text"
                                name="ListenIPv4"
                                value={formData.ListenIPv4}
                                onChange={handleChange}
                            />
                            <label>Listen IPv6 (optional)</label>
                            <input
                                type="text"
                                name="ListenIPv6"
                                value={formData.ListenIPv6}
                                onChange={handleChange}
                            />
                            <label>Listen Port</label>
                            <input
                                type="text"
                                name="ListenPort"
                                value={formData.ListenPort}
                                onChange={handleChange}
                            />
                        </div>
                    </div>
                );
            case 2:
                return (
                    <div className="wizard-step">
                        <h3>Forward Settings</h3>
                        <div className="wizard-form">
                            <label>Forward To (online redirector/listener)</label>
                            <select
                                id="ForwardTarget"
                                onChange={handleForwardTargetChange}
                                value={selectedForwardTargetId}
                            >
                                <option value="">-- Select an online redirector (optional) --</option>
                                {onlineTargets.map(r => (
                                    <option key={`${r.id}-${r.listenport}`} value={r.id}>
                                        {r.name} — {r.listenip}:{r.listenport}
                                        {r.transportprotocol
                                            ? ` (${r.transportprotocol.toUpperCase()})`
                                            : ''}
                                    </option>
                                ))}
                            </select>
                            <label>Forward IP</label>
                            <input
                                type="text"
                                name="ForwardIP"
                                value={formData.ForwardIP}
                                onChange={handleChange}
                            />
                            <label>Forward Port</label>
                            <input
                                type="text"
                                name="ForwardPort"
                                value={formData.ForwardPort}
                                onChange={handleChange}
                            />
                        </div>
                    </div>
                );
            case 3:
                return (
                    <div className="wizard-step">
                        <h3>Review & Create</h3>
                        <div className="wizard-review">
                            <div><strong>Name:</strong> {formData.Name || '—'}</div>
                            <div><strong>Description:</strong> {formData.Description || '—'}</div>
                            <div><strong>Listen:</strong> {formData.ListenIPv4}:{formData.ListenPort}</div>
                            <div><strong>Listen IPv6:</strong> {formData.ListenIPv6 || '—'}</div>
                            <div><strong>Forward:</strong> {formData.ForwardIP}:{formData.ForwardPort}</div>
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
                        <h2>Create Redirector</h2>
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
                            {loading ? 'Creating...' : 'Create Redirector'}
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

export default RedirectorWizard;
