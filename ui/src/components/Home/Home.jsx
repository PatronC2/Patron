import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import './Home.css';

const Home = () => {
  const { auth } = useContext(AuthContext);
  const [data, setData] = useState([]);
  const [error, setError] = useState(null);

  const fetchData = async () => {
    try {
      const response = await axios.get('/api/agents');
      const responseData = response.data.data;
      if (Array.isArray(responseData)) {
        setData(responseData);
      } else {
        setError('Data format is not as expected');
      }
    } catch (err) {
      setError(err.message);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(() => {
      fetchData();
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  const onlineCount = data.filter(item => item.status === 'Online').length;
  const offlineCount = data.filter(item => item.status === 'Offline').length;

  if (error) {
    return <div>Error: {error}</div>;
  }

  return (
    <div className="home-container">
      <h1>Agents</h1>
      <div className="status-boxes">
        <div className="status-box online">
          <p>Online</p>
          <h2>{onlineCount}</h2>
        </div>
        <div className="status-box offline">
          <p>Offline</p>
          <h2>{offlineCount}</h2>
        </div>
      </div>
      {data.length > 0 ? (
        <table>
          <thead>
            <tr>
              <th>UUID</th>
              <th>Hostname</th>
              <th>Agent IP</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {data.map(item => (
              <tr key={item.uuid}>
                <td>{item.uuid.substring(0, 6)}</td>
                <td>{item.hostname}</td>
                <td>{item.agentip}</td>
                <td>{item.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p>No Agents</p>
      )}
    </div>
  );
};

export default Home;
