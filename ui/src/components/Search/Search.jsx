import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import qs from 'qs';
import { useAxios } from '../../context/AxiosProvider';
import './Search.css';

const Search = () => {
    const axios = useAxios();

    const [q, setQ] = useState('');
    const [ip, setIp] = useState('');
    const [tagsInput, setTagsInput] = useState('');
    const [tagOptions, setTagOptions] = useState([]);
    const [tagKey, setTagKey] = useState('');
    const [tagValue, setTagValue] = useState('');
    const [tagLogic, setTagLogic] = useState('and');
    const [startDate, setStartDate] = useState('');
    const [startTime, setStartTime] = useState('');
    const [endDate, setEndDate] = useState('');
    const [endTime, setEndTime] = useState('');
    const [limit, setLimit] = useState(50);
    const [offset, setOffset] = useState(0);
    const [total, setTotal] = useState(null);
    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [sortDirection, setSortDirection] = useState('desc');
    const [activeContent, setActiveContent] = useState(null);

    const parseTags = (input) => {
        return input
            .split(',')
            .map(t => t.trim())
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

    const fetchResults = async (nextOffset = 0, sortOverride = null) => {
        setLoading(true);
        setError('');
        try {
            const tags = parseTags(tagsInput);
            const buildDateTime = (date, time) => {
                if (!date) return '';
                const t = time ? time : '00:00';
                return `${date}T${t}`;
            };
            const effectiveSort = sortOverride || sortDirection;
            const params = {
                ...(q && { q }),
                ...(ip && { ip }),
                ...(startDate && { start: buildDateTime(startDate, startTime) }),
                ...(endDate && { end: buildDateTime(endDate, endTime) }),
                ...(tags.length > 0 && { tag: tags }),
                tag_logic: tagLogic,
                sort: `created_at:${effectiveSort}`,
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

    const toggleSort = () => {
        const next = sortDirection === 'asc' ? 'desc' : 'asc';
        setSortDirection(next);
        fetchResults(0, next);
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

    const removeTagToken = (token) => {
        const next = tagsInput
            .split(',')
            .map(t => t.trim())
            .filter(Boolean)
            .filter(t => t !== token);
        setTagsInput(next.join(', '));
    };

    const handleAddTagSelection = () => {
        if (!tagKey) return;
        const token = tagValue ? `${tagKey}:${tagValue}` : tagKey;
        addTagToken(token);
        setTagKey('');
        setTagValue('');
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
                <p className="search-subtitle">Search contents, filter by IP, tags, and time range.</p>
            </div>

            <form className="search-form" onSubmit={handleSearch}>
                <div className="search-panel search-panel-left">
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
                        <label htmlFor="startDate">Start Date</label>
                        <input
                            id="startDate"
                            type="date"
                            value={startDate}
                            onChange={e => setStartDate(e.target.value)}
                        />
                    </div>
                    <div className="search-row">
                        <label htmlFor="startTime">Start Time</label>
                        <input
                            id="startTime"
                            type="time"
                            value={startTime}
                            onChange={e => setStartTime(e.target.value)}
                        />
                    </div>
                    <div className="search-row">
                        <label htmlFor="endDate">End Date</label>
                        <input
                            id="endDate"
                            type="date"
                            value={endDate}
                            onChange={e => setEndDate(e.target.value)}
                        />
                    </div>
                    <div className="search-row">
                        <label htmlFor="endTime">End Time</label>
                        <input
                            id="endTime"
                            type="time"
                            value={endTime}
                            onChange={e => setEndTime(e.target.value)}
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
                </div>
                <div className="search-panel search-panel-right">
                    <div className="search-row">
                        <label>Tag Quick Add</label>
                        <div className="tag-quick-row">
                            <select
                                id="tagKey"
                                value={tagKey}
                                onChange={e => {
                                    setTagKey(e.target.value);
                                    setTagValue('');
                                }}
                            >
                                <option value="">Key</option>
                                {tagOptions.map(opt => (
                                    <option key={opt.key} value={opt.key}>{opt.key}</option>
                                ))}
                            </select>
                            <select
                                id="tagValue"
                                value={tagValue}
                                onChange={e => setTagValue(e.target.value)}
                                disabled={!tagKey}
                            >
                                <option value="">Value (Any)</option>
                                {(tagOptions.find(opt => opt.key === tagKey)?.values || []).map(val => (
                                    <option key={val} value={val}>{val}</option>
                                ))}
                            </select>
                            <button type="button" className="tag-add-button" onClick={handleAddTagSelection}>
                                Add Tag
                            </button>
                            <select id="tagLogic" value={tagLogic} onChange={e => setTagLogic(e.target.value)}>
                                <option value="and">Tag Logic: AND</option>
                                <option value="or">Tag Logic: OR</option>
                            </select>
                        </div>
                    </div>
                    <div className="search-row">
                        <label>Selected Tags</label>
                        <div className="tag-selected-row">
                            {selectedTags.length === 0 ? (
                                <span className="tag-empty">None</span>
                            ) : (
                                selectedTags.map(tag => (
                                    <button
                                        key={tag}
                                        type="button"
                                        className="tag-pill removable"
                                        onClick={() => removeTagToken(tag)}
                                        title="Remove tag"
                                    >
                                        {tag} ×
                                    </button>
                                ))
                            )}
                        </div>
                    </div>
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
                                    <th className="sortable time-col" onClick={toggleSort}>
                                        Time {sortDirection === 'asc' ? '↑' : '↓'}
                                    </th>
                                    <th>Agent</th>
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
                                                <td className="time-col">{formatTimestamp(src.created_at)}</td>
                                                <td>
                                                    <Link className="agent-link" to={`/agent?agt=${src.uuid}`}>
                                                        Open
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
