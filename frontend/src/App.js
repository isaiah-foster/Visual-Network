import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import InPage from './InPage';
import OutPage from './OutPage';
import MixedPage from './MixedPage';

function App() {
  return (
    <Router basename="/WSU-CPTS322-SP26/Visual-Network">
      <div>
        <nav style={{ marginBottom: '20px', padding: '15px', backgroundColor: '#969896' }}>
          <Link to="/" className='mr-6 inline-block transition-transform durtion-150 ease-out hover:-translate-y-0.5 transition-all ease-in'>Home</Link>
          <Link to="/in" className='mr-6 inline-block transition-transform durtion-150 ease-out hover:-translate-y-0.5 transition-all ease-in'>In Page</Link>
          <Link to="/out" className='mr-6 inline-block transition-transform durtion-150 ease-out hover:-translate-y-0.5 transition-all ease-in'>Out Page</Link>
          <Link to="/mixed" className='mr-6 inline-block transition-transform durtion-150 ease-out hover:-translate-y-0.5 transition-all ease-in'>Mixed Page</Link>
        </nav>

        <div style={{ backgroundColor: '#ffffff', minHeight: 'calc(100vh - 74px)', padding: '40px' }}>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/in" element={<InPage />} />
            <Route path="/out" element={<OutPage />} />
            <Route path="/mixed" element={<MixedPage />} />
          </Routes>
        </div>
      </div>
    </Router>
  );
}

function Home() {
  return (
    <div style={{ maxWidth: '2000px', lineHeight: '1.6' , fontSize: "1.3rem"}}>
      <h1 style={{ fontSize: '3rem', marginBottom: '12px' }}>Visual Network</h1>
      <p style={{ marginBottom: '14px' }}>
        This project shows a visualization of traffic thats coming into and out of your computer. Some of the things were measuring include: 
      </p>
      <p><strong>Volume In: </strong>  Measures the amount of packets recieved from each IP address which is expressed in packets per second (pps).</p>
      <p><strong>Volume Out: </strong>  Measures the amount of packets going out of the computer which is expressed in packets per second (pps).</p>
      <p><strong>Latency:</strong> Measures the delay between sending a request and getting a response expressed in milliseconds (ms). Examples Include: Google | 36ms or Netflix | 79ms. </p>
      <p><strong>Networking Protocol: </strong> Measures moving data across machines. Examples include UDP, TCP, ICMP, etc.</p>
      <br></br>
      <p style={{ marginBottom: '14px' }}>
        Use the navigation above to switch between focused views:
      </p>
      <ul style={{ marginBottom: '14px', paddingLeft: '20px' }}>
        <li><strong>In Page:</strong> emphasizes inbound relationships and incoming connections.</li>
        <li><strong>Out Page:</strong> emphasizes outbound relationships and outgoing connections.</li>
        <li><strong>Mixed Page:</strong> combines both perspectives for a balanced network view.</li>
      </ul>
      <p>
        The goal of this tool is to support learning, debugging, and analysis by showing network behavior
        in a clear and visual way.
      </p>
    </div>
  );
}

export default App;
