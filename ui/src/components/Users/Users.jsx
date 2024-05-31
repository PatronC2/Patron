import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import NewUserForm from './NewUser';
import './Users.css'

const Users = () => {
    const { auth } = useContext(AuthContext);
    const [data, setData] = useState([]);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('current_users');

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
        } catch (err) {
            setError(err.message);
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
                            </tr>
                        </thead>
                        <tbody>
                            {data.map(user => (
                                <tr key={user.ID}>
                                    <td>{user.ID}</td>
                                    <td>{user.Username}</td>
                                    <td>{user.Role}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                ) : (
                    <p>No Users</p>
                )
            ) : (
                <div>
                    <NewUserForm fetchData={fetchData} setActiveTab={setActiveTab} />
                </div>
            )}
        </div>
    );
};

export default Users;
