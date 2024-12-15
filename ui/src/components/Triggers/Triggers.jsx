import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';

const Triggers = () => {
    const { auth } = useContext(AuthContext);
    const [triggers, setTriggers] = useState([]);
    const [error, setError] = useState(null);

    useEffect(() => {
        fetchTriggers();
    }, [auth.accessToken]);

    const fetchTriggers = async () => {
        try {
            const response = await axios.get('/api/triggers', {
                headers: {
                    'Authorization': `${auth.accessToken}`
                }
            });

            setTriggers(response.data.data || []);
        } catch (err) {
            setError(err.message);
        }
    };

    const handleDelete = async (triggerID) => {
        try {
            await axios.delete(`/api/triggers/${triggerID}`, {
                headers: {
                    'Authorization': `${auth.accessToken}`
                }
            });
            fetchTriggers();
        } catch (err) {
            setError(err.message);
        }
    };

    if (error) {
        return <div>Error: {error}</div>;
    }

    return (
        <div className="triggers-container">
            <h1>Triggers</h1>
            <table>
                <thead>
                    <tr>
                        <th>Event ID</th>
                        <th>Action ID</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {triggers.map(trigger => (
                        <tr key={trigger.id}>
                            <td>{trigger.event_id}</td>
                            <td>{trigger.action_id}</td>
                            <td>
                                <button onClick={() => handleDelete(trigger.id)}>Delete</button>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
};

export default Triggers;
