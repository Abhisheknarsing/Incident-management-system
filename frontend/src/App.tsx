import { Routes, Route } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { UploadPage } from '@/pages/UploadPage'
import { ProcessingPage } from '@/pages/ProcessingPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { ErrorBoundary } from '@/components/ui/error-boundary'
import { NotificationContainer } from '@/components/ui/notifications'

function App() {
  return (
    <ErrorBoundary>
      <NotificationContainer />
      <Layout>
        <Routes>
          <Route path="/" element={<UploadPage />} />
          <Route path="/processing" element={<ProcessingPage />} />
          <Route path="/dashboard" element={<DashboardPage />} />
        </Routes>
      </Layout>
    </ErrorBoundary>
  )
}

export default App