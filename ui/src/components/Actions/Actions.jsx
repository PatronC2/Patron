import React, { useEffect, useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import CreateNewActionForm from './CreateNewActionForm';
import './Actions.css';

const Actions = () => {
    const { auth } = useContext(AuthContext);
    const [actions, setActions] = useState([]);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('current_actions');
    const navigate = useNavigate();

    const fetchData = async () => {
        try {
            const response = await axios.get('/api/actions');
            setActions(response.data.data || []);
        } catch (err) {
            setError(err.message);
        }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 10000);
        return () => clearInterval(interval);
    }, []);

    const handleDelete = async (actionID) => {
        try {
            await axios.delete(`/api/actions/${actionID}`);
            fetchData();
        } catch (err) {
            setError(err.message);
        }
    };

    const handleTabChange = (tab) => {
        setActiveTab(tab);
    };

    if (error) {
        return <div className="error">{error}</div>;
    }

    return (
        <div className="actions-container">
            <div className="header">
                <h1>Actions</h1>
                <div className="header-buttons">
                    <button
                        className={activeTab === 'current_actions' ? 'active' : ''}
                        onClick={() => handleTabChange('current_actions')}
                    >
                        Existing Actions
                    </button>
                    <button
                        className={activeTab === 'new' ? 'active' : ''}
                        onClick={() => handleTabChange('new')}
                    >
                        Create New Action
                    </button>
                </div>
            </div>
            {activeTab === 'current_actions' ? (
                <div className="actions-content">
                    <table>
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Description</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {actions.map((action) => (
                                <tr key={action.ActionID}>
                                    <td>{action.Name}</td>
                                    <td>{action.Description}</td>
                                    <td>
                                        <button onClick={() => handleDelete(action.ActionID)}>Delete</button>
                                        <button onClick={() => navigate(`/actions/edit?actionID=${action.ActionID}`)}>Edit</button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            ) : (
                <CreateNewActionForm fetchData={fetchData} setActiveTab={setActiveTab} />
            )}
        </div>
    );
};

export default Actions;
