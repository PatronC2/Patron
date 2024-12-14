import React, { useEffect, useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import CreateNewEventForm from './CreateNewEventForm';
import './Events.css';

const Events = () => {
    const { auth } = useContext(AuthContext);
    const [events, setEvents] = useState([]);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('current_events');
    const navigate = useNavigate();

    const fetchData = async () => {
        try {
            const response = await axios.get('/api/events');
            setEvents(response.data.data || []);
        } catch (err) {
            setError(err.message);
        }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 10000);
        return () => clearInterval(interval);
    }, []);

    const handleDelete = async (eventID) => {
        try {
            await axios.delete(`/api/events/${eventID}`, {
                headers: {
                    Authorization: `${auth.accessToken}`,
                },
            });
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
        <div className="events-container">
            <div className="header">
                <h1>Events</h1>
                <div className="header-buttons">
                    <button
                        className={activeTab === 'current_events' ? 'active' : ''}
                        onClick={() => handleTabChange('current_events')}
                    >
                        Existing Events
                    </button>
                    <button
                        className={activeTab === 'new' ? 'active' : ''}
                        onClick={() => handleTabChange('new')}
                    >
                        Create New Event
                    </button>
                </div>
            </div>
            {activeTab === 'current_events' ? (
                <div className="events-content">
                    <table>
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Description</th>
                                <th>Schedule</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {events.map((event) => (
                                <tr key={event.EventID}>
                                    <td>{event.Name}</td>
                                    <td>{event.Description}</td>
                                    <td>{event.Schedule}</td>
                                    <td>
                                        <button onClick={() => handleDelete(event.EventID)}>Delete</button>
                                        <button onClick={() => navigate(`/events/edit?eventID=${event.EventID}`)}>Edit</button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            ) : (
                <CreateNewEventForm fetchData={fetchData} setActiveTab={setActiveTab} />
            )}
        </div>
    );
};

export default Events;
