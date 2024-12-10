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
            setError("Event ID not specified");
            return;
        }

        try {
            const response = await axios.get(`/api/events/${eventID}`, {
                headers: {
                    'Authorization': `${auth.accessToken}`,
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
                    'Authorization': `${auth.accessToken}`,
                },
            });
            navigate('/events');
        } catch (err) {
            setError(err.message);
        }
    };

    if (error) {
        return <div>Error: {error}</div>;
    }

    if (!event) {
        return <div>Loading...</div>;
    }

    return (
        <div className="edit-event-container">
            <h1>Edit Event</h1>
            <form>
                <div>
                    <label>Name:</label>
                    <input
                        type="text"
                        value={event.Name}
                        onChange={(e) => setEvent({ ...event, Name: e.target.value })}
                    />
                </div>
                <div>
                    <label>Description:</label>
                    <textarea
                        value={event.Description}
                        onChange={(e) => setEvent({ ...event, Description: e.target.value })}
                    ></textarea>
                </div>
                <div>
                    <label>Schedule:</label>
                    <input
                        type="text"
                        value={event.Schedule}
                        onChange={(e) => setEvent({ ...event, Schedule: e.target.value })}
                    />
                </div>
                <button type="button" onClick={handleSave}>
                    Save
                </button>
                <button type="button" onClick={() => navigate('/events')}>
                    Cancel
                </button>
            </form>
        </div>
    );
};

export default EditEvent;
