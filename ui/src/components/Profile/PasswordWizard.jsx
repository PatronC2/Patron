import React, { useContext, useState } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import './ProfileWizard.css';

const PasswordWizard = ({ onClose }) => {
    const axios = useAxios();
    const { auth } = useContext(AuthContext);
    const [loading, setLoading] = useState(false);
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [formData, setFormData] = useState({
        oldPassword: '',
        newPassword: '',
        confirmNewPassword: '',
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
        const { oldPassword, newPassword, confirmNewPassword } = formData;
        if (newPassword !== confirmNewPassword) {
            setNotice('New passwords do not match', 'error');
            return;
        }
        setLoading(true);
        try {
            await axios.put('/api/profile/password', {
                oldPassword,
                newPassword,
            }, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `${auth.accessToken}`,
                },
            });
            setNotice('Password updated successfully!', 'success');
        } catch (error) {
            const msg = error.response?.data?.error || error.message || 'Failed to update password';
            setNotice(`Error: ${msg}`, 'error');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="wizard-overlay" onClick={onClose}>
            <div className="wizard-card" onClick={(e) => e.stopPropagation()}>
                <div className="wizard-header">
                    <div>
                        <h2>Change Password</h2>
                        <div className="wizard-steps">
                            <span className="active">Password</span>
                        </div>
                    </div>
                    <button type="button" className="wizard-close" onClick={onClose}>Close</button>
                </div>
                <div className="wizard-body">
                    <div className="wizard-form">
                        <label>Old Password</label>
                        <input
                            type="password"
                            name="oldPassword"
                            value={formData.oldPassword}
                            onChange={handleChange}
                            required
                        />
                        <label>New Password</label>
                        <input
                            type="password"
                            name="newPassword"
                            value={formData.newPassword}
                            onChange={handleChange}
                            required
                        />
                        <label>Confirm New Password</label>
                        <input
                            type="password"
                            name="confirmNewPassword"
                            value={formData.confirmNewPassword}
                            onChange={handleChange}
                            required
                        />
                        <button type="button" onClick={handleSubmit} disabled={loading}>
                            {loading ? 'Saving...' : 'Change Password'}
                        </button>
                    </div>
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

export default PasswordWizard;
