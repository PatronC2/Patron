import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';

const LogLevelSettings = ({ appName }) => {
	const { auth } = useContext(AuthContext);
	const [logLevel, setLogLevel] = useState('');
	const [logSize, setLogSize] = useState('');
	const [unit, setUnit] = useState('MB');
	const [notification, setNotification] = useState('');
	const [notificationType, setNotificationType] = useState('');

	useEffect(() => {
		const fetchLogSettings = async () => {
            try {
                const levelRes = await axios.get('/api/admin/logging', {
                    params: { app: appName }
                });
                setLogLevel(levelRes.data.log_level);
        
                const sizeRes = await axios.get('/api/admin/log-size', {
                    params: { app: appName }
                });
        
                // Choose unit based on size
                if (sizeRes.data.size_mb >= 1024) {
                    setLogSize((sizeRes.data.size_bytes / (1024 * 1024 * 1024)).toFixed(1)); // GB
                    setUnit("GB");
                } else {
                    setLogSize(sizeRes.data.size_mb);
                    setUnit("MB");
                }
            } catch (err) {
                setNotification('Failed to load logging settings');
                setNotificationType('error');
            }
        };        
		fetchLogSettings();
	}, [appName, auth.accessToken]);

    useEffect(() => {
        if (notification) {
            const timer = setTimeout(() => {
                setNotification('');
                setNotificationType('');
            }, 3000);
            return () => clearTimeout(timer);
        }
    }, [notification]);

	const handleLevelUpdate = async () => {
		try {
			await axios.put('/api/admin/logging', null, {
				params: { app: appName, log_level: logLevel }
			});
			setNotification('Log level updated successfully');
			setNotificationType('success');
		} catch (err) {
			setNotification('Failed to update log level');
			setNotificationType('error');
		}
	};

	const handleSizeUpdate = async () => {
		if (!logSize || isNaN(logSize)) {
			setNotification('Invalid size value');
			setNotificationType('error');
			return;
		}
		try {
			await axios.put('/api/admin/log-size', {
				app: appName,
				size: parseInt(logSize, 10),
				unit
			});
			setNotification('Log size updated successfully');
			setNotificationType('success');
		} catch (err) {
			setNotification('Failed to update log size');
			setNotificationType('error');
		}
	};

	return (
		<div className="log-settings">
			<div className="log-section">
				<h3>Log Level</h3>
				<select value={logLevel} onChange={(e) => setLogLevel(e.target.value)}>
					<option value="debug">Debug</option>
					<option value="info">Info</option>
					<option value="warning">Warning</option>
					<option value="error">Error</option>
				</select>
				<button onClick={handleLevelUpdate}>Update Log Level</button>
			</div>

			<div className="log-section">
                <h3>Max Log File Size</h3>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                    <input
                        type="number"
                        value={logSize}
                        onChange={(e) => setLogSize(e.target.value)}
                        placeholder="Enter size"
                        style={{ width: '100px' }}
                    />
                    <select value={unit} onChange={(e) => setUnit(e.target.value)}>
                        <option value="MB">MB</option>
                        <option value="GB">GB</option>
                    </select>
                    <button onClick={handleSizeUpdate}>Update Log Size</button>
                </div>
            </div>

			{notification && (
				<div className={`notification ${notificationType}`}>
					{notification}
				</div>
			)}
		</div>
	);
};

export default LogLevelSettings;
