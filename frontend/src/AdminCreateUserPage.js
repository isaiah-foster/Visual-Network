import React, { useState } from 'react';

function AdminCreateUserPage() {
  const [form, setForm] = useState({ user_id: '', email: '', password: '', role: 'user' });
  const [message, setMessage] = useState('');

  const handleChange = (e) => setForm({ ...form, [e.target.name]: e.target.value });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setMessage('');

    const response = await fetch(`${process.env.REACT_APP_AUTH_URL}/auth/admin/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(form),
    });

    if (response.ok) {
      setMessage('User created successfully');
      setForm({ user_id: '', email: '', password: '', role: 'user' });
    } else if (response.status === 409) {
      setMessage('User already exists');
    } else {
      setMessage('Error creating user');
    }
  };

  return (
    <div className="form-wrap">
      <div className="form-card">
        <h1 className="form-title">Create User</h1>
        <p className="form-text">Add a new account and assign its role.</p>
        {message && (
          <p className={`form-message ${message.includes('successfully') ? 'form-message-success' : 'form-message-error'}`}>
            {message}
          </p>
        )}
        <form onSubmit={handleSubmit}>
          <div className="field">
            <label>User ID</label>
            <input name="user_id" value={form.user_id} onChange={handleChange} placeholder="User ID" />
          </div>

          <div className="field">
            <label>Email</label>
            <input name="email" value={form.email} onChange={handleChange} placeholder="Email" />
          </div>

          <div className="field">
            <label>Password</label>
            <input name="password" value={form.password} onChange={handleChange} placeholder="Password" type="password" />
          </div>

          <div className="field">
            <label>Role</label>
            <select name="role" value={form.role} onChange={handleChange}>
              <option value="user">User</option>
              <option value="admin">Admin</option>
            </select>
          </div>

          <button className="button button-primary" type="submit">Create User</button>
        </form>
      </div>
    </div>
  );
}

export default AdminCreateUserPage;