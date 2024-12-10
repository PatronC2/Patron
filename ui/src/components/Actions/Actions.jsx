import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';

const Actions = () => {
    const { auth } = useContext(AuthContext);
    const [actions, setActions] = useState([]);
    const [error, setError] = useState(null);

    useEffect(() => {
        fetchActions();
    }, [auth.accessToken]);

    const fetchActions = async () => {
        try {
            const response = await axios.get('/api/actions', {
                headers: {
                    'Authorization': `${auth.accessToken}`
                }
            });

            setActions(response.data.data || []);
        } catch (err) {
            setError(err.message);
        }
    };

    const handleDelete = async (actionID) => {
        try {
            await axios.delete(`/api/actions/${actionID}`, {
                headers: {
                    'Authorization': `${auth.accessToken}`
                }
            });
            fetchActions();
        } catch (err) {
            setError(err.message);
        }
    };

    if (error) {
        return <div>Error: {error}</div>;
    }

    return (
        <div className="actions-container">
            <h1>Actions</h1>
            <table>
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Description</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {actions.map(action => (
                        <tr key={action.action_id}>
                            <td>{action.name}</td>
                            <td>{action.description}</td>
                            <td>
                                <button onClick={() => handleDelete(action.action_id)}>Delete</button>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
};

export default Actions;
