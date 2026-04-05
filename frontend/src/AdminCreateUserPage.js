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
    <form onSubmit={handleSubmit}>
      <h1>Create User</h1>
      {message && <p>{message}</p>}
      <input name="user_id" value={form.user_id} onChange={handleChange} placeholder="User ID" />
      <input name="email" value={form.email} onChange={handleChange} placeholder="Email" />
      <input name="password" value={form.password} onChange={handleChange} placeholder="Password" type="password" />
      <select name="role" value={form.role} onChange={handleChange}>
        <option value="user">User</option>
        <option value="admin">Admin</option>
      </select>
      <button type="submit">Create User</button>
    </form>
  );
}

export default AdminCreateUserPage;