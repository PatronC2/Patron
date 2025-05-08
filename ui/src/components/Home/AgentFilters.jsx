import React from 'react';
import './AgentFilters.css';

const AgentFilters = ({
    hostnameFilter, setHostnameFilter,
    ipFilter, setIpFilter,
    statusFilter, setStatusFilter,
    logic, setLogic,
    tagConditions, setTagConditions,
    tagOptions
}) => {

    const updateTagCondition = (index, field, value) => {
        const updated = [...tagConditions];
        updated[index][field] = value;
        setTagConditions(updated);
    };

    const addTagCondition = () => {
        setTagConditions(prev => [...prev, { key: '', value: '' }]);
    };

    const removeTagCondition = (index) => {
        setTagConditions(prev => prev.filter((_, i) => i !== index));
    };

    return (
        <div className="filters-side-panel">
            <input type="text" placeholder="Hostname" value={hostnameFilter} onChange={(e) => setHostnameFilter(e.target.value)} />
            <input type="text" placeholder="IP" value={ipFilter} onChange={(e) => setIpFilter(e.target.value)} />
            <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
                <option value="All">All</option>
                <option value="Online">Online</option>
                <option value="Offline">Offline</option>
            </select>

            <select value={logic} onChange={(e) => setLogic(e.target.value)}>
                <option value="or">Any Tag Match</option>
                <option value="and">All Tags Must Match</option>
            </select>

            {tagConditions.map((condition, index) => (
                <div key={index} className="tag-condition">
                    <select value={condition.key} onChange={(e) => updateTagCondition(index, 'key', e.target.value)}>
                        <option value="">Key</option>
                        {tagOptions.map(opt => (
                            <option key={opt.key} value={opt.key}>{opt.key}</option>
                        ))}
                    </select>
                    <select value={condition.value} onChange={(e) => updateTagCondition(index, 'value', e.target.value)} disabled={!condition.key}>
                        <option value="">Value</option>
                        {(tagOptions.find(opt => opt.key === condition.key)?.values || []).map(val => (
                            <option key={val} value={val}>{val}</option>
                        ))}
                    </select>
                    <button type="button" onClick={() => removeTagCondition(index)}>×</button>
                </div>
            ))}
            <button type="button" onClick={addTagCondition}>+ Tag</button>
        </div>
    );
};

export default AgentFilters;
