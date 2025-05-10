import React, { useEffect, useMemo, useState, useContext } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import NewRedirectorForm from './NewRedirectorForm';
import { useTable } from 'react-table';
import { ResizableBox } from 'react-resizable';
import 'react-resizable/css/styles.css';
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
        const interval = setInterval(fetchData, 10000);
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
            setData(Array.isArray(responseData) ? responseData : []);
        } catch (err) {
            setError(err.message);
        }
    };

    const filteredData = useMemo(() => {
        return data.filter(item => statusFilter === 'All' || item.status === statusFilter);
    }, [data, statusFilter]);

    const columns = useMemo(() => [
        { Header: 'Name', accessor: 'name', minWidth: 120 },
        { Header: 'Description', accessor: 'description', minWidth: 150 },
        { Header: 'Forward IP', accessor: 'forwardip', minWidth: 130 },
        { Header: 'Forward Port', accessor: 'forwardport', minWidth: 100 },
        { Header: 'Listener Port', accessor: 'listenport', minWidth: 100 },
        { Header: 'Status', accessor: 'status', minWidth: 100 },
    ], []);

    const {
        getTableProps,
        getTableBodyProps,
        headerGroups,
        rows,
        prepareRow,
    } = useTable({ columns, data: filteredData });

    const renderResizableHeader = (column, index) => {
        const width = column.minWidth || 150;
        return (
            <th key={index} {...column.getHeaderProps()}>
                <ResizableBox
                    width={width}
                    height={30}
                    axis="x"
                    resizeHandles={['e']}
                    minConstraints={[50, 30]}
                    maxConstraints={[300, 30]}
                >
                    <div className="header-content" title={column.render('Header')}>
                        {column.render('Header')}
                    </div>
                </ResizableBox>
            </th>
        );
    };

    if (error) return <div>Error: {error}</div>;

    return (
        <div className="redirector-container">
            <div className="header">
                <h1>Redirectors</h1>
                <div className="header-buttons">
                    <button
                        className={activeTab === 'current_redirectors' ? 'active' : ''}
                        onClick={() => setActiveTab('current_redirectors')}
                    >
                        Existing Redirectors
                    </button>
                    <button
                        className={activeTab === 'new' ? 'active' : ''}
                        onClick={() => setActiveTab('new')}
                    >
                        Create New Redirector
                    </button>
                </div>
            </div>

            {activeTab === 'current_redirectors' ? (
                <div className="redirectors-container">
                    <div className="status-boxes">
                        <div className="status-box online">
                            <p>Online</p>
                            <h2>{data.filter(d => d.status === 'Online').length}</h2>
                        </div>
                        <div className="status-box offline">
                            <p>Offline</p>
                            <h2>{data.filter(d => d.status === 'Offline').length}</h2>
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
                        <div className="table-wrapper">
                            <table {...getTableProps()}>
                                <thead>
                                    {headerGroups.map(headerGroup => (
                                        <tr {...headerGroup.getHeaderGroupProps()}>
                                            {headerGroup.headers.map(renderResizableHeader)}
                                        </tr>
                                    ))}
                                </thead>
                                <tbody {...getTableBodyProps()}>
                                    {rows.map(row => {
                                        prepareRow(row);
                                        return (
                                            <tr {...row.getRowProps()}>
                                                {row.cells.map(cell => (
                                                    <td {...cell.getCellProps()}>
                                                        <div className="cell-content">
                                                            {cell.render('Cell')}
                                                        </div>
                                                    </td>
                                                ))}
                                            </tr>
                                        );
                                    })}
                                </tbody>
                            </table>
                        </div>
                    ) : (
                        <p>No Redirectors</p>
                    )}
                </div>
            ) : (
                <NewRedirectorForm fetchData={fetchData} setActiveTab={setActiveTab} />
            )}
        </div>
    );
};

export default Redirectors;
