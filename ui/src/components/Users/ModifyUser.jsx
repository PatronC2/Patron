import React, { useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import './PasswordChange.css';

const ChangePasswordForm = ({ username, setActiveTab }) => {
    const { auth } = useContext(AuthContext);
    const [formData, setFormData] = useState({
        newPassword: '',
        confirmNewPassword: '',
    });
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData({ ...formData, [name]: value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        const { newPassword, confirmNewPassword } = formData;

        if (newPassword !== confirmNewPassword) {
            setNotification('New passwords do not match');
            setNotificationType('error');
            setTimeout(() => setNotification(''), 3000);
            return;
        }

        try {
            const response = await axios.put(`/api/admin/users/${username}`, {
                newPassword,
            }, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `${auth.accessToken}`,
                },
            });

            setNotification('Password updated successfully!');
            setNotificationType('success');
            setTimeout(() => {
                setNotification('');
                setNotificationType('');
            }, 3000);
        } catch (error) {
            if (error.response) {
                setNotification(`Error: ${error.response.data.error}`);
                setNotificationType('error');
            } else if (error.request) {
                setNotification('Error: No response received from server.');
                setNotificationType('error');
            } else {
                setNotification(`Error: ${error.message}`);
                setNotificationType('error');
            }
            setTimeout(() => {
                setNotification('');
                setNotificationType('');
         }, 3000);
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <div>
                <label htmlFor="newPassword">New Password:</label>
                <input
                    type="password"
                    id="newPassword"
                    name="newPassword"
                    value={formData.newPassword}
                    onChange={handleChange}
                    required
                />
            </div>
            <div>
                <label htmlFor="confirmNewPassword">Confirm New Password:</label>
                <input
                    type="password"
                    id="confirmNewPassword"
                    name="confirmNewPassword"
                    value={formData.confirmNewPassword}
                    onChange={handleChange}
                    required
                />
            </div>
            <button type="submit">Change Password</button>
            {notification && (
                <div className={`notification ${notificationType}`}>
                    {notification}
                </div>
            )}
        </form>
    );
};

export default ChangePasswordForm;
