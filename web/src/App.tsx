import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Dashboard } from './pages/Dashboard';
import { NotebookWorkspace } from './pages/NotebookWorkspace';

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gray-900 text-gray-100">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/notebook/:id" element={<NotebookWorkspace />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
