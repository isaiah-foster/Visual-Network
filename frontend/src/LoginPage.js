import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './AuthContext';

function LoginPage() {
  const [userId, setUserId] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [error, setError] = React.useState('');
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
    <div style={{ maxWidth: '400px', margin: '100px auto', padding: '20px' }}>
      <h1>Login</h1>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '10px' }}>
          <label>User ID</label>
          <input
            type="text"
            value={userId}
            onChange={e => setUserId(e.target.value)}
            style={{ display: 'block', width: '100%', padding: '8px', marginTop: '4px' }}
          />
        </div>
        <div style={{ marginBottom: '10px' }}>
          <label>Password</label>
          <input
            type="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            style={{ display: 'block', width: '100%', padding: '8px', marginTop: '4px' }}
          />
        </div>
        <button type="submit" style={{ padding: '8px 16px' }}>Login</button>
      </form>
    </div>
  );
}

export default LoginPage;