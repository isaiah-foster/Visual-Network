import React from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
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
    <div className="app-nav-wrap">
      <nav className="app-nav">
        <div className="app-nav-left">
          <NavLink to="/" end className="app-brand">VisualNet</NavLink>
          <div className="app-nav-links">
            <NavLink to="/" end className={({ isActive }) => `app-nav-link${isActive ? ' active' : ''}`}>Home</NavLink>
            <NavLink to="/in" className={({ isActive }) => `app-nav-link${isActive ? ' active' : ''}`}>In Page</NavLink>
            <NavLink to="/out" className={({ isActive }) => `app-nav-link${isActive ? ' active' : ''}`}>Out Page</NavLink>
            <NavLink to="/mixed" className={({ isActive }) => `app-nav-link${isActive ? ' active' : ''}`}>Mixed Page</NavLink>
            {user?.role === 'admin' && (
              <NavLink to="/admin/users" className={({ isActive }) => `app-nav-link${isActive ? ' active' : ''}`}>
                Add User
              </NavLink>
            )}
          </div>
        </div>
        {user && <button className="button button-ghost" onClick={handleLogout}>Logout</button>}
      </nav>
    </div>
  );
}

export default NavBar;