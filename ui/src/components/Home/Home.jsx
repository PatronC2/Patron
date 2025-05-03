import React, { useEffect, useState, useContext } from 'react';
import axios from '../../api/axios';
import { useNavigate } from 'react-router-dom';
import AuthContext from '../../context/AuthProvider';
import './Home.css';

const Home = ({ isMenuOpen }) => {
    const { auth } = useContext(AuthContext);
    const navigate = useNavigate();

    const [agents, setAgents] = useState([]);
    const [metrics, setMetrics] = useState({ onlineCount: '0', offlineCount: '0' });
    const [error, setError] = useState(null);

    const [hostnameFilter, setHostnameFilter] = useState('');
    const [ipFilter, setIpFilter] = useState('');
    const [statusFilter, setStatusFilter] = useState('All');
    const [tagFilter, setTagFilter] = useState('');

    const [offset, setOffset] = useState(0);
    const [totalCount, setTotalCount] = useState(0);
    const [sortField, setSortField] = useState('hostname');
    const [sortDirection, setSortDirection] = useState('asc');
    const limit = 10;

    const fetchMetrics = async () => {
        try {
            const response = await axios.get('/api/agentsmetrics');
            setMetrics(response.data.data || { onlineCount: '0', offlineCount: '0' });
        } catch (err) {
            console.error('Failed to fetch agent metrics:', err.message);
        }
    };

    const fetchAgents = async () => {
        try {
            const params = {
                limit,
                offset,
                sort: `${sortField}:${sortDirection}`,
                ...(hostnameFilter && { hostname: hostnameFilter }),
                ...(ipFilter && { ip: ipFilter }),
                ...(statusFilter !== 'All' && { status: statusFilter }),
                ...(tagFilter && { tag: tagFilter })
            };
            const response = await axios.get('/api/agents/search', { params });
            setAgents(response.data.data || []);
            setTotalCount(response.data.totalCount || 0);
        } catch (err) {
            setError(err.message);
        }
    };

    useEffect(() => {
        fetchAgents();
        fetchMetrics();
    }, [hostnameFilter, ipFilter, statusFilter, tagFilter, offset, sortField, sortDirection]);

    const handleSort = (field) => {
        if (sortField === field) {
            setSortDirection(prev => (prev === 'asc' ? 'desc' : 'asc'));
        } else {
            setSortField(field);
            setSortDirection('asc');
        }
    };

    const handleNextPage = () => {
        if (offset + limit < totalCount) {
            setOffset(prev => prev + limit);
        }
    };

    const handlePreviousPage = () => {
        setOffset(prev => Math.max(0, prev - limit));
    };

    const handleRowClick = (uuid) => {
        navigate(`/agent?agt=${uuid}`);
    };

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
                <input type="text" placeholder="Hostname" value={hostnameFilter} onChange={(e) => setHostnameFilter(e.target.value)} />
                <input type="text" placeholder="IP" value={ipFilter} onChange={(e) => setIpFilter(e.target.value)} />
                <input type="text" placeholder="Tag (key:value)" value={tagFilter} onChange={(e) => setTagFilter(e.target.value)} />
                <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
                    <option value="All">All</option>
                    <option value="Online">Online</option>
                    <option value="Offline">Offline</option>
                </select>
            </div>
    
            {agents.length > 0 ? (
                <>
                    <div className="table-container">
                        <table>
                            <thead>
                                <tr>
                                    <th onClick={() => handleSort('uuid')}>
                                        UUID {sortField === 'uuid' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th onClick={() => handleSort('agent_user')}>
                                        User {sortField === 'agent_user' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th onClick={() => handleSort('hostname')}>
                                        Hostname {sortField === 'hostname' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th onClick={() => handleSort('ip')}>
                                        IP {sortField === 'ip' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                    <th onClick={() => handleSort('status')}>
                                        Status {sortField === 'status' && (sortDirection === 'asc' ? '↑' : '↓')}
                                    </th>
                                </tr>
                            </thead>
                            <tbody>
                                {agents.map(agent => (
                                    <tr key={agent.uuid} onClick={() => handleRowClick(agent.uuid)} className="go-to-agent">
                                        <td>{agent.uuid.substring(0, 6)}</td>
                                        <td>{agent.username}</td>
                                        <td>{agent.hostname}</td>
                                        <td>{agent.agentip}</td>
                                        <td>{agent.status}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>              
                    <div className="pagination-controls">
                        <button onClick={handlePreviousPage} disabled={offset === 0}>Prev</button>
                        <span>Page {Math.floor(offset / limit) + 1} of {Math.ceil(totalCount / limit)}</span>
                        <button onClick={handleNextPage} disabled={offset + limit >= totalCount}>Next</button>
                    </div>
                </>
            ) : (
                <p className="no-agents-message">No Agents</p>
            )}
        </div>
    );
};

export default Home;
