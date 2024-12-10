import React, { useEffect, useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import './EditEvent.css';
import { useLocation } from 'react-router-dom';

const EditEvent = () => {
    const { auth } = useContext(AuthContext);
    const location = useLocation();
    const navigate = useNavigate();
    const [event, setEvent] = useState(null);
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');
    const [error, setError] = useState(null);

    const getQueryParam = (param) => {
        const searchParams = new URLSearchParams(location.search);
        return searchParams.get(param);
    };

    const eventID = getQueryParam('eventID');

    useEffect(() => {
        fetchEventDetails();
    }, []);

    const fetchEventDetails = async () => {
        if (!eventID) {
            setError('Event ID not specified');
            return;
        }

        try {
            const response = await axios.get(`/api/events/${eventID}`, {
                headers: {
                    Authorization: `${auth.accessToken}`,
                },
            });
            setEvent(response.data);
        } catch (err) {
            setError(err.message);
        }
    };

    const handleSave = async () => {
        try {
            await axios.put(`/api/events/${eventID}`, event, {
                headers: {
                    Authorization: `${auth.accessToken}`,
                },
            });
            setNotification('Event updated successfully!');
            setNotificationType('success');
            setTimeout(() => navigate('/events'), 3000);
        } catch (err) {
            setNotification('Failed to update event');
            setNotificationType('error');
            setTimeout(() => setNotification(''), 3000);
        }
    };

    if (error) {
        return <div className="error-message">Error: {error}</div>;
    }

    if (!event) {
        return <div>Loading...</div>;
    }

    return (
        <div className="edit-event-container">
            <h1>Edit Event</h1>
            <form>
                <div>
                    <label htmlFor="name">Name:</label>
                    <input
                        type="text"
                        id="name"
                        value={event.Name}
                        onChange={(e) => setEvent({ ...event, Name: e.target.value })}
                    />
                </div>
                <div>
                    <label htmlFor="description">Description:</label>
                    <textarea
                        id="description"
                        value={event.Description}
                        onChange={(e) => setEvent({ ...event, Description: e.target.value })}
                    ></textarea>
                </div>
                <div>
                    <label htmlFor="schedule">Schedule:</label>
                    <input
                        type="text"
                        id="schedule"
                        value={event.Schedule}
                        onChange={(e) => setEvent({ ...event, Schedule: e.target.value })}
                    />
                </div>
                <div>
                    <button type="button" onClick={handleSave}>
                        Save
                    </button>
                    <button type="button" onClick={() => navigate('/events')}>
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

export default EditEvent;
