import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import { useLocation } from 'react-router-dom';
import './Agent.css';

const Agent = () => {
  const { auth } = useContext(AuthContext);
  const [data, setData] = useState([]);
  const [error, setError] = useState(null);

  const location = useLocation();

  const getQueryParam = (param) => {
    const searchParams = new URLSearchParams(location.search);
    return searchParams.get(param);
  };

  const fetchData = async () => {
    try {
      const queryParam = getQueryParam('agt');
      const response = await axios.get(`/api/oneagent/${queryParam}`);
      const responseData = response.data.data;

      if (Array.isArray(responseData)) {
        setData(responseData);
      } else {
        setError(responseData);
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
    return <div>Error: {error}, Data: {data}</div>;
  }

  return (
    <div className="agent-container">
      <h1>Agent Details</h1>
      <div>
        {data && data.length > 0 ? (
          data.map((item, index) => (
            <div key={index} className="agent-details">
              <h3>Agent Info</h3>
              <ul>
                <li><strong>UUID:</strong> {item.uuid}</li>
                <li><strong>Callback to:</strong> {item.callbackto}</li>
                <li><strong>Callback Frequency:</strong> {item.callbackfrequency} seconds</li>
                <li><strong>Callback Jitter:</strong> {item.callbackjitter}%</li>
                <li><strong>Agent IP:</strong> {item.agentip || 'N/A'}</li>
                <li><strong>Username:</strong> {item.username || 'N/A'}</li>
                <li><strong>Hostname:</strong> {item.hostname || 'N/A'}</li>
                <li><strong>Status:</strong> {item.status || 'Unknown'}</li>
              </ul>
            </div>
          ))
        ) : (
          <p>No data available</p>
        )}
      </div>
    </div>
  );
};

export default Agent;
