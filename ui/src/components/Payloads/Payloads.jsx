import React, { useEffect, useMemo, useState, useContext } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import NewPayloadForm from './NewPayloadForm';
import { useTable } from 'react-table';
import { ResizableBox } from 'react-resizable';
import 'react-resizable/css/styles.css';
import './Payloads.css';

const Payloads = () => {
    const cfg = window.runtimeConfig;
    const FILE_SERVER = `https://${cfg.REACT_APP_NGINX_IP}:${cfg.REACT_APP_NGINX_PORT}/fileserver/`;
    const axios = useAxios();
    const { auth } = useContext(AuthContext);
    const [data, setData] = useState([]);
    const [error, setError] = useState(null);
    const [activeTab, setActiveTab] = useState('current_payloads');
    const [notification, setNotification] = useState('');
    const [notificationType, setNotificationType] = useState('');

    useEffect(() => {
        document.body.classList.add('payloads-page');
        fetchData();
        const interval = setInterval(fetchData, 10000);
        return () => {
            document.body.classList.remove('payloads-page');
            clearInterval(interval);
        };
    }, []);

    const fetchData = async () => {
        try {
            const response = await axios.get('/api/payloads');
            const responseData = response.data.data;
            setData(Array.isArray(responseData) ? responseData : []);
        } catch (err) {
            setError(err.message);
        }
    };

    const handleDelete = async (payloadid) => {
        try {
            await axios.delete(`/api/payloads/${payloadid}`);
            setNotification('Payload deleted successfully!');
            setNotificationType('success');
            fetchData();
        } catch {
            setNotification('Failed to delete payload.');
            setNotificationType('error');
        } finally {
            setTimeout(() => setNotification(''), 3000);
        }
    };

    const columns = useMemo(() => [
        {
            Header: 'UUID',
            accessor: row => row.uuid.substring(0, 6),
            minWidth: 80
        },
        {
            Header: 'Name',
            accessor: 'concat',
            Cell: ({ value }) => (
                <div className="cell-content">
                    <a href={`${FILE_SERVER}${value}`} target="_blank" rel="noopener noreferrer">{value}</a>
                </div>
            ),
            minWidth: 150
        },
        {
            Header: 'Description',
            accessor: 'description',
            minWidth: 120
        },
        {
            Header: 'Listener IP',
            accessor: 'serverip',
            minWidth: 130
        },
        {
            Header: 'Listener Port',
            accessor: 'serverport',
            minWidth: 100
        },
        {
            Header: 'Callback Frequency',
            accessor: 'callbackfrequency',
            minWidth: 150
        },
        {
            Header: 'Callback Jitter',
            accessor: 'callbackjitter',
            minWidth: 120
        },
        {
            Header: 'Action',
            accessor: 'payloadid',
            disableResizing: true,
            minWidth: 100,
            Cell: ({ value }) => (
                <div className="cell-content no-ellipsis">
                    <button onClick={() => handleDelete(value)} className="delete-button">Delete</button>
                </div>
            )
        }
    ], []);

    const {
        getTableProps,
        getTableBodyProps,
        headerGroups,
        rows,
        prepareRow,
    } = useTable({ columns, data });

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
        <div className="payloads-content">
            <div className="header">
                <h1>Payloads</h1>
                <div className="header-buttons">
                    <button className={activeTab === 'current_payloads' ? 'active' : ''} onClick={() => setActiveTab('current_payloads')}>Existing Payloads</button>
                    <button className={activeTab === 'new' ? 'active' : ''} onClick={() => setActiveTab('new')}>Create New Payload</button>
                </div>
            </div>

            {activeTab === 'current_payloads' ? (
                <div className="payloads-container">
                    {data.length > 0 ? (
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
                                                        <div className="cell-content">{cell.render('Cell')}</div>
                                                    </td>
                                                ))}
                                            </tr>
                                        );
                                    })}
                                </tbody>
                            </table>
                        </div>
                    ) : (
                        <p>No Payloads</p>
                    )}
                </div>
            ) : (
                <NewPayloadForm fetchData={fetchData} setActiveTab={setActiveTab} />
            )}

            {notification && (
                <div className={`notification ${notificationType}`}>{notification}</div>
            )}
        </div>
    );
};

export default Payloads;
