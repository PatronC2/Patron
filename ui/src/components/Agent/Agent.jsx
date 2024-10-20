import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import { useLocation } from 'react-router-dom';
import './Agent.css';

const Agent = () => {
  const { auth } = useContext(AuthContext);
  const [data, setData] = useState(null);
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

      if (responseData) {
        setData(responseData);
      } else {
        setError('No data found');
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

  return (
    <div className="agent-container">
      <h1>Agent Details</h1>
      <div className="agent-details">
        <h3>Agent Info</h3>
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
    </div>
  );
};

export default Agent;
