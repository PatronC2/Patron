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
      const response = await axios.get(`/api/agent?=${queryParam}`);
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
    <div className="home-container">
      <h1>Agent</h1>
      <div>
        {data && data.length > 0 ? (
          data.map((item, index) => (
            <div key={index}>
              {/* Render your data here */}
              {JSON.stringify(item)}
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
