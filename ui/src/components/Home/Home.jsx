import React, { useEffect, useState, useContext, useRef } from 'react';
import qs from 'qs';
import { useAxios } from '../../context/AxiosProvider';
import { useNavigate, useLocation } from 'react-router-dom';
import AuthContext from '../../context/AuthProvider';
import AgentFilters from './AgentFilters';
import './Home.css';

const Home = ({ isMenuOpen }) => {
    const axios = useAxios();
    const { auth } = useContext(AuthContext);
    const navigate = useNavigate();
    const location = useLocation();
    const didInitFromQuery = useRef(false);

    const [error, setError] = useState(null);
    const [agents, setAgents] = useState([]);
    const [metrics, setMetrics] = useState({ onlineCount: '0', offlineCount: '0' });
    const [now, setNow] = useState(Date.now());

    const [hostnameFilter, setHostnameFilter] = useState('');
    const [ipFilter, setIpFilter] = useState('');
    const [statusFilter, setStatusFilter] = useState('Online');
    const [tagConditions, setTagConditions] = useState([{ key: '', value: '' }]);
    const [tagOptions, setTagOptions] = useState([]);
    const [logic, setLogic] = useState('or');

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

    const fetchTagOptions = async () => {
        try {
            const response = await axios.get('/api/tags/options');
            setTagOptions(response.data.tags || []);
        } catch (err) {
            console.error('Failed to fetch tag options:', err.message);
        }
    };

    const fetchAgents = async () => {
        try {
            const params = {
                limit,
                offset,
                logic,
                sort: `${sortField}:${sortDirection}`,
                ...(hostnameFilter && { hostname: hostnameFilter }),
                ...(ipFilter && { ip: ipFilter }),
                ...(statusFilter !== 'All' && { status: statusFilter })
            };
            const tags = tagConditions.filter(tc => tc.key && tc.value).map(tc => `${tc.key}:${tc.value}`);
            if (tags.length > 0) params.tag = tags;
            const response = await axios.get('/api/agents/search', {
                params,
                paramsSerializer: params => qs.stringify(params, { arrayFormat: 'repeat' })
            });
            setAgents(response.data.data || []);
            setTotalCount(response.data.totalCount || 0);
        } catch (err) {
            setError(err.message);
        }
    };

    useEffect(() => {
        fetchTagOptions();
        fetchMetrics();
        fetchAgents();
    }, []);

    useEffect(() => {
        if (didInitFromQuery.current) return;
        const params = qs.parse(location.search, { ignoreQueryPrefix: true });
        if (params.hostname) setHostnameFilter(params.hostname);
        if (params.ip) setIpFilter(params.ip);
        if (params.status) setStatusFilter(params.status);
        if (params.logic) setLogic(params.logic);
        if (params.sort) {
            const [field, direction] = String(params.sort).split(':');
            if (field) setSortField(field);
            if (direction) setSortDirection(direction);
        }
        if (params.offset) {
            const parsedOffset = parseInt(params.offset, 10);
            if (!Number.isNaN(parsedOffset)) setOffset(parsedOffset);
        }
        const tagParams = params.tag
            ? (Array.isArray(params.tag) ? params.tag : [params.tag])
            : [];
        if (tagParams.length > 0) {
            const nextConditions = tagParams.map((t) => {
                const [key, ...rest] = String(t).split(':');
                return { key, value: rest.join(':') };
            });
            setTagConditions(nextConditions);
        }
        didInitFromQuery.current = true;
    }, [location.search]);

    useEffect(() => {
        const id = setInterval(() => setNow(Date.now()), 1000);
        return () => clearInterval(id);
    }, []);

    useEffect(() => {
        const id = setInterval(() => {
            fetchMetrics();
            fetchAgents();
        }, 10000);
        return () => clearInterval(id);
    }, [offset, hostnameFilter, ipFilter, statusFilter, tagConditions, logic, sortField, sortDirection]);

    useEffect(() => {
        setOffset(0);
    }, [hostnameFilter, ipFilter, statusFilter, tagConditions, logic, sortField, sortDirection]);

    useEffect(() => {
        fetchAgents();
    }, [offset, hostnameFilter, ipFilter, statusFilter, tagConditions, logic, sortField, sortDirection]);

    useEffect(() => {
        if (!didInitFromQuery.current) return;
        const params = {
            ...(hostnameFilter && { hostname: hostnameFilter }),
            ...(ipFilter && { ip: ipFilter }),
            ...(statusFilter && { status: statusFilter }),
            ...(logic && { logic }),
            ...(offset > 0 && { offset }),
            sort: `${sortField}:${sortDirection}`
        };
        const tags = tagConditions.filter(tc => tc.key && tc.value).map(tc => `${tc.key}:${tc.value}`);
        if (tags.length > 0) params.tag = tags;
        const query = qs.stringify(params, { arrayFormat: 'repeat' });
        navigate({ pathname: '/home', search: query ? `?${query}` : '' }, { replace: true });
    }, [hostnameFilter, ipFilter, statusFilter, tagConditions, logic, sortField, sortDirection, offset, navigate]);

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

    const formatCountdown = (nextCallback) => {
        if (!nextCallback) return '—';
        const ts = new Date(nextCallback).getTime();
        if (Number.isNaN(ts)) return '—';
        const delta = Math.floor((ts - now) / 1000);
        if (delta <= 0) return 'due';
        return `${delta}s`;
    };

    return (
        <div className="home-container horizontal-layout">
            <div className="main-content-column">
                <div className="header-wrapper">
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
                </div>
                <div className="table-wrapper">
                    {agents.length > 0 ? (
                        <>
                            <div className="table-container">
                                <table>
                                    <thead>
                                        <tr>
                                            <th onClick={() => handleSort('agent_user')}>
                                                User {sortField === 'agent_user' && (sortDirection === 'asc' ? '↑' : '↓')}
                                            </th>
                                            <th onClick={() => handleSort('hostname')}>
                                                Hostname {sortField === 'hostname' && (sortDirection === 'asc' ? '↑' : '↓')}
                                            </th>
                                            <th onClick={() => handleSort('ip')}>
                                                IP {sortField === 'ip' && (sortDirection === 'asc' ? '↑' : '↓')}
                                            </th>
                                            <th>Next Callback</th>
                                            <th onClick={() => handleSort('status')}>
                                                Status {sortField === 'status' && (sortDirection === 'asc' ? '↑' : '↓')}
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {agents.map(agent => (
                                            <tr key={agent.uuid} onClick={() => handleRowClick(agent.uuid)} className="go-to-agent">
                                                <td>{agent.username}</td>
                                                <td>{agent.hostname}</td>
                                                <td>{agent.agentip}</td>
                                                <td>{formatCountdown(agent.nextcallback)}</td>
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
            </div>
            <AgentFilters
                hostnameFilter={hostnameFilter} setHostnameFilter={setHostnameFilter}
                ipFilter={ipFilter} setIpFilter={setIpFilter}
                statusFilter={statusFilter} setStatusFilter={setStatusFilter}
                logic={logic} setLogic={setLogic}
                tagConditions={tagConditions} setTagConditions={setTagConditions}
                tagOptions={tagOptions}
            />
        </div>
    );
};

export default Home;
