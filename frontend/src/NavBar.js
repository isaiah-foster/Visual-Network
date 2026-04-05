import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from './AuthContext';

function NavBar() {
  const { user, setUser } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await fetch(`${process.env.REACT_APP_AUTH_URL}/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    });
    setUser(null);
    navigate('/login');
  };

  return (
    <nav style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '20px', padding: '15px', backgroundColor: '#969896' }}>
      <div>
        <Link to="/" className='mr-6 inline-block transition-transform durtion-150 ease-out hover:-translate-y-0.5 transition-all ease-in'>Home</Link>
        <Link to="/in" className='mr-6 inline-block transition-transform durtion-150 ease-out hover:-translate-y-0.5 transition-all ease-in'>In Page</Link>
        <Link to="/out" className='mr-6 inline-block transition-transform durtion-150 ease-out hover:-translate-y-0.5 transition-all ease-in'>Out Page</Link>
        <Link to="/mixed" className='mr-6 inline-block transition-transform durtion-150 ease-out hover:-translate-y-0.5 transition-all ease-in'>Mixed Page</Link>
        {user?.role === 'admin' && <Link to="/admin/users" className='mr-6 inline-block'>Add User</Link>}
      </div>
      {user && <button onClick={handleLogout}>Logout</button>}
    </nav>
  );
}

export default NavBar;