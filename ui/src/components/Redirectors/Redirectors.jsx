import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import NewRedirectorForm from './NewRedirectorForm';
import './Redirectors.css';

const Redirectors = () => {
    const { auth } = useContext(AuthContext);
    const [data, setData] = useState([]);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('current_redirectors');

    useEffect(() => {
        document.body.classList.add('redirectors-page');
        fetchData();
        const interval = setInterval(() => {
            fetchData();
        }, 5000);

        return () => {
            document.body.classList.remove('redirectors-page');
            clearInterval(interval);
        };
    }, [auth.accessToken]);

    const fetchData = async () => {
        try {
            const response = await axios.get('/api/redirectors', {
                headers: {
                    'Authorization': `${auth.accessToken}`
                }
            });

            const responseData = response.data.data;
            if (Array.isArray(responseData)) {
                setData(responseData);
            } else {
                setData('');
            }
        } catch (err) {
            setError(err.message);
        }
    };

    const handleTabChange = (tab) => {
        setActiveTab(tab);
    };

    if (error) {
        return <div>Error: {error}</div>;
    }

    return (
        <div className="main-content">
            <div className="header">
                <h1>Redirectors</h1>
                <button
                    className={activeTab === 'current_redirectors' ? 'active' : ''}
                    onClick={() => handleTabChange('current_redirectors')}
                >
                    Existing Redirectors
                </button>
                <button
                    className={activeTab === 'new' ? 'active' : ''}
                    onClick={() => handleTabChange('new')}
                >
                    Create New Redirector
                </button>
            </div>
            {activeTab === 'current_redirectors' ? (
                data.length > 0 ? (
                    <table>
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Description</th>
                                <th>Forward IP</th>
                                <th>Forward Port</th>
                                <th>Listener IP</th>
                                <th>Listener Port</th>
                                <th>Status</th>
                            </tr>
                        </thead>
                        <tbody>
                          {data.map(item => (
                              <tr key={item.id}>
                                  <td>{item.name}</td>
                                  <td>{item.description}</td>
                                  <td>{item.forwardip}</td>
                                  <td>{item.forwardport}</td>
                                  <td>{item.listenip}</td>
                                  <td>{item.listenport}</td>
                                  <td>{item.status}</td>
                              </tr>
                          ))}
                      </tbody>
                    </table>
                ) : (
                    <p>No Redirectors</p>
                )
            ) : (
                <div>
                    <NewRedirectorForm fetchData={fetchData} setActiveTab={setActiveTab} />
                </div>
            )}
        </div>
    );
};

export default Redirectors;
