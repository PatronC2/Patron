import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import './Profile.css';
import PasswordChangeForm from './PasswordChange';

const Profile = () => {
    const { auth } = useContext(AuthContext);
    const [user, setUser] = useState(null);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('user_profile');
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');

    useEffect(() => {
        document.body.classList.add('profile-page');
        fetchData();
        const interval = setInterval(() => {
            fetchData();
        }, 5000);

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

    const handleTabChange = (tab) => {
        setActiveTab(tab);
    };

    if (error) {
        return <div>Error: {error}</div>;
    }

    return (
        <div className="main-content">
            <div className="header">
                <h1>User Profile</h1>
                <button
                    className={activeTab === 'user_profile' ? 'active' : ''}
                    onClick={() => handleTabChange('user_profile')}
                >
                    Existing User
                </button>
                <button
                    className={activeTab === 'new' ? 'active' : ''}
                    onClick={() => handleTabChange('new')}
                >
                    Password Change
                </button>
            </div>
            {activeTab === 'user_profile' ? (
                user ? (
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
                ) : (
                    <p>Loading...</p>
                )
            ) : (
                <div>
                    <PasswordChangeForm fetchData={fetchData} setActiveTab={setActiveTab} />
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

export default Profile;
