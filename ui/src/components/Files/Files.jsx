import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import qs from 'qs';
import { useAxios } from '../../context/AxiosProvider';
import './Files.css';

const Files = () => {
    const axios = useAxios();

    const [tagsInput, setTagsInput] = useState('');
    const [tagOptions, setTagOptions] = useState([]);
    const [tagKey, setTagKey] = useState('');
    const [tagValue, setTagValue] = useState('');
    const [tagLogic, setTagLogic] = useState('or');
    const [limit, setLimit] = useState(20);
    const [offset, setOffset] = useState(0);
    const [total, setTotal] = useState(0);
    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const parseTags = (input) => {
        return input
            .split(',')
            .map(tag => tag.trim())
            .filter(Boolean);
    };

    const selectedTags = parseTags(tagsInput);

    useEffect(() => {
        const fetchTagOptions = async () => {
            try {
                const response = await axios.get('/api/tags/options');
                setTagOptions(response.data.tags || []);
            } catch (err) {
                setTagOptions([]);
            }
        };

        fetchTagOptions();
    }, [axios]);

    const fetchFiles = async (nextOffset = 0) => {
        setLoading(true);
        setError('');

        try {
            const tags = parseTags(tagsInput);
            const params = {
                ...(tags.length > 0 && { tag: tags }),
                logic: tagLogic,
                limit,
                offset: nextOffset,
            };

            const response = await axios.get('/api/files/list', {
                params,
                paramsSerializer: (queryParams) => qs.stringify(queryParams, { arrayFormat: 'repeat' }),
            });

            setResults(response.data.data || []);
            setTotal(response.data.totalCount || 0);
            setOffset(nextOffset);
        } catch (err) {
            setError(err?.response?.data?.error || err.message || 'File search failed');
        } finally {
            setLoading(false);
        }
    };

    const addTagToken = (token) => {
        const next = tagsInput
            .split(',')
            .map(tag => tag.trim())
            .filter(Boolean);

        if (!next.includes(token)) {
            next.push(token);
        }

        setTagsInput(next.join(', '));
    };

    const removeTagToken = (token) => {
        const next = tagsInput
            .split(',')
            .map(tag => tag.trim())
            .filter(Boolean)
            .filter(tag => tag !== token);

        setTagsInput(next.join(', '));
    };

    const handleAddTagSelection = () => {
        if (!tagKey || !tagValue) return;

        const token = `${tagKey}:${tagValue}`;
        addTagToken(token);
        setTagKey('');
        setTagValue('');
    };

    const hasNextPage = offset + limit < total;

    const handleNext = () => {
        if (!hasNextPage) return;
        fetchFiles(offset + limit);
    };

    const handlePrev = () => {
        fetchFiles(Math.max(0, offset - limit));
    };

    const handleDownloadFile = async (fileId) => {
        try {
            const response = await axios.get(`/api/files/download/${fileId}`, {
                responseType: 'blob',
            });

            const contentDisposition = response.headers['content-disposition'];
            let filename = 'file';

            if (contentDisposition && contentDisposition.indexOf('attachment') !== -1) {
                const filenameMatch = contentDisposition.match(/filename=["']?([^"']+)["']?/);
                if (filenameMatch && filenameMatch[1]) {
                    filename = filenameMatch[1];
                }
            }

            const url = window.URL.createObjectURL(new Blob([response.data]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', filename);
            document.body.appendChild(link);
            link.click();
            link.remove();
            window.URL.revokeObjectURL(url);
        } catch (err) {
            setError('Failed to download file');
        }
    };

    const formatTimestamp = (ts) => {
        if (!ts) return '-';

        const date = new Date(ts);
        if (Number.isNaN(date.getTime())) return ts;

        return date.toLocaleString(undefined, {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
        });
    };

    useEffect(() => {
        fetchFiles(0);
    }, [tagsInput, tagLogic, limit]);

    return (
        <div className="files-container">
            <div className="files-header">
                <h1>Files</h1>
                <p className="files-subtitle">Browse and filter file transfers by agent tags.</p>
            </div>

            <div className="files-form">
                <div className="files-form-inner">
                    <div className="files-panel files-panel-left">
                        <div className="files-row files-page-size-row">
                            <label htmlFor="limit">Page Size</label>
                            <input
                                id="limit"
                                type="number"
                                min="1"
                                max="500"
                                value={limit}
                                onChange={(e) => setLimit(Number(e.target.value || 20))}
                            />
                        </div>
                    </div>

                    <div className="files-panel files-panel-right">
                        <div className="files-row">
                            <label>Tag Quick Add</label>
                            <div className="files-tag-quick-row">
                                <select
                                    id="tagKey"
                                    value={tagKey}
                                    onChange={(e) => {
                                        setTagKey(e.target.value);
                                        setTagValue('');
                                    }}
                                >
                                    <option value="">Key</option>
                                    {tagOptions.map((opt) => (
                                        <option key={opt.key} value={opt.key}>{opt.key}</option>
                                    ))}
                                </select>
                                <select
                                    id="tagValue"
                                    value={tagValue}
                                    onChange={(e) => setTagValue(e.target.value)}
                                    disabled={!tagKey}
                                >
                                    <option value="">Value (Required)</option>
                                    {(tagOptions.find((opt) => opt.key === tagKey)?.values || []).map((val) => (
                                        <option key={val} value={val}>{val}</option>
                                    ))}
                                </select>
                                <button
                                    type="button"
                                    className="files-tag-add-button"
                                    onClick={handleAddTagSelection}
                                    disabled={!tagKey || !tagValue}
                                >
                                    Add Tag
                                </button>
                                <select id="tagLogic" value={tagLogic} onChange={(e) => setTagLogic(e.target.value)}>
                                    <option value="and">Tag Logic: AND</option>
                                    <option value="or">Tag Logic: OR</option>
                                </select>
                            </div>
                        </div>
                        <div className="files-row">
                            <label>Selected Tags</label>
                            <div className="files-tag-selected-row">
                                {selectedTags.length === 0 ? (
                                    <span className="files-tag-empty">None</span>
                                ) : (
                                    selectedTags.map((tag) => (
                                        <button
                                            key={tag}
                                            type="button"
                                            className="files-tag-pill files-tag-pill-removable"
                                            onClick={() => removeTagToken(tag)}
                                            title={`${tag} (click to remove)`}
                                        >
                                            {tag} ×
                                        </button>
                                    ))
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {error && <div className="files-error">{error}</div>}

            <div className="files-results">
                <div className="files-results-header">
                    <h2>Results</h2>
                    <div className="files-results-meta">
                        <span>Total: {total}</span>
                        <span>Offset: {offset}</span>
                    </div>
                </div>
                <div className="files-table-wrapper">
                    <div className="files-table-container">
                        <table>
                            <thead>
                                <tr>
                                    <th className="files-time-col">Created</th>
                                    <th>Agent</th>
                                    <th>Type</th>
                                    <th>Path</th>
                                    <th>Status</th>
                                    <th className="files-time-col">Updated</th>
                                    <th>Download</th>
                                </tr>
                            </thead>
                            <tbody>
                                {results.length === 0 ? (
                                    <tr>
                                        <td colSpan="7" className="files-empty-state">
                                            No results
                                        </td>
                                    </tr>
                                ) : (
                                    results.map((file) => (
                                        <tr key={file.file_id}>
                                            <td className="files-time-col">{formatTimestamp(file.created_at)}</td>
                                            <td>
                                                <Link className="files-agent-link" to={`/agent?agt=${file.uuid}`}>
                                                    Open
                                                </Link>
                                            </td>
                                            <td>{file.type}</td>
                                            <td className="files-path-cell" title={file.path}>{file.path}</td>
                                            <td>{file.status}</td>
                                            <td className="files-time-col">{formatTimestamp(file.updated_at)}</td>
                                            <td>
                                                <button
                                                    type="button"
                                                    className="files-download-button"
                                                    onClick={() => handleDownloadFile(file.file_id)}
                                                >
                                                    Download
                                                </button>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
                <div className="files-pagination-controls">
                    <button type="button" onClick={handlePrev} disabled={offset === 0 || loading}>Prev</button>
                    <button type="button" onClick={handleNext} disabled={loading || !hasNextPage}>Next</button>
                </div>
            </div>
        </div>
    );
};

export default Files;
