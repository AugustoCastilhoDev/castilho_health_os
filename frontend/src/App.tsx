import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider } from './lib/auth/AuthContext'
import { ProtectedRoute } from './lib/auth/ProtectedRoute'
import { AppLayout } from './layouts/AppLayout'
import { LoginPage } from './pages/LoginPage'
import { DashboardPage } from './pages/DashboardPage'
import { AgendaPage } from './pages/AgendaPage'
import { PatientListPage } from './pages/PatientListPage'
import { PatientFormPage } from './pages/PatientFormPage'
import { PatientRecordPage } from './pages/PatientRecordPage'
import { FinancialPage } from './pages/FinancialPage'
import { DocumentTemplatesPage } from './pages/DocumentTemplatesPage'

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<ProtectedRoute />}>
            <Route element={<AppLayout />}>
              <Route index element={<DashboardPage />} />
              <Route path="agenda" element={<AgendaPage />} />
              <Route path="pacientes" element={<PatientListPage />} />
              <Route path="pacientes/novo" element={<PatientFormPage />} />
              <Route path="pacientes/:patientId/editar" element={<PatientFormPage />} />
              <Route path="pacientes/:patientId" element={<PatientRecordPage />} />
              <Route path="financeiro" element={<FinancialPage />} />
              <Route path="documentos" element={<DocumentTemplatesPage />} />
            </Route>
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
