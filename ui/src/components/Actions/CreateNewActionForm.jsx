import React, { useState } from 'react';
import axios from '../../api/axios';
import './CreateNewActionForm.css';

const CreateNewActionForm = ({ fetchData, setActiveTab }) => {
    const [formData, setFormData] = useState({
        name: '',
        description: '',
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
            setNotification('Please upload a zip file.');
            setNotificationType('error');
            setTimeout(() => setNotification(''), 3000);
            return;
        }

        const formDataObj = new FormData();
        formDataObj.append('name', formData.name);
        formDataObj.append('description', formData.description);
        formDataObj.append('file', file);

        try {
            await axios.post('/api/actions', formDataObj, {
                headers: {
                    'Content-Type': 'multipart/form-data',
                },
            });

            setNotification('Action created successfully!');
            setNotificationType('success');
            fetchData();
            setTimeout(() => {
                setActiveTab('current_actions');
                setNotification('');
            }, 3000);
        } catch (err) {
            setNotification('Failed to create action.');
            setNotificationType('error');
            setTimeout(() => setNotification(''), 3000);
        }
    };

    return (
        <div className="create-new-action-form">
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
                    <label htmlFor="file">Action Zip File:</label>
                    <input
                        type="file"
                        id="file"
                        name="file"
                        onChange={handleFileChange}
                        required
                    />
                </div>
                <button type="submit">Create Action</button>
                {notification && (
                    <div className={`notification ${notificationType}`}>{notification}</div>
                )}
            </form>
        </div>
    );
};

export default CreateNewActionForm;
