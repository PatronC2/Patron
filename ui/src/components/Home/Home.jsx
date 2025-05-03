import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import { useNavigate } from 'react-router-dom';
import AuthContext from '../../context/AuthProvider';
import './Home.css';

const Home = ({ isMenuOpen }) => {
    const { auth } = useContext(AuthContext);
    const [data, setData] = useState([]);
    const [metrics, setMetrics] = useState({ onlineCount: '0', offlineCount: '0' });
    const [error, setError] = useState(null);
    const [hostnameFilter, setHostnameFilter] = useState('');
    const [ipFilter, setIpFilter] = useState('');
    const [statusFilter, setStatusFilter] = useState('Online');

    const navigate = useNavigate();

    const fetchData = async () => {
        try {
            const response = await axios.get('/api/agents');
            const responseData = response.data.data;
            setData(Array.isArray(responseData) ? responseData : []);
        } catch (err) {
            setError(err.message);
        }
    };

    const fetchMetrics = async () => {
        try {
            const response = await axios.get('/api/agentsmetrics');
            setMetrics(response.data.data || { onlineCount: '0', offlineCount: '0' });
        } catch (err) {
            console.error('Failed to fetch agent metrics:', err.message);
        }
    };    

    useEffect(() => {
        fetchData();
        fetchMetrics();
        const interval = setInterval(() => {
            fetchData();
            fetchMetrics();
        }, 10000);
        return () => clearInterval(interval);
    }, []);    

    const filteredData = data.filter(
        (item) =>
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
      <div className="shared-container home-container">
          <header className="home-header">
              <h1>Agents</h1>
          </header>
  
          <div className="status-boxes">
              <div className="status-box online">
                  <p>Online</p>
                  <h2>{metrics.onlineCount}</h2>
              </div>
              <div className="status-box offline">
                  <p>Offline</p>
                  <h2>{metrics.offlineCount}</h2>
              </div>
          </div>
  
          <div className="filters">
              <input
                  type="text"
                  placeholder="Filter by Hostname"
                  value={hostnameFilter}
                  onChange={(e) => setHostnameFilter(e.target.value)}
              />
              <input
                  type="text"
                  placeholder="Filter by Agent IP"
                  value={ipFilter}
                  onChange={(e) => setIpFilter(e.target.value)}
              />
              <select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
              >
                  <option value="All">Status: All</option>
                  <option value="Online">Status: Online</option>
                  <option value="Offline">Status: Offline</option>
              </select>
          </div>
  
          {filteredData.length > 0 ? (
              <div className="table-container">
                  <table>
                      <thead>
                          <tr>
                              <th>UUID</th>
                              <th>User</th>
                              <th>Hostname</th>
                              <th>Agent IP</th>
                              <th>Status</th>
                          </tr>
                      </thead>
                      <tbody>
                          {filteredData.map((item) => (
                              <tr
                                  key={item.uuid}
                                  onClick={() => handleRowClick(item.uuid)}
                                  className="go-to-agent"
                              >
                                  <td>{item.uuid.substring(0, 6)}</td>
                                  <td>{item.username}</td>
                                  <td>{item.hostname}</td>
                                  <td>{item.agentip}</td>
                                  <td>{item.status}</td>
                              </tr>
                          ))}
                      </tbody>
                  </table>
              </div>
          ) : (
              <p className="no-agents-message">No Agents</p>
          )}
      </div>
  );  
};

export default Home;
