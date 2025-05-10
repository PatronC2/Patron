import React, { useEffect, useState, useContext } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import NewRedirectorForm from './NewRedirectorForm';
import './Redirectors.css';

const Redirectors = () => {
    const axios = useAxios();
    const { auth } = useContext(AuthContext);
    const [data, setData] = useState([]);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('current_redirectors');
    const [statusFilter, setStatusFilter] = useState('Online');

    useEffect(() => {
        document.body.classList.add('redirectors-page');
        fetchData();
        const interval = setInterval(() => {
            fetchData();
        }, 10000);

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
                setData([]);
            }
        } catch (err) {
            setError(err.message);
        }
    };

    const activeCount = data.filter(item => item.status === 'Online').length;
    const inactiveCount = data.filter(item => item.status === 'Offline').length;

    const filteredData = data.filter(item =>
        (statusFilter === 'All' || item.status === statusFilter)
    );

    const handleTabChange = (tab) => {
        setActiveTab(tab);
    };

    if (error) {
        return <div>Error: {error}</div>;
    }

    return (
        <div className="redirector-container">
            <div className="header">
                <h1>Redirectors</h1>
                <div className="header-buttons">
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
            </div>

            {activeTab === 'current_redirectors' ? (
                <div className="redirectors-container">
                    <h1>Redirectors</h1>
                    <div className="status-boxes">
                        <div className="status-box online">
                        <p>Online</p>
                        <h2>{activeCount}</h2>
                        </div>
                        <div className="status-box offline">
                        <p>Offline</p>
                        <h2>{inactiveCount}</h2>
                        </div>
                    </div>

                    <div className="filters-container">
                        <div className="filters">
                            <select
                                value={statusFilter}
                                onChange={(e) => setStatusFilter(e.target.value)}
                            >
                                <option value="All">All</option>
                                <option value="Online">Online</option>
                                <option value="Offline">Offline</option>
                            </select>
                        </div>
                    </div>

                    {filteredData.length > 0 ? (
                        <table>
                            <thead>
                                <tr>
                                    <th>Name</th>
                                    <th>Description</th>
                                    <th>Forward IP</th>
                                    <th>Forward Port</th>
                                    <th>Listener Port</th>
                                    <th>Status</th>
                                </tr>
                            </thead>
                            <tbody>
                                {filteredData.map(item => (
                                    <tr key={item.id}>
                                        <td>{item.name}</td>
                                        <td>{item.description}</td>
                                        <td>{item.forwardip}</td>
                                        <td>{item.forwardport}</td>
                                        <td>{item.listenport}</td>
                                        <td>{item.status}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    ) : (
                        <p>No Redirectors</p>
                    )}
                </div>
            ) : (
                <div>
                    <NewRedirectorForm fetchData={fetchData} setActiveTab={setActiveTab} />
                </div>
            )}
        </div>
    );
};

export default Redirectors;
