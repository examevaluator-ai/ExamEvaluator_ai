import React from 'react';
import VideoUpload from './components/VideoUpload';
import './App.css';

function App() {
  return (
    <div className="App">
      <header className="App-header">
        <h1>ExamEvaluator.ai</h1>
      </header>
      <main className="App-main">
        <VideoUpload />
      </main>
    </div>
  );
}

export default App;
