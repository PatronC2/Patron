import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import { useLocation } from 'react-router-dom';
import './Agent.css';

const Agent = () => {
  const { auth } = useContext(AuthContext);
  const [data, setData] = useState(null);
  const [commands, setCommands] = useState([]);
  const [activeTab, setActiveTab] = useState('commands');
  const [error, setError] = useState(null);

  const location = useLocation();

  const getQueryParam = (param) => {
    const searchParams = new URLSearchParams(location.search);
    return searchParams.get(param);
  };

  const fetchData = async () => {
    try {
      const queryParam = getQueryParam('agt');
      const agentResponse = await axios.get(`/api/agent/${queryParam}`);
      const commandsResponse = await axios.get(`/api/commands/${queryParam}`);
      const responseData = agentResponse.data.data;

      if (responseData) {
        setData(responseData);
      } else {
        setError('No data found');
      }
      
      if (commandsResponse.data.data) {
        setCommands(commandsResponse.data.data);
      } else {
        setCommands([]);
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
  }, [location.search]);

  if (error) {
    return <div>Error: {error}</div>;
  }

  if (!data) {
    return <p>No data available</p>;
  }

  const renderCommandsTab = () => (
    <div className="commands-list">
      <h3>Commands</h3>
      {commands.length === 0 ? (
        <p>No commands available.</p>
      ) : (
        <ul>
          {commands.map((cmd) => (
            <li key={cmd.commanduuid}>
              <strong>Command:</strong> {cmd.command} <br />
              <strong>Output:</strong> {cmd.output !== "" ? cmd.output : "(No output)"} <br /> {}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
  

  const renderKeylogsTab = () => (
    <div>
      <h3>Keylogs</h3>
      <p>TODO.</p>
    </div>
  );

  return (
    <div className="agent-container">
      {/* Agent Details */}
      <div className="agent-details">
        <h1>Agent Details</h1>
        <ul>
          <li><strong>UUID:</strong> {data.uuid}</li>
          <li><strong>Callback to:</strong> {data.callbackto}</li>
          <li><strong>Callback Frequency:</strong> {data.callbackfrequency} seconds</li>
          <li><strong>Callback Jitter:</strong> {data.callbackjitter}%</li>
          <li><strong>Agent IP:</strong> {data.agentip || 'N/A'}</li>
          <li><strong>Username:</strong> {data.username || 'N/A'}</li>
          <li><strong>Hostname:</strong> {data.hostname || 'N/A'}</li>
          <li><strong>Status:</strong> {data.status || 'Unknown'}</li>
        </ul>
      </div>

      {/* Commands & Tabs */}
      <div className="agent-tabs">
        <div className="tabs">
          <button
            className={activeTab === 'commands' ? 'active' : ''}
            onClick={() => setActiveTab('commands')}
          >
            Command History
          </button>
          <button
            className={activeTab === 'keys' ? 'active' : ''}
            onClick={() => setActiveTab('keys')}
          >
            Keylogs
          </button>
        </div>

        <div className="tab-content">
          {activeTab === 'commands' && renderCommandsTab()}
          {activeTab === 'keys' && renderKeylogsTab()}
        </div>
      </div>
    </div>
  );
};

export default Agent;
