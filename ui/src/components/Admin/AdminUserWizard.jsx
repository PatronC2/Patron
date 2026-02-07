import React, { useContext, useState } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import './AdminWizard.css';

const AdminUserWizard = ({ username, onClose }) => {
    const axios = useAxios();
    const { auth } = useContext(AuthContext);
    const [loading, setLoading] = useState(false);
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [formData, setFormData] = useState({
        newPassword: '',
        confirmNewPassword: '',
        newRole: ''
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
        const { newPassword, confirmNewPassword, newRole } = formData;
        if (newPassword && newPassword !== confirmNewPassword) {
            setNotice('New passwords do not match', 'error');
            return;
        }

        const requestBody = {
            ...(newPassword && { newPassword }),
            ...(newRole && { newRole })
        };

        setLoading(true);
        try {
            await axios.put(`/api/admin/users/${username}`, requestBody, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `${auth.accessToken}`,
                },
            });
            setNotice('Changes saved successfully!', 'success');
        } catch (error) {
            const msg = error.response?.data?.error || error.message || 'Failed to update user';
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
                        <h2>Edit User</h2>
                        <div className="wizard-steps">
                            <span className="active">{username}</span>
                        </div>
                    </div>
                    <button type="button" className="wizard-close" onClick={onClose}>Close</button>
                </div>
                <div className="wizard-body">
                    <div className="wizard-form">
                        <label>New Password</label>
                        <input
                            type="password"
                            name="newPassword"
                            value={formData.newPassword}
                            onChange={handleChange}
                        />
                        <label>Confirm New Password</label>
                        <input
                            type="password"
                            name="confirmNewPassword"
                            value={formData.confirmNewPassword}
                            onChange={handleChange}
                        />
                        <label>User Role</label>
                        <select name="newRole" value={formData.newRole} onChange={handleChange}>
                            <option value="">No Change</option>
                            <option value="readOnly">Read-Only</option>
                            <option value="operator">Operator</option>
                            <option value="admin">Admin</option>
                        </select>
                        <button type="button" onClick={handleSubmit} disabled={loading}>
                            {loading ? 'Saving...' : 'Save Changes'}
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

export default AdminUserWizard;
