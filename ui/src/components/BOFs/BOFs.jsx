import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAxios } from '../../context/AxiosProvider';
import './BOFs.css';

const BOFs = ({ isMenuOpen }) => {
    const axios = useAxios();

    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [showUploadModal, setShowUploadModal] = useState(false);
    const [uploading, setUploading] = useState(false);
    const [uploadForm, setUploadForm] = useState({
        uuid: '',
        platform: 'windows',
        functionName: '',
        description: ''
    });
    const [fileInput, setFileInput] = useState(null);


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

        const handleCloseModal = () => {
        if (!uploading) setShowUploadModal(false);
    };

    const handleFileChange = (e) => {
        setFileInput(e.target.files[0]);
    };

    const handleSubmitUpload = async (e) => {
        e.preventDefault();
        if (!fileInput) {
            setError('Please select a file');
            return;
        }

        setUploading(true);
        setError('');

        const formData = new FormData();
        formData.append('uuid', uploadForm.uuid);
        formData.append('platform', uploadForm.platform);
        formData.append('function_name', uploadForm.functionName);
        formData.append('description', uploadForm.description);
        formData.append('payload', fileInput);

        try {
            const response = await axios.post('/api/bofs', formData, {
                headers: { 'Content-Type': 'multipart/form-data' }
            });
            setShowUploadModal(false);
            setFileInput(null);
            setUploadForm({ uuid: '', platform: 'windows', functionName: '', description: '' });
            await fetchBOFs();
        } catch (err) {
            setError(err?.response?.data?.error || err.message || 'Failed to upload BOF');
        } finally {
            setUploading(false);
        }
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
                <button 
                    className="upload-btn"
                    onClick={() => setShowUploadModal(true)}
                >
                    📤 Upload BOF
                </button>

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
                <button 
                    className="upload-btn"
                    onClick={() => setShowUploadModal(true)}
                >
                    📤 Upload BOF
                </button>

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

                {/* Upload Modal */}
                {showUploadModal && (
                    <div className="modal-overlay" onClick={handleCloseModal}>
                        <div className="modal-content" onClick={(e) => e.stopPropagation()}>
                            <div className="modal-header">
                                <h2>Upload BOF</h2>
                                <button className="close-btn" onClick={handleCloseModal}>×</button>
                            </div>
                            
                            <form onSubmit={handleSubmitUpload} className="upload-form">
                                <div className="form-group">
                                    <label>UUID (Optional)</label>
                                    <input
                                        type="text"
                                        value={uploadForm.uuid}
                                        onChange={(e) => setUploadForm({...uploadForm, uuid: e.target.value})}
                                        placeholder="Leave empty for auto-generated UUID"
                                    />
                                </div>

                                <div className="form-group">
                                    <label>Platform</label>
                                    <select
                                        value={uploadForm.platform}
                                        onChange={(e) => setUploadForm({...uploadForm, platform: e.target.value})}
                                    >
                                        <option value="windows">Windows</option>
                                        <option value="linux">Linux</option>
                                    </select>
                                </div>

                                <div className="form-group">
                                    <label>Function Name *</label>
                                    <input
                                        type="text"
                                        value={uploadForm.functionName}
                                        onChange={(e) => setUploadForm({...uploadForm, functionName: e.target.value})}
                                        placeholder="e.g., go, powershell, etc."
                                        required
                                    />
                                </div>

                                <div className="form-group">
                                    <label>Description (Optional)</label>
                                    <textarea
                                        value={uploadForm.description}
                                        onChange={(e) => setUploadForm({...uploadForm, description: e.target.value})}
                                        placeholder="Brief description of this BOF"
                                        rows="3"
                                    />
                                </div>

                                <div className="form-group">
                                    <label>BOF File *</label>
                                    <input
                                        type="file"
                                        accept=".dll,.so,.exe"
                                        onChange={handleFileChange}
                                        required
                                    />
                                    {fileInput && (
                                        <p className="file-name">Selected: {fileInput.name}</p>
                                    )}
                                </div>

                                {error && <div className="form-error">{error}</div>}

                                <div className="form-actions">
                                    <button type="button" onClick={handleCloseModal} disabled={uploading}>Cancel</button>
                                    <button type="submit" disabled={uploading}>
                                        {uploading ? 'Uploading...' : 'Upload BOF'}
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );

};

export default BOFs;
