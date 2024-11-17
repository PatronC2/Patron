import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import { useNavigate } from 'react-router-dom';
import AuthContext from '../../context/AuthProvider';
import './Home.css';

const Home = ({ isMenuOpen }) => {
  const { auth } = useContext(AuthContext);
  const [data, setData] = useState([]);
  const [error, setError] = useState(null);
  const [hostnameFilter, setHostnameFilter] = useState('');
  const [ipFilter, setIpFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('All');

  const navigate = useNavigate();

  const fetchData = async () => {
    try {
      const response = await axios.get('/api/agents');
      const responseData = response.data.data;
      if (Array.isArray(responseData)) {
        setData(responseData);
      } else {
        setData([]);
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

  const filteredData = data.filter(item =>
    (hostnameFilter === '' || item.hostname.includes(hostnameFilter)) &&
    (ipFilter === '' || item.agentip.includes(ipFilter)) &&
    (statusFilter === 'All' || item.status === statusFilter)
  );

  const handleRowClick = (uuid) => {
    navigate(`/agent?agt=${uuid}`);
  };

  if (error) {
    return <div className="error-message">Error: {error}</div>;
  }

  return (
    <div className="home-container">
      <header className="home-header">
        <h1>Agents</h1>
      </header>

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

      <div className="filters">
        <input
          type="text"
          placeholder="Filter by Hostname"
          value={hostnameFilter}
          onChange={e => setHostnameFilter(e.target.value)}
        />
        <input
          type="text"
          placeholder="Filter by Agent IP"
          value={ipFilter}
          onChange={e => setIpFilter(e.target.value)}
        />
        <select
          value={statusFilter}
          onChange={e => setStatusFilter(e.target.value)}
        >
          <option value="All">All</option>
          <option value="Online">Online</option>
          <option value="Offline">Offline</option>
        </select>
      </div>

      {filteredData.length > 0 ? (
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
            {filteredData.map(item => (
              <tr
                key={item.uuid}
                onClick={() => handleRowClick(item.uuid)}
                className="go-to-agent"
              >
                <td>{item.uuid.substring(0, 6)}</td>
                <td>{item.hostname}</td>
                <td>{item.agentip}</td>
                <td>{item.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p className="no-agents-message">No Agents</p>
      )}
    </div>
  );
};

export default Home;
