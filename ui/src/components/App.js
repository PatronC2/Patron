import React, { useState } from 'react';
import Login from './Login';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false); // Initially not logged in

  const handleSuccessfulLogin = () => {
    setIsLoggedIn(true);
  };

  return (
    <Router>
      <Routes>
        <Route path="/" element={<Login onSuccessfulLogin={handleSuccessfulLogin} />} />
      </Routes>
    </Router>
  );
}

export default App;