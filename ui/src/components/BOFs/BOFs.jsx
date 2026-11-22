import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAxios } from '../../context/AxiosProvider';
import './BOFs.css';

const BOFs = ({ isMenuOpen }) => {
    const axios = useAxios();

    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        fetchBOFs();
    }, []);

    const fetchBOFs = async () => {
        setLoading(true);
        setError('');

        try {
            const response = await axios.get('/api/bofs');
            setResults(response.data || []);
        } catch (err) {
            setError(err?.response?.data?.error || err.message || 'Failed to fetch BOFs');
        } finally {
            setLoading(false);
        }
    };

    const formatTimestamp = (timestamp) => {
        if (!timestamp) return '-';
        const date = new Date(timestamp);
        return date.toLocaleString();
    };

    const getPlatformIcon = (platform) => {
        switch (platform?.toLowerCase()) {
            case 'windows':
                return '🪟';
            case 'linux':
                return '🐧';
            default:
                return '💻';
        }
    };

    if (loading) {
        return (
            <div className="bofs-container">
                <div className="bofs-header">
                    <h1>BOFs</h1>
                    <p className="bofs-subtitle">Beacon Object Files Repository</p>
                </div>
                <div className="bofs-loading">Loading...</div>
            </div>
        );
    }

    return (
        <div className="bofs-container">
            <div className="bofs-header">
                <h1>BOFs</h1>
                <p className="bofs-subtitle">Beacon Object Files Repository</p>
            </div>

            {error && <div className="bofs-error">{error}</div>}

            <div className="bofs-results">
                <div className="bofs-results-header">
                    <h2>All BOFs</h2>
                    <div className="bofs-results-meta">
                        <span>Total: {results.length}</span>
                    </div>
                </div>

                <div className="bofs-table-wrapper">
                    <div className="bofs-table-container">
                        <table>
                            <thead>
                                <tr>
                                    <th className="bofs-time-col">Uploaded</th>
                                    <th>UUID</th>
                                    <th>Platform</th>
                                    <th>Function Name</th>
                                    <th>Description</th>
                                    <th className="bofs-time-col">Size</th>
                                </tr>
                            </thead>
                            <tbody>
                                {results.length === 0 ? (
                                    <tr>
                                        <td colSpan="6" className="bofs-empty-state">
                                            No BOFs found. Upload a BOF to get started.
                                        </td>
                                    </tr>
                                ) : (
                                    results.map((bof) => (
                                        <tr key={bof.file_id}>
                                            <td className="bofs-time-col">{formatTimestamp(bof.uploaded_at)}</td>
                                            <td>{bof.uuid}</td>
                                            <td>{getPlatformIcon(bof.platform)}</td>
                                            <td>{bof.function_name}</td>
                                            <td>{bof.description || '-'}</td>
                                            <td className="bofs-time-col">
                                                {(bof.uploaded_at && bof.uploaded_at.length > 0) ? 
                                                    new Date(bof.uploaded_at).toLocaleTimeString() : '-'}
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default BOFs;
