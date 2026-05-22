import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { HomePage } from './pages/HomePage'
import { UploadPage } from './pages/UploadPage'
import { AnalysisPage } from './pages/AnalysisPage'
import { ReviewPage } from './pages/ReviewPage'
import { SharePage } from './pages/SharePage'
import { Layout } from './components/Layout'

function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/upload" element={<UploadPage />} />
          {/* /import merged into /upload */}
          <Route path="/import" element={<Navigate to="/upload" replace />} />
          <Route path="/analysis" element={<AnalysisPage />} />
          <Route path="/analysis/:id" element={<AnalysisPage />} />
          <Route path="/review/:id" element={<ReviewPage />} />
          <Route path="/s/:snapshotId" element={<SharePage />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  )
}

export default App
