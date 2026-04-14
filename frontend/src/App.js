import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import LoginPage from './LoginPage'; 
import InPage from './InPage';
import OutPage from './OutPage';
import MixedPage from './MixedPage';
import AdminCreateUserPage from './AdminCreateUserPage';
import { ProtectedRoute } from './ProtectedRoute';
import NavBar from './NavBar';

function App() {
  return (
    <Router basename={process.env.NODE_ENV === 'production' ? (process.env.PUBLIC_URL || '/') : '/'}>
      <div className="app-shell">
        <NavBar />
        <main className="app-main">
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<ProtectedRoute><Home /></ProtectedRoute>} />
            <Route path="/in" element={<ProtectedRoute><InPage /></ProtectedRoute>} />
            <Route path="/out" element={<ProtectedRoute><OutPage /></ProtectedRoute>} />
            <Route path="/mixed" element={<ProtectedRoute><MixedPage /></ProtectedRoute>} />
            <Route path="/admin/users" element={<ProtectedRoute requiredRole="admin"><AdminCreateUserPage /></ProtectedRoute>} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

function Home() {
  return (
    <div className="home-page">
      <section className="home-hero">
        <h1 className="page-title">Visual Network</h1>
        <p className="page-subtitle">
          Monitor inbound and outbound traffic in real time, then break it down by volume, latency, and protocol.
        </p>
      </section>

      <section className="info-card">
        <h3>View Guide</h3>
        <p><strong>In Page (Inbound):</strong> Focuses on traffic coming into your machine and incoming relationships.</p>
        <p><strong>Out Page (Outbound):</strong> Focuses on traffic leaving your machine and outgoing destinations.</p>
        <p><strong>Mixed Page:</strong> Combines inbound and outbound perspectives into one dashboard.</p>
      </section>

      <section className="home-grid">
        <article className="info-card">
          <h3>Volume In</h3>
          <p>Shows traffic entering your machine, grouped by destination labels and hosts.</p>
        </article>
        <article className="info-card">
          <h3>Volume Out</h3>
          <p>Shows traffic leaving your machine so you can quickly identify active outbound endpoints.</p>
        </article>
        <article className="info-card">
          <h3>Latency</h3>
          <p>Measures average response delay in milliseconds for each visible source or destination.</p>
        </article>
        <article className="info-card">
          <h3>Protocols</h3>
          <p>Summarizes usage across TCP, UDP, ICMP, and detected L7 protocol categories.</p>
        </article>
      </section>
    </div>
  );
}

export default App;
