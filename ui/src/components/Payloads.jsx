import React, { useEffect, useState, useContext } from 'react';
import axios from '../api/axios';
import AuthContext from '../context/AuthProvider';

const Payloads = () => {
  const { auth } = useContext(AuthContext);
  const [data, setData] = useState([]);
  const [error, setError] = useState(null);

  const fetchData = async () => {
    try {
      const response = await axios.get('/api/payloads', {
        headers: {
          'Authorization': `${auth.accessToken}`
        }
      });

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
  }, [auth.accessToken]);

  if (error) {
    return <div>Error: {error}</div>;
  }
  return (
    <div>
      <h1>Payloads</h1>
      {data.length > 0 ? ( 
        <table>
          <thead>
            <tr>
              <th>UUID</th>
              <th>Name</th>
              <th>Description</th>
              <th>Listener IP</th>
              <th>Listener Port</th>
              <th>Callback Frequency</th>
              <th>Callback Jitter</th>
            </tr>
          </thead>
          <tbody>
            {data.map(item => (
              <tr key={item.uuid}>
                <td>{item.uuid.substring(0, 6)}</td>
                <td>{item.concat}</td>
                <td>{item.description}</td>
                <td>{item.serverip}</td>
                <td>{item.serverport}</td>
                <td>{item.callbackfrequency}</td>
                <td>{item.callbackjitter}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p>No Payloads</p>
      )}
    </div>
  );
};

export default Payloads;
