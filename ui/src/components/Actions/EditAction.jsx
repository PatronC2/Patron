import React, { useEffect, useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import './EditAction.css';
import { useLocation } from 'react-router-dom';

const EditAction = () => {
    const { auth } = useContext(AuthContext);
    const location = useLocation();
    const navigate = useNavigate();
    const [action, setAction] = useState(null);
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [error, setError] = useState(null);

    const getQueryParam = (param) => {
        const searchParams = new URLSearchParams(location.search);
        return searchParams.get(param);
    };

    const actionID = getQueryParam('actionID');

    useEffect(() => {
        fetchActionDetails();
    }, []);

    const fetchActionDetails = async () => {
        if (!actionID) {
            setError('Action ID not specified');
            return;
        }

        try {
            const response = await axios.get(`/api/actions/${actionID}`);
            const actionData = response.data;
            setAction(actionData);
        } catch (err) {
            setError(err.message);
        }
    };

    const handleSave = async () => {
        try {
            const updatedAction = {
                ...action,
            };

            await axios.put(`/api/actions/${actionID}`, updatedAction);
            setNotification('Action updated successfully!');
            setNotificationType('success');
            setTimeout(() => navigate('/actions'), 3000);
        } catch (err) {
            setNotification('Failed to update action');
            setNotificationType('error');
            setTimeout(() => setNotification(''), 3000);
        }
    };

    if (error) {
        return <div className="error-message">Error: {error}</div>;
    }

    if (!action) {
        return <div>Loading...</div>;
    }

    return (
        <div className="edit-action-container">
            <h1>Edit Action</h1>
            <form>
                <div>
                    <label htmlFor="name">Name:</label>
                    <input
                        type="text"
                        id="name"
                        value={action.Name}
                        onChange={(e) => setAction({ ...action, Name: e.target.value })}
                    />
                </div>
                <div>
                    <label htmlFor="description">Description:</label>
                    <textarea
                        id="description"
                        value={action.Description}
                        onChange={(e) => setAction({ ...action, Description: e.target.value })}
                    ></textarea>
                </div>
                <div className="button-container">
                    <button type="button" onClick={handleSave}>
                        Save
                    </button>
                    <button type="button" onClick={() => navigate('/actions')}>
                        Cancel
                    </button>
                </div>
            </form>
            {notification && (
                <div className={`notification ${notificationType}`}>{notification}</div>
            )}
        </div>
    );
};

export default EditAction;
