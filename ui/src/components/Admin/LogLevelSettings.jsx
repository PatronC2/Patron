import React, { useEffect, useState, useContext } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import './LogLevelSettings.css';

const LogLevelSettings = ({ appName }) => {
	const axios = useAxios();
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
			<div className="log-settings-panel">
				<div className="log-card">
					<div className="log-card-header">
						<h3>Log Level</h3>
						<p>Control how verbose the server logs are.</p>
					</div>
					<div className="log-card-body">
						<div className="log-control">
							<label className="log-label" htmlFor={`log-level-${appName}`}>Level</label>
							<select id={`log-level-${appName}`} value={logLevel} onChange={(e) => setLogLevel(e.target.value)}>
								<option value="debug">Debug</option>
								<option value="info">Info</option>
								<option value="warning">Warning</option>
								<option value="error">Error</option>
							</select>
						</div>
						<button className="log-action" onClick={handleLevelUpdate}>Update Log Level</button>
					</div>
				</div>

				<div className="log-card">
					<div className="log-card-header">
						<h3>Max Log File Size</h3>
						<p>Set the maximum size before log rotation.</p>
					</div>
					<div className="log-card-body">
						<div className="log-control">
							<label className="log-label" htmlFor={`log-size-${appName}`}>Size</label>
							<div className="log-inputs">
								<input
									id={`log-size-${appName}`}
									type="number"
									value={logSize}
									onChange={(e) => setLogSize(e.target.value)}
									placeholder="Enter size"
								/>
								<select value={unit} onChange={(e) => setUnit(e.target.value)}>
									<option value="MB">MB</option>
									<option value="GB">GB</option>
								</select>
							</div>
						</div>
						<button className="log-action" onClick={handleSizeUpdate}>Update Log Size</button>
					</div>
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
