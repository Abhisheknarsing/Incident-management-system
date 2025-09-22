import { Routes, Route } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { UploadPage } from '@/pages/UploadPage'
import { ProcessingPage } from '@/pages/ProcessingPage'
import { DashboardPage } from '@/pages/DashboardPage'

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<UploadPage />} />
        <Route path="/processing" element={<ProcessingPage />} />
        <Route path="/dashboard" element={<DashboardPage />} />
      </Routes>
    </Layout>
  )
}

export default App