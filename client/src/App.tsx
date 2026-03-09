import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import { UserUpload } from './components/UserUpload';
import { SolicitorDashboard } from './components/SolicitorDashboard';
import './App.css';

function App() {
  return (
    <BrowserRouter>
      <nav style={{ padding: '1rem', borderBottom: '1px solid #ccc', marginBottom: '1rem' }}>
        <Link to="/user" style={{ marginRight: '1rem' }}>User Upload</Link>
        <Link to="/solicitor">Solicitor Dashboard</Link>
      </nav>
      <Routes>
        <Route path="/user" element={<UserUpload />} />
        <Route path="/solicitor" element={<SolicitorDashboard />} />
        <Route path="/" element={<UserUpload />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
