import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';

const LogLevelSettings = ({ appName }) => {
	const { auth } = useContext(AuthContext);
	const [logLevel, setLogLevel] = useState('');
	const [notification, setNotification] = useState('');
	const [notificationType, setNotificationType] = useState('');

	useEffect(() => {
		const fetchLogLevel = async () => {
			try {
				const response = await axios.get('/api/admin/logging', {
                    params: { app: appName }
                });                
				setLogLevel(response.data.log_level);
			} catch (err) {
				setNotification('Failed to load log level');
				setNotificationType('error');
			}
		};
		fetchLogLevel();
	}, [appName, auth.accessToken]);

	const handleUpdate = async () => {
		try {
			await axios.put('/api/admin/logging', null, {
                params: {
                    app: appName,
                    log_level: logLevel
                }
            });            
			setNotification('Log level updated successfully');
			setNotificationType('success');
		} catch (err) {
			setNotification('Failed to update log level');
			setNotificationType('error');
		}
	};

	return (
		<div className="log-settings">
			<h2>{appName.charAt(0).toUpperCase() + appName.slice(1)} Log Level</h2>
			<select value={logLevel} onChange={(e) => setLogLevel(e.target.value)}>
				<option value="debug">Debug</option>
				<option value="info">Info</option>
				<option value="warning">Warning</option>
				<option value="error">Error</option>
			</select>
			<button onClick={handleUpdate}>Update</button>
			{notification && (
				<div className={`notification ${notificationType}`}>{notification}</div>
			)}
		</div>
	);
};

export default LogLevelSettings;
