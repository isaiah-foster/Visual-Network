import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './AuthContext';

function LoginPage() {
  const [userId, setUserId] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { setUser } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    // step 1 - login
    const loginRes = await fetch(`${process.env.REACT_APP_AUTH_URL}/auth/login`, {
      method: 'POST',
      mode: 'cors',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ user_id: userId, password }),
    });

    if (!loginRes.ok) {
      setError('Invalid credentials');
      return;
    }

    // step 2 - fetch user info and store in context
    const meRes = await fetch(`${process.env.REACT_APP_AUTH_URL}/auth/me`, {
      credentials: 'include',
    });

    if (meRes.ok) {
      const data = await meRes.json();
      setUser(data);
      navigate('/');
    } else {
      setError('Login failed');
    }
  };

  return (
    <div className="form-wrap">
      <div className="form-card">
        <h1 className="form-title">Login</h1>
        <p className="form-text">Sign in to view live network dashboards.</p>
        {error && <p className="form-message form-message-error">{error}</p>}
        <form onSubmit={handleSubmit}>
          <div className="field">
            <label>User ID</label>
            <input
              type="text"
              value={userId}
              onChange={e => setUserId(e.target.value)}
            />
          </div>
          <div className="field">
            <label>Password</label>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
            />
          </div>
          <button className="button button-primary" type="submit">Login</button>
        </form>
      </div>
    </div>
  );
}

export default LoginPage;