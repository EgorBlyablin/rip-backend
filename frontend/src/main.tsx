import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Route, Routes } from 'react-router'
import {registerSW} from "virtual:pwa-register"

import { MainPage } from './pages/main'
import { Layout } from './components/layout'
import { TurbinesListPage } from './pages/turbines-list'
import { TurbineDetailsPage } from './pages/turbine-details'
import 'bootstrap/dist/css/bootstrap.min.css'
import { ROUTES } from './routes'


createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter basename='/RIP/'>
      <Routes>
        <Route path={ROUTES.HOME} element={<MainPage />} />
        <Route path={ROUTES.TURBINES_LIST} element={<Layout />}>
          <Route index element={<TurbinesListPage />} />
          <Route path=":turbineId" element={<TurbineDetailsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)

if ("serviceWorker" in navigator) {
  registerSW()
}