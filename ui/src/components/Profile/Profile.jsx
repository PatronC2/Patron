import React, { useEffect, useState, useContext } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import './Profile.css';
import PasswordWizard from './PasswordWizard';
import ApiKeyWizard from './ApiKeyWizard';

const Profile = () => {
    const axios = useAxios();
    const { auth } = useContext(AuthContext);
    const [user, setUser] = useState(null);
    const [error, setError] = useState(null);
    const [showPasswordWizard, setShowPasswordWizard] = useState(false);
    const [showApiKeyWizard, setShowApiKeyWizard] = useState(false);
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');

    useEffect(() => {
        document.body.classList.add('profile-page');
        fetchData();
        const interval = setInterval(() => {
            fetchData();
        }, 10000);

        return () => {
            document.body.classList.remove('profile-page');
            clearInterval(interval);
        };
    }, [auth.accessToken]);

    const fetchData = async () => {
        try {
            const response = await axios.get('/api/profile/user', {
                headers: {
                    'Authorization': `${auth.accessToken}`
                }
            });

            const responseData = response.data.data;
            setUser(responseData);
        } catch (error) {
            if (error.response) {
                if (error.response.status === 401) {
                    setNotification('Error: Unauthorized.');
                    setNotificationType('error');
                } else {
                    console.error(`Failed: ${error.response.data}`);
                    setNotification(`Failed: ${error.response.data}`);
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

    if (error) {
        return <div>Error: {error}</div>;
    }

    return (
        <div className="profile-container">
            <div className="header">
                <h1>User Profile</h1>
                <div className="header-buttons">
                    <button className="active" onClick={() => setShowPasswordWizard(true)}>
                        Change Password
                    </button>
                    <button className="active" onClick={() => setShowApiKeyWizard(true)}>
                        API Key
                    </button>
                </div>
            </div>
            {user ? (
                <table>
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Username</th>
                            <th>Role</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td>{user.ID}</td>
                            <td>{user.Username}</td>
                            <td>{user.Role}</td>
                        </tr>
                    </tbody>
                </table>
            ) : null}

            {showPasswordWizard && (
                <PasswordWizard onClose={() => setShowPasswordWizard(false)} />
            )}
            {showApiKeyWizard && (
                <ApiKeyWizard
                    username={user?.Username || ''}
                    onClose={() => setShowApiKeyWizard(false)}
                />
            )}
        </div>
    );
};

export default Profile;
