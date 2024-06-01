import React, { useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import './NewUser.css'

const NewUserForm = ({ fetchData, setActiveTab }) => {
    const { auth } = useContext(AuthContext);
    const [formData, setFormData] = useState({
        username: '',
        role: 'readOnly',
        password: '',
    });
    const [error, setError] = useState('');
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData({ ...formData, [name]: value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        const { username, password, confirmPassword, role } = formData;
        if (password !== confirmPassword) {
            setError('Passwords do not match');
            setNotification('Passwords do not match')
            setNotificationType('error');
            setTimeout(() => {
                setNotification('');
                setError('');
            }, 3000);
            return;
        }
        const url = `/api/admin/users`;
        try {
            const response = await axios.post(url, formData, {

                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `${auth.accessToken}`,
                },
            });

            if (response.status !== 200) {
                throw new Error(`Failed to compile: ${response.data}`);
            }

            setNotification('User created successfully!');
            setNotificationType('success');
            fetchData();
            setTimeout(() => {
                setActiveTab('current_users');
                setNotification('');
            }, 3000);
        } catch (error) {
            if (error.response) {
                if (error.response.status === 401) {
                    setNotification('Error: Unauthorized.');
                    setNotificationType('error');
                } else {
                    console.error(`Failed to create user: ${error.response.data}`);
                    setNotification(`Failed to create user: ${error.response.data}`);
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
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <div>
                <label htmlFor="username">New username:</label>
                <input type="text" id="username" name="username" value={formData.username} onChange={handleChange} />
            </div>
            <div>
                <label htmlFor="role">User role:</label>
                <select id="role" name="role" value={formData.role} onChange={handleChange}>
                    <option value="readOnly">Read-Only</option>
                    <option value="operator">Operator</option>
                    <option value="admin">Admin</option>
                </select>
            </div>
            <div className="input-container">
                <div className="label-input-container">
                    <label htmlFor="password">Password:</label>
                    <input type="password" id="password" name="password" value={formData.password} onChange={handleChange} />
                </div>
                <div className="label-input-container">
                    <label htmlFor="confirmPassword">Confirm Password:</label>
                    <input type="password" id="confirmPassword" name="confirmPassword" value={formData.confirmPassword} onChange={handleChange} />
                </div>
            </div>
            <button type="submit">Create User</button>
            {notification && <div className="notification">{notification}</div>}
        </form>
    );
};

export default NewUserForm;
