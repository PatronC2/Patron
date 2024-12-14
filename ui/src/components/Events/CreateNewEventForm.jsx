import React, { useState } from 'react';
import axios from '../../api/axios';
import './CreateNewEventForm.css';

const CreateNewEventForm = ({ fetchData, setActiveTab }) => {
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        schedule: '',
    });
    const [file, setFile] = useState(null);
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData({ ...formData, [name]: value });
    };

    const handleFileChange = (e) => {
        setFile(e.target.files[0]);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();

        if (!file) {
            setNotification('Please upload a script file.');
            setNotificationType('error');
            setTimeout(() => setNotification(''), 3000);
            return;
        }

        const formDataObj = new FormData();
        formDataObj.append('name', formData.name);
        formDataObj.append('description', formData.description);
        formDataObj.append('schedule', formData.schedule);
        formDataObj.append('script', file);

        try {
            await axios.post('/api/events', formDataObj, {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
            });

            setNotification('Event created successfully!');
            setNotificationType('success');
            fetchData();
            setTimeout(() => {
                setActiveTab('current_events');
                setNotification('');
            }, 3000);
        } catch (err) {
            setNotification('Failed to create event.');
            setNotificationType('error');
            setTimeout(() => setNotification(''), 3000);
        }
    };

    return (
        <div className="create-new-event-form">
            <form onSubmit={handleSubmit} encType="multipart/form-data">
                <div>
                    <label htmlFor="name">Name:</label>
                    <input
                        type="text"
                        id="name"
                        name="name"
                        value={formData.name}
                        onChange={handleChange}
                        required
                    />
                </div>
                <div>
                    <label htmlFor="description">Description:</label>
                    <textarea
                        id="description"
                        name="description"
                        value={formData.description}
                        onChange={handleChange}
                        rows="4"
                        required
                    />
                </div>
                <div>
                    <label htmlFor="schedule">Schedule:</label>
                    <input
                        type="text"
                        id="schedule"
                        name="schedule"
                        value={formData.schedule}
                        onChange={handleChange}
                        required
                    />
                </div>
                <div>
                    <label htmlFor="script">Script File:</label>
                    <input
                        type="file"
                        id="script"
                        name="script"
                        onChange={handleFileChange}
                        required
                    />
                </div>
                <button type="submit">Create Event</button>
                {notification && (
                    <div className={`notification ${notificationType}`}>{notification}</div>
                )}
            </form>
        </div>
    );
};

export default CreateNewEventForm;
