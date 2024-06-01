import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import NewUserForm from './NewUser';
import ChangePasswordForm from './ChangePasswordForm';  // Import ChangePasswordForm
import './Users.css';

const Users = () => {
    const { auth } = useContext(AuthContext);
    const [data, setData] = useState([]);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('current_users');
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [selectedUser, setSelectedUser] = useState(null);

    useEffect(() => {
        document.body.classList.add('users-page');
        fetchData();
        const interval = setInterval(() => {
            fetchData();
        }, 5000);

        return () => {
            document.body.classList.remove('users-page');
            clearInterval(interval);
        };
    }, [auth.accessToken]);

    const fetchData = async () => {
        try {
            const response = await axios.get('/api/admin/users', {
                headers: {
                    'Authorization': `${auth.accessToken}`
                }
            });

            const responseData = response.data.data;
            if (Array.isArray(responseData)) {
                setData(responseData);
            } else {
                setError('Data format is not as expected');
            }
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
        setSelectedUser(null);  // Clear selected user when tab changes
    };

    const handleUserClick = (user) => {
        setSelectedUser(user);
        setActiveTab('edit_user');
    };

    const handleDeleteUser = async (userId) => {
        try {
            await axios.delete(`/api/admin/users/${userId}`, {
                headers: {
                    'Authorization': `${auth.accessToken}`
                }
            });
            setNotification('User deleted successfully');
            setNotificationType('success');
            fetchData();
        } catch (error) {
            setNotification(`Error deleting user: ${error.message}`);
            setNotificationType('error');
        }
    };

    const handleUpdateUser = (user) => {
        setSelectedUser(user);
        setActiveTab('edit_user');
    };

    if (error) {
        return <div>Error: {error}</div>;
    }

    return (
        <div className="main-content">
            <div className="header">
                <h1>Users</h1>
                <button
                    className={activeTab === 'current_users' ? 'active' : ''}
                    onClick={() => handleTabChange('current_users')}
                >
                    Existing Users
                </button>
                <button
                    className={activeTab === 'new' ? 'active' : ''}
                    onClick={() => handleTabChange('new')}
                >
                    Create New User
                </button>
            </div>
            {activeTab === 'current_users' ? (
                data.length > 0 ? (
                    <table>
                        <thead>
                            <tr>
                                <th>ID</th>
                                <th>Name</th>
                                <th>Role</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {data.map(user => (
                                <tr key={user.ID}>
                                    <td>{user.ID}</td>
                                    <td>{user.Username}</td>
                                    <td>{user.Role}</td>
                                    <td>
                                        <button onClick={() => handleUserClick(user)}>Edit</button>
                                        <button onClick={() => handleDeleteUser(user.ID)}>Delete</button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                ) : (
                    <p>No users available</p>
                )
            ) : activeTab === 'new' ? (
                <NewUserForm fetchData={fetchData} setActiveTab={setActiveTab} />
            ) : activeTab === 'edit_user' && selectedUser ? (
                <div>
                    <h2>Edit User: {selectedUser.Username}</h2>
                    <ChangePasswordForm setActiveTab={setActiveTab} />
                    {/* Add form to change user role here */}
                </div>
            ) : (
                <div>
                    {/* Other content */}
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

export default Users;
